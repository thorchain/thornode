package thorchain

import (
	"github.com/blang/semver"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

type OutboundTxHandler struct {
	keeper                Keeper
	ch                    CommonOutboundTxHandler
	versionedEventManager VersionedEventManager
}

func NewOutboundTxHandler(keeper Keeper, versionedEventManager VersionedEventManager) OutboundTxHandler {
	return OutboundTxHandler{
		keeper: keeper,
		ch:     NewCommonOutboundTxHandler(keeper, versionedEventManager),
	}
}

func (h OutboundTxHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, _ constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgOutboundTx)
	if !ok {
		return errInvalidMessage.Result()
	}
	if err := h.validate(ctx, msg, version); err != nil {
		return err.Result()
	}
	return h.handle(ctx, msg, version)
}

func (h OutboundTxHandler) validate(ctx sdk.Context, msg MsgOutboundTx, version semver.Version) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	}
	ctx.Logger().Error(errInvalidVersion.Error())
	return errBadVersion
}

func (h OutboundTxHandler) validateV1(ctx sdk.Context, msg MsgOutboundTx) sdk.Error {
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error(err.Error())
		return err
	}

	if !isSignedByActiveNodeAccounts(ctx, h.keeper, msg.GetSigners()) {
		ctx.Logger().Error(notAuthorized.Error())
		return sdk.ErrUnauthorized("Not Authorized")
	}
	return nil
}

func (h OutboundTxHandler) handle(ctx sdk.Context, msg MsgOutboundTx, version semver.Version) sdk.Result {
	ctx.Logger().Info("receive MsgOutboundTx", "request outbound tx hash", msg.Tx.Tx.ID)
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, version, msg)
	}
	ctx.Logger().Error(errInvalidVersion.Error())
	return errBadVersion.Result()
}

func (h OutboundTxHandler) handleV1(ctx sdk.Context, version semver.Version, msg MsgOutboundTx) sdk.Result {
	return h.ch.handle(ctx, version, msg.Tx, msg.InTxID, EventSuccess)
}
