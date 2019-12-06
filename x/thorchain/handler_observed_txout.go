package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ObservedTxOutHandler struct {
	keeper       Keeper
	txOutStore   TxOutStore
	poolAddrMgr  PoolAddressManager
	validatorMgr *ValidatorManager
}

func NewObservedTxOutHandler(keeper Keeper, txOutStore TxOutStore, poolAddrMgr PoolAddressManager, validatorMgr *ValidatorManager) ObservedTxOutHandler {
	return ObservedTxOutHandler{
		keeper:       keeper,
		txOutStore:   txOutStore,
		poolAddrMgr:  poolAddrMgr,
		validatorMgr: validatorMgr,
	}
}

func (h ObservedTxOutHandler) Run(ctx sdk.Context, msg MsgObservedTxOut, version semver.Version) sdk.Result {
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

func (h ObservedTxOutHandler) Validate(ctx sdk.Context, msg MsgObservedTxOut, version semver.Version) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.ValidateV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

func (h ObservedTxOutHandler) ValidateV1(ctx sdk.Context, msg MsgObservedTxOut) error {
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

func (h ObservedTxOutHandler) Handle(ctx sdk.Context, msg MsgObservedTxOut, version semver.Version) error {
	ctx.Logger().Info("handleMsgObservedTxOut request", "Tx:", msg.Txs[0].String())
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.HandleV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

// Handle a message to observe inbound tx
func (h ObservedTxOutHandler) HandleV1(ctx sdk.Context, msg MsgObservedTxOut) error {
	return nil
}
