package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ObservedTxInHandler struct {
	keeper Keeper
}

func NewObservedTxInHandler(keeper Keeper) ObservedTxInHandler {
	return ObservedTxInHandler{
		keeper: keeper,
	}
}

func (h ObservedTxInHandler) Run(ctx sdk.Context, msg MsgObservedTxIn, version semver.Version) sdk.Result {
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

func (h ObservedTxInHandler) Validate(ctx sdk.Context, msg MsgObservedTxIn, version semver.Version) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.ValidateV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

func (h ObservedTxInHandler) ValidateV1(ctx sdk.Context, msg MsgObservedTxIn) error {
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return err
	}

	if !isSignedByActiveObserver(ctx, h.keeper, msg.GetSigners()) {
		ctx.Logger().Error(notAuthorized.Error())
		return notAuthorized
	}

	return nil

}

func (h ObservedTxInHandler) Handle(ctx sdk.Context, msg MsgObservedTxIn, version semver.Version) error {
	ctx.Logger().Info("handleMsgObservedTxIn request", "Tx:", msg.Txs[0].String())
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.HandleV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

// Handle a message to observe inbound tx
func (h ObservedTxInHandler) HandleV1(ctx sdk.Context, msg MsgObservedTxIn) error {
	return nil
}
