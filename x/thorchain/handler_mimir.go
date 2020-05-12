package thorchain

import (
	"fmt"
	"strconv"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

var ADMIN = sdk.AccAddress("thor1x0akdepu6vs40cv30xqz3qnd85mh7gkf5a0z89")

// MimirHandler is to handle admin messages
type MimirHandler struct {
	keeper Keeper
}

// NewMimirHandler create new instance of MimirHandler
func NewMimirHandler(keeper Keeper) MimirHandler {
	return MimirHandler{
		keeper: keeper,
	}
}

// Run it the main entry point to execute ip address logic
func (h MimirHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, _ constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgMimir)
	if !ok {
		return errInvalidMessage.Result()
	}
	ctx.Logger().Info("receive mimir", "key", msg.Key, "value", msg.Value)
	if err := h.validate(ctx, msg, version); err != nil {
		ctx.Logger().Error("msg mimir failed validation", "error", err)
		return err.Result()
	}
	if err := h.handle(ctx, msg, version); err != nil {
		ctx.Logger().Error("fail to process msg set mimir", "error", err)
		return err.Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func (h MimirHandler) validate(ctx sdk.Context, msg MsgMimir, version semver.Version) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	} else {
		return errBadVersion
	}
}

func (h MimirHandler) validateV1(ctx sdk.Context, msg MsgMimir) sdk.Error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	if !msg.Signer.Equals(ADMIN) {
		ctx.Logger().Error("unauthorized account", "address", msg.Signer.String())
		return sdk.ErrUnauthorized(fmt.Sprintf("%s is not authorizaed", msg.Signer))
	}

	return nil
}

func (h MimirHandler) handle(ctx sdk.Context, msg MsgMimir, version semver.Version) sdk.Error {
	ctx.Logger().Info("handleMsgMimir request", "key", msg.Key, "value", msg.Value)
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, msg)
	} else {
		ctx.Logger().Error(errInvalidVersion.Error())
		return errBadVersion
	}
}

func (h MimirHandler) handleV1(ctx sdk.Context, msg MsgMimir) sdk.Error {
	h.keeper.SetMimir(ctx, msg.Key, msg.Value)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent("set_mimir",
			sdk.NewAttribute("key", msg.Key),
			sdk.NewAttribute("value", strconv.FormatInt(msg.Value, 10))))

	return nil
}
