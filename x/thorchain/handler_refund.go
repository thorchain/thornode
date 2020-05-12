package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

// RefundHandler a handle to process tx that had refund memo
// usually this type or tx is because Thorchain fail to process the tx, which result in a refund, signer honour the tx and refund customer accordingly
type RefundHandler struct {
	keeper                Keeper
	ch                    CommonOutboundTxHandler
	versionedEventManager VersionedEventManager
}

// NewRefundHandler create a new refund handler
func NewRefundHandler(keeper Keeper, versionedEventManager VersionedEventManager) RefundHandler {
	return RefundHandler{
		keeper:                keeper,
		ch:                    NewCommonOutboundTxHandler(keeper, versionedEventManager),
		versionedEventManager: versionedEventManager,
	}
}

func (h RefundHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgRefundTx)
	if !ok {
		return errInvalidMessage.Result()
	}
	ctx.Logger().Info("receive MsgRefund",
		"tx ID", msg.InTxID.String())
	if err := h.validate(ctx, msg, version, constAccessor); err != nil {
		ctx.Logger().Error("msg refund fail validation", "error", err)
		return err.Result()
	}

	return h.handle(ctx, msg, version)
}

func (h RefundHandler) validate(ctx sdk.Context, msg MsgRefundTx, version semver.Version, constAccessor constants.ConstantValues) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, version, msg, constAccessor)
	}
	return errBadVersion
}

func (h RefundHandler) validateV1(ctx sdk.Context, version semver.Version, msg MsgRefundTx, constAccessor constants.ConstantValues) sdk.Error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	if !isSignedByActiveNodeAccounts(ctx, h.keeper, msg.GetSigners()) {
		return sdk.ErrUnauthorized("msg is not signed by an active node account")
	}
	return nil
}

func (h RefundHandler) handle(ctx sdk.Context, msg MsgRefundTx, version semver.Version) sdk.Result {
	return h.ch.handle(ctx, version, msg.Tx, msg.InTxID, RefundStatus)
}
