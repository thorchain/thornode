package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type PoolHandler struct {
	keeper Keeper
}

func NewPoolHandler(keeper Keeper) PoolHandler {
	return PoolHandler{
		keeper: keeper,
	}
}

func (h PoolHandler) Run(ctx sdk.Context, msg MsgPool, version semver.Version) sdk.Result {
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

func (h PoolHandler) Validate(ctx sdk.Context, msg MsgPool, version semver.Version) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.ValidateV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

func (h PoolHandler) ValidateV1(ctx sdk.Context, msg MsgPool) error {
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return err
	}

	if !isSignedByActiveNodeAccounts(ctx, h.keeper, msg.GetSigners()) {
		ctx.Logger().Error(notAuthorized.Error(), "asset", msg.Asset.String())
		return notAuthorized
	}

	return nil

}

func (h PoolHandler) Handle(ctx sdk.Context, msg MsgPool, version semver.Version) error {
	ctx.Logger().Info("handleMsgPool request", "Asset:", msg.Asset.String())
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.HandleV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

// Handle a message to set pooldata
func (h PoolHandler) HandleV1(ctx sdk.Context, msg MsgPool) error {
	pool, err := h.keeper.GetPool(ctx, msg.Asset)
	if err != nil {
		return err
	}

	pool.Status = msg.Status
	pool.Asset = msg.Asset
	return h.keeper.SetPool(ctx, pool)
}
