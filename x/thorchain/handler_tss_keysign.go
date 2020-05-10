package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// TssKeysignHandler is design to process MsgTssKeysignFail
type TssKeysignHandler struct {
	keeper Keeper
}

// NewTssKeysignHandler create a new instance of TssKeysignHandler
// when a signer fail to join tss keysign , thorchain need to slash their node account
func NewTssKeysignHandler(keeper Keeper) TssKeysignHandler {
	return TssKeysignHandler{
		keeper: keeper,
	}
}

func (h TssKeysignHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgTssKeysignFail)
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

func (h TssKeysignHandler) validate(ctx sdk.Context, msg MsgTssKeysignFail, version semver.Version) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	}
	return errBadVersion
}

func (h TssKeysignHandler) validateV1(ctx sdk.Context, msg MsgTssKeysignFail) sdk.Error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	if !isSignedByActiveNodeAccounts(ctx, h.keeper, msg.GetSigners()) {
		return sdk.ErrUnauthorized("not authorized")
	}

	return nil
}

func (h TssKeysignHandler) handle(ctx sdk.Context, msg MsgTssKeysignFail, version semver.Version) sdk.Result {
	ctx.Logger().Info("handle MsgTssKeysignFail request", "ID:", msg.ID)
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, msg, version)
	}
	return errBadVersion.Result()
}

// Handle a message to observe inbound tx
func (h TssKeysignHandler) handleV1(ctx sdk.Context, msg MsgTssKeysignFail, version semver.Version) sdk.Result {
	active, err := h.keeper.ListActiveNodeAccounts(ctx)
	if err != nil {
		err = wrapError(ctx, err, "fail to get list of active node accounts")
		return sdk.ErrInternal(err.Error()).Result()
	}

	if !msg.Blame.IsEmpty() {
		ctx.Logger().Error(msg.Blame.String())
	}

	voter, err := h.keeper.GetTssKeysignFailVoter(ctx, msg.ID)
	if err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}

	voter.Sign(msg.Signer)
	h.keeper.SetTssKeysignFailVoter(ctx, voter)
	// doesn't have consensus yet
	if !voter.HasConsensus(active) {
		ctx.Logger().Info("not having consensus yet, return")
		return sdk.Result{
			Code:      sdk.CodeOK,
			Codespace: DefaultCodespace,
		}
	}

	if voter.Height == 0 {
		voter.Height = ctx.BlockHeight()
		h.keeper.SetTssKeysignFailVoter(ctx, voter)

		constAccessor := constants.GetConstantValues(version)
		slashPoints := constAccessor.GetInt64Value(constants.FailKeySignSlashPoints)
		// fail to generate a new tss key let's slash the node account
		for _, node := range msg.Blame.BlameNodes {
			nodePubKey, err := common.NewPubKey(node.Pubkey)
			if err != nil {
				ctx.Logger().Error("fail to parse pubkey")
				return sdk.ErrInternal("fail to parse pubkey").Result()
			}
			na, err := h.keeper.GetNodeAccountByPubKey(ctx, nodePubKey)
			if err != nil {
				ctx.Logger().Error("fail to get node from it's pub key", "error", err, "pub key", nodePubKey.String())
				return sdk.ErrInternal("fail to get node account").Result()
			}
			if err := h.keeper.IncNodeAccountSlashPoints(ctx, na.NodeAddress, slashPoints); err != nil {
				ctx.Logger().Error("fail to inc slash points", "error", err)
			}
		}
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
