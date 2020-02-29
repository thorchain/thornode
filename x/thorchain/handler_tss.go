package thorchain

import (
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

	if !isSignedByActiveNodeAccounts(ctx, h.keeper, msg.GetSigners()) {
		return sdk.ErrUnauthorized("not authorized")
	}

	return nil
}

func (h TssHandler) handle(ctx sdk.Context, msg MsgTssPool, version semver.Version) sdk.Result {
	ctx.Logger().Info("handleMsgTssPool request", "ID:", msg.ID)
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, msg, version)
	}
	return errBadVersion.Result()
}

// Handle a message to observe inbound tx
func (h TssHandler) handleV1(ctx sdk.Context, msg MsgTssPool, version semver.Version) sdk.Result {
	active, err := h.keeper.ListActiveNodeAccounts(ctx)
	if err != nil {
		err = wrapError(ctx, err, "fail to get list of active node accounts")
		return sdk.ErrInternal(err.Error()).Result()
	}

	if !msg.Blame.IsEmpty() {
		ctx.Logger().Error(msg.Blame.String())
	}

	voter, err := h.keeper.GetTssVoter(ctx, msg.ID)
	if err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}

	if voter.PoolPubKey.IsEmpty() {
		voter.PoolPubKey = msg.PoolPubKey
		voter.PubKeys = msg.PubKeys
	}

	voter.Sign(msg.Signer)
	h.keeper.SetTssVoter(ctx, voter)
	// doesn't have consensus yet
	if !voter.HasConsensus(active) {
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
			vault := NewVault(ctx.BlockHeight(), ActiveVault, vaultType, voter.PoolPubKey)
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
			constAccessor := constants.GetConstantValues(version)
			slashPoints := constAccessor.GetInt64Value(constants.FailKeygenSlashPoints)
			// fail to generate a new tss key let's slash the node account
			for _, pubkeyStr := range msg.Blame.BlameNodes {
				nodePubKey, err := common.NewPubKey(pubkeyStr)
				if err != nil {
					ctx.Logger().Error("fail to parse pubkey", "error", err, "pub key", pubkeyStr)
					return sdk.ErrInternal("fail to parse pubkey").Result()
				}

				na, err := h.keeper.GetNodeAccountByPubKey(ctx, nodePubKey)
				if err != nil {
					ctx.Logger().Error("fail to get node from it's pub key", "error", err, "pub key", nodePubKey.String())
					return sdk.ErrInternal("fail to get node account").Result()
				}
				// 720 blocks per hour
				na.SlashPoints += slashPoints
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
