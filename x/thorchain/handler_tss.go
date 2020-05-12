package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type TssHandler struct {
	keeper                Keeper
	versionedVaultManager VersionedVaultManager
}

// NewTssHandler create a new handler to process MsgTssPool
func NewTssHandler(keeper Keeper, versionedVaultManager VersionedVaultManager) TssHandler {
	return TssHandler{
		keeper:                keeper,
		versionedVaultManager: versionedVaultManager,
	}
}

func (h TssHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgTssPool)
	if !ok {
		return errInvalidMessage.Result()
	}
	err := h.validate(ctx, msg, version)
	if err != nil {
		ctx.Logger().Error("msg_tss_pool failed validation", "error", err)
		return err.Result()
	}
	return h.handle(ctx, msg, version)
}

func (h TssHandler) validate(ctx sdk.Context, msg MsgTssPool, version semver.Version) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	}
	return errBadVersion
}

func (h TssHandler) validateV1(ctx sdk.Context, msg MsgTssPool) sdk.Error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	keygenBlock, err := h.keeper.GetKeygenBlock(ctx, msg.Height)
	if err != nil {
		return sdk.ErrUnauthorized(fmt.Errorf("fail to get keygen block from data store: %w", err).Error())
	}

	for _, keygen := range keygenBlock.Keygens {
		for _, member := range keygen.Members {
			addr, err := member.GetThorAddress()
			if addr.Equals(msg.Signer) && err == nil {
				return nil
			}
		}
	}

	return sdk.ErrUnauthorized("not authorized")
}

func (h TssHandler) handle(ctx sdk.Context, msg MsgTssPool, version semver.Version) sdk.Result {
	ctx.Logger().Info("handleMsgTssPool request", "ID:", msg.ID)
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, msg, version)
	}
	return errBadVersion.Result()
}

// Handle a message to observe inbound tx (v0.1.0)
func (h TssHandler) handleV1(ctx sdk.Context, msg MsgTssPool, version semver.Version) sdk.Result {
	if !msg.Blame.IsEmpty() {
		ctx.Logger().Error(msg.Blame.String())
	}

	voter, err := h.keeper.GetTssVoter(ctx, msg.ID)
	if err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}

	// when PoolPubKey is empty , which means TssVoter with id(msg.ID) doesn't
	// exist before, this is the first time to create it
	// set the PoolPubKey to the one in msg, there is no reason voter.PubKeys
	// have anything in it either, thus override it with msg.PubKeys as well
	if voter.PoolPubKey.IsEmpty() {
		voter.PoolPubKey = msg.PoolPubKey
		voter.PubKeys = msg.PubKeys
	}

	voter.Sign(msg.Signer, msg.Chains)
	h.keeper.SetTssVoter(ctx, voter)
	// doesn't have consensus yet
	if !voter.HasConsensus() {
		ctx.Logger().Info("not having consensus yet, return")
		return sdk.Result{
			Code:      sdk.CodeOK,
			Codespace: DefaultCodespace,
		}
	}

	if voter.BlockHeight == 0 {
		voter.BlockHeight = ctx.BlockHeight()
		h.keeper.SetTssVoter(ctx, voter)
		if msg.IsSuccess() {
			vaultType := YggdrasilVault
			if msg.KeygenType == AsgardKeygen {
				vaultType = AsgardVault
			}
			vault := NewVault(ctx.BlockHeight(), ActiveVault, vaultType, voter.PoolPubKey, voter.ConsensusChains())
			vault.Membership = voter.PubKeys
			if err := h.keeper.SetVault(ctx, vault); err != nil {
				ctx.Logger().Error("fail to save vault", "error", err)
				return sdk.ErrInternal("fail to save vault").Result()
			}
			vaultMgr, err := h.versionedVaultManager.GetVaultManager(ctx, h.keeper, version)
			if err != nil {
				ctx.Logger().Error("fail to get a valid vault manager", "error", err)
				return sdk.ErrInternal(err.Error()).Result()
			}
			if err := vaultMgr.RotateVault(ctx, vault); err != nil {
				return sdk.ErrInternal(err.Error()).Result()
			}
		} else {
			// if a node fail to join the keygen, thus hold off the network from churning then it will be slashed accordingly
			constAccessor := constants.GetConstantValues(version)
			slashPoints := constAccessor.GetInt64Value(constants.FailKeygenSlashPoints)
			for _, node := range msg.Blame.BlameNodes {
				nodePubKey, err := common.NewPubKey(node.Pubkey)
				if err != nil {
					ctx.Logger().Error("fail to parse pubkey", "error", err, "pub key", node.Pubkey)
					return sdk.ErrInternal("fail to parse pubkey").Result()
				}

				na, err := h.keeper.GetNodeAccountByPubKey(ctx, nodePubKey)
				if err != nil {
					ctx.Logger().Error("fail to get node from it's pub key", "error", err, "pub key", nodePubKey.String())
					return sdk.ErrInternal("fail to get node account").Result()
				}
				if na.Status == NodeActive {
					// 720 blocks per hour
					if err := h.keeper.IncNodeAccountSlashPoints(ctx, na.NodeAddress, slashPoints); err != nil {
						ctx.Logger().Error("fail to inc slash points", "error", err)
					}
				} else {
					// take out bond from the node account and add it to vault bond reward RUNE
					// thus good behaviour node will get reward
					reserveVault, err := h.keeper.GetVaultData(ctx)
					if err != nil {
						ctx.Logger().Error("fail to get reserve vault", "error", err)
						return sdk.ErrInternal("fail to get reserve vault").Result()
					}

					slashBond := reserveVault.CalcNodeRewards(sdk.NewUint(uint64(slashPoints)))
					na.Bond = common.SafeSub(na.Bond, slashBond)
					if common.RuneAsset().Chain.Equals(common.THORChain) {
						coin := common.NewCoin(common.RuneNative, slashBond)
						if err := h.keeper.SendFromModuleToModule(ctx, BondName, ReserveName, coin); err != nil {
							ctx.Logger().Error("fail to transfer funds from bond to reserve", "error", err)
							return err.Result()

						}
					} else {
						reserveVault.TotalReserve = reserveVault.TotalReserve.Add(slashBond)
						if err := h.keeper.SetVaultData(ctx, reserveVault); err != nil {
							ctx.Logger().Error("fail to set vault data", "error", err)
						}
					}

				}
				if err := h.keeper.SetNodeAccount(ctx, na); err != nil {
					ctx.Logger().Error("fail to save node account", "error", err)
					return sdk.ErrInternal("fail to save node account").Result()
				}
			}

		}
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
