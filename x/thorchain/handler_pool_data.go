package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type PoolDataHandler struct {
	keeper Keeper
}

func NewPoolDataHandler(keeper Keeper) PoolDataHandler {
	return PoolDataHandler{
		keeper: keeper,
	}
}

func (h PoolDataHandler) Run(ctx sdk.Context, msg MsgSetPoolData, version int64) sdk.Result {
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

func (h PoolDataHandler) Validate(ctx sdk.Context, msg MsgSetPoolData, version int64) error {
	switch version {
	case 0:
		return h.ValidateV0(ctx, msg)
	default:
		panic(fmt.Sprintf("Unable to validate version %d", version))
	}
}

func (h PoolDataHandler) ValidateV0(ctx sdk.Context, msg MsgSetPoolData) error {
	if !isSignedByActiveNodeAccounts(ctx, h.keeper, msg.GetSigners()) {
		ctx.Logger().Error(notAuthorized.Error(), "asset", msg.Asset.String())
		return notAuthorized
	}

	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return err
	}
	return nil

}

func (h PoolDataHandler) Log(ctx sdk.Context, msg MsgSetPoolData) {
	ctx.Logger().Info("handleMsgSetPoolData request", "Asset:", msg.Asset.String())
}

func (h PoolDataHandler) Handle(ctx sdk.Context, msg MsgSetPoolData, version int64) error {
	switch version {
	case 0:
		return h.HandleV0(ctx, msg)
	default:
		panic(fmt.Sprintf("Unable to validate version %d", version))
	}
}

// Handle a message to set pooldata
func (h PoolDataHandler) HandleV0(ctx sdk.Context, msg MsgSetPoolData) error {
	h.keeper.SetPoolData(
		ctx,
		msg.Asset,
		msg.Status,
	)
	return nil
}
