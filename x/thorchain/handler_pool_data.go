package thorchain

import (
	"github.com/blang/semver"
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

func (h PoolDataHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version) sdk.Result {
	msg, ok := m.(MsgSetPoolData)
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

func (h PoolDataHandler) Validate(ctx sdk.Context, msg MsgSetPoolData, version semver.Version) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.ValidateV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

func (h PoolDataHandler) ValidateV1(ctx sdk.Context, msg MsgSetPoolData) error {
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

func (h PoolDataHandler) Handle(ctx sdk.Context, msg MsgSetPoolData, version semver.Version) error {

	ctx.Logger().Info("handleMsgSetPoolData request", "Asset:", msg.Asset.String())
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.HandleV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

// Handle a message to set pooldata
func (h PoolDataHandler) HandleV1(ctx sdk.Context, msg MsgSetPoolData) error {
	pool, err := h.keeper.GetPool(ctx, msg.Asset)
	if err != nil {
		return err
	}

	pool.Status = msg.Status
	pool.Asset = msg.Asset
	return h.keeper.SetPool(ctx, pool)
}
