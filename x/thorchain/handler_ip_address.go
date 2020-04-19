package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

// IPAddressHandler is to handle ip address message
type IPAddressHandler struct {
	keeper Keeper
}

// NewIPAddressHandler create new instance of IPAddressHandler
func NewIPAddressHandler(keeper Keeper) IPAddressHandler {
	return IPAddressHandler{
		keeper: keeper,
	}
}

// Run it the main entry point to execute ip address logic
func (h IPAddressHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, _ constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgSetIPAddress)
	if !ok {
		return errInvalidMessage.Result()
	}
	ctx.Logger().Info("receive ip address", "address", msg.IPAddress)
	if err := h.validate(ctx, msg, version); err != nil {
		ctx.Logger().Error("msg set version failed validation", "error", err)
		return err.Result()
	}
	if err := h.handle(ctx, msg, version); err != nil {
		ctx.Logger().Error("fail to process msg set version", "error", err)
		return err.Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func (h IPAddressHandler) validate(ctx sdk.Context, msg MsgSetIPAddress, version semver.Version) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	} else {
		return errBadVersion
	}
}

func (h IPAddressHandler) validateV1(ctx sdk.Context, msg MsgSetIPAddress) sdk.Error {
	if err := msg.ValidateBasic(); err != nil {
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

func (h IPAddressHandler) handle(ctx sdk.Context, msg MsgSetIPAddress, version semver.Version) sdk.Error {
	ctx.Logger().Info("handleMsgSetIPAddress request", "ip address", msg.IPAddress)
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, msg)
	} else {
		ctx.Logger().Error(errInvalidVersion.Error())
		return errBadVersion
	}
}

func (h IPAddressHandler) handleV1(ctx sdk.Context, msg MsgSetIPAddress) sdk.Error {
	nodeAccount, err := h.keeper.GetNodeAccount(ctx, msg.Signer)
	if err != nil {
		ctx.Logger().Error("fail to get node account", "error", err, "address", msg.Signer.String())
		return sdk.ErrUnauthorized(fmt.Sprintf("unable to find account: %s", msg.Signer))
	}

	nodeAccount.IPAddress = msg.IPAddress

	if err := h.keeper.SetNodeAccount(ctx, nodeAccount); err != nil {
		ctx.Logger().Error("fail to save node account", "error", err)
		return sdk.ErrInternal("fail to save node account")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent("set_ip_address",
			sdk.NewAttribute("thor_address", msg.Signer.String()),
			sdk.NewAttribute("address", msg.IPAddress)))

	return nil
}
