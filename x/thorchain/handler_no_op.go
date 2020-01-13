package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

type NoOpHandler struct {
	keeper Keeper
}

func NewNoOpHandler(keeper Keeper) NoOpHandler {
	return NoOpHandler{
		keeper: keeper,
	}
}

func (h NoOpHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, _ constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgNoOp)
	if !ok {
		return errInvalidMessage.Result()
	}
	if err := h.Validate(ctx, msg, version); err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	if err := h.Handle(ctx, msg, version); err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func (h NoOpHandler) Validate(ctx sdk.Context, msg MsgNoOp, version semver.Version) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.ValidateV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

func (h NoOpHandler) ValidateV1(ctx sdk.Context, msg MsgNoOp) error {
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return err
	}
	return nil
}

func (h NoOpHandler) Handle(ctx sdk.Context, msg MsgNoOp, version semver.Version) error {
	ctx.Logger().Info("handleMsgNoOp request")
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.HandleV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

// Handle doesn't do anything, its a no op
func (h NoOpHandler) HandleV1(ctx sdk.Context, msg MsgNoOp) error {
	ctx.Logger().Info("receive no op msg")
	return nil
}
