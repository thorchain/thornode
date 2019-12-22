package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

// RefundHandler a handle to process tx that had refund memo
// usually this type or tx is because Thorchain fail to process the tx, which result in a refund, signer honour the tx and refund customer accordingly
type RefundHandler struct {
	keeper Keeper
}

// NewRefundHandler create a new refund handler
func NewRefundHandler(keeper Keeper) RefundHandler {
	return RefundHandler{keeper: keeper}
}
func (h RefundHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgBond)
	if !ok {
		return errInvalidMessage.Result()
	}
	ctx.Logger().Info("receive MsgRefund",
		"node address", msg.NodeAddress,
		"request hash", msg.RequestTxHash,
		"bond", msg.Bond)
	if err := h.validate(ctx, msg, version, constAccessor); nil != err {
		ctx.Logger().Error("msg bond fail validation", err)
		return err.Result()
	}

	//if err := h.handle(ctx, msg, version); nil != err {
	//	ctx.Logger().Error("fail to process msg bond", err)
	//	return err.Result()
	//}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func (h RefundHandler) validate(ctx sdk.Context, msg MsgBond, version semver.Version, constAccessor constants.ConstantValues) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, version, msg, constAccessor)
	}
	return errBadVersion
}

func (h RefundHandler) validateV1(ctx sdk.Context, version semver.Version, msg MsgBond, constAccessor constants.ConstantValues) sdk.Error {
	if err := msg.ValidateBasic(); nil != err {
		return err
	}

	if !isSignedByActiveNodeAccounts(ctx, h.keeper, msg.GetSigners()) {
		return sdk.ErrUnauthorized("msg is not signed by an active node account")
	}
	return nil
}

func (h RefundHandler) handle(ctx sdk.Context, msg MsgBond, version semver.Version) sdk.Error {
	return nil
}
