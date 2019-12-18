package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

// VersionHandler is to handle Version message
type VersionHandler struct {
	keeper Keeper
}

// NewVersionHandler create new instance of VersionHandler
func NewVersionHandler(keeper Keeper) VersionHandler {
	return VersionHandler{
		keeper: keeper,
	}
}

// Run it the main entry point to execute Version logic
func (h VersionHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, _ constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgSetVersion)
	if !ok {
		return errInvalidMessage.Result()
	}
	ctx.Logger().Info("receive version number",
		"version", msg.Version.String())
	if err := h.validate(ctx, msg, version); err != nil {
		ctx.Logger().Error("msg set version failed validation", err)
		return err.Result()
	}
	if err := h.handle(ctx, msg, version); err != nil {
		ctx.Logger().Error("fail to process msg set version", err)
		return err.Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func (h VersionHandler) validate(ctx sdk.Context, msg MsgSetVersion, version semver.Version) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	} else {
		return errBadVersion
	}
}

func (h VersionHandler) validateV1(ctx sdk.Context, msg MsgSetVersion) sdk.Error {
	if err := msg.ValidateBasic(); nil != err {
		return err
	}

	nodeAccount, err := h.keeper.GetNodeAccount(ctx, msg.Signer)
	if err != nil {
		ctx.Logger().Error("fail to get node account", "error", err, "address", msg.Signer.String())
		return sdk.ErrUnauthorized(fmt.Sprintf("%s is not authorizaed", msg.Signer))
	}
	if nodeAccount.IsEmpty() {
		ctx.Logger().Error("unauthorized account", "address", msg.Signer.String())
		return sdk.ErrUnauthorized(fmt.Sprintf("%s is not authorizaed", msg.Signer))
	}

	return nil

}

func (h VersionHandler) handle(ctx sdk.Context, msg MsgSetVersion, version semver.Version) sdk.Error {
	ctx.Logger().Info("handleMsgSetVersion request", "Version:", msg.Version.String())
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return errBadVersion
	}
}

func (h VersionHandler) handleV1(ctx sdk.Context, msg MsgSetVersion) sdk.Error {
	nodeAccount, err := h.keeper.GetNodeAccount(ctx, msg.Signer)
	if err != nil {
		ctx.Logger().Error("fail to get node account", "error", err, "address", msg.Signer.String())
		return sdk.ErrUnauthorized(fmt.Sprintf("unable to find account: %s", msg.Signer))
	}

	if nodeAccount.Version.LT(msg.Version) {
		nodeAccount.Version = msg.Version
	}

	if err := h.keeper.SetNodeAccount(ctx, nodeAccount); nil != err {
		ctx.Logger().Error("fail to save node account", err)
		return sdk.ErrInternal("fail to save node account")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent("set_version",
			sdk.NewAttribute("thor_address", msg.Signer.String()),
			sdk.NewAttribute("version", msg.Version.String())))

	return nil
}
