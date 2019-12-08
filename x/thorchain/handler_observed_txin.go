package thorchain

import (
	"encoding/json"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ObservedTxInHandler struct {
	keeper       Keeper
	txOutStore   TxOutStore
	poolAddrMgr  PoolAddressManager
	validatorMgr ValidatorManager
}

func NewObservedTxInHandler(keeper Keeper, txOutStore TxOutStore, poolAddrMgr PoolAddressManager, validatorMgr ValidatorManager) ObservedTxInHandler {
	return ObservedTxInHandler{
		keeper:       keeper,
		txOutStore:   txOutStore,
		poolAddrMgr:  poolAddrMgr,
		validatorMgr: validatorMgr,
	}
}

func (h ObservedTxInHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version) sdk.Result {
	msg, ok := m.(MsgObservedTxIn)
	if !ok {
		return errInvalidMessage.Result()
	}
	if err := h.validate(ctx, msg, version); err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	return h.handle(ctx, msg, version)
}

func (h ObservedTxInHandler) validate(ctx sdk.Context, msg MsgObservedTxIn, version semver.Version) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

func (h ObservedTxInHandler) validateV1(ctx sdk.Context, msg MsgObservedTxIn) error {
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

func (h ObservedTxInHandler) handle(ctx sdk.Context, msg MsgObservedTxIn, version semver.Version) sdk.Result {
	ctx.Logger().Info("handleMsgObservedTxIn request", "Tx:", msg.Txs[0].String())
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return errBadVersion.Result()
	}
}

func (h ObservedTxInHandler) inboundFailure(ctx sdk.Context, tx ObservedTx) error {
	err := refundTx(ctx, tx, h.txOutStore, h.keeper, true)
	if err != nil {
		return err
	}
	ee := NewEmptyRefundEvent()
	buf, err := json.Marshal(ee)
	if nil != err {
		return err
	}
	event := NewEvent(
		ee.Type(),
		ctx.BlockHeight(),
		tx.Tx,
		buf,
		EventRefund,
	)

	return h.keeper.AddIncompleteEvents(ctx, event)
}

func (h ObservedTxInHandler) preflight(ctx sdk.Context, voter ObservedTxVoter, nas NodeAccounts, tx ObservedTx, signer sdk.AccAddress) (ObservedTxVoter, bool) {
	voter.Add(tx, signer)

	ok := false
	if voter.HasConensus(nas) && voter.Height == 0 {
		ok = true
		voter.Height = ctx.BlockHeight()
	}
	h.keeper.SetObservedTxVoter(ctx, voter)

	// Check to see if we have enough identical observations to process the transaction
	return voter, ok
}

// Handle a message to observe inbound tx
func (h ObservedTxInHandler) handleV1(ctx sdk.Context, msg MsgObservedTxIn) sdk.Result {
	activeNodeAccounts, err := h.keeper.ListActiveNodeAccounts(ctx)
	if nil != err {
		err = wrapError(ctx, err, "fail to get list of active node accounts")
		return sdk.ErrInternal(err.Error()).Result()
	}

	handler := NewHandler(h.keeper, h.poolAddrMgr, h.txOutStore, h.validatorMgr)

	for _, tx := range msg.Txs {

		voter, err := h.keeper.GetObservedTxVoter(ctx, tx.Tx.ID)
		if err != nil {
			return sdk.ErrInternal(err.Error()).Result()
		}

		voter, ok := h.preflight(ctx, voter, activeNodeAccounts, tx, msg.Signer)
		if !ok {
			continue
		}

		txIn := voter.GetTx(activeNodeAccounts)

		if ok := isCurrentVaultPubKey(ctx, h.keeper, h.poolAddrMgr, tx); !ok {
			if err := refundTx(ctx, tx, h.txOutStore, h.keeper, false); err != nil {
				return sdk.ErrInternal(err.Error()).Result()
			}
			continue
		}

		m, err := processOneTxIn(ctx, h.keeper, txIn, msg.Signer)
		if nil != err || tx.Tx.Chain.IsEmpty() {
			ctx.Logger().Error("fail to process inbound tx", "error", err, "txhash", tx.Tx.ID.String())
			if err := h.inboundFailure(ctx, tx); err != nil {
				return sdk.ErrInternal(err.Error()).Result()
			}
			continue
		}

		if err := h.keeper.SetLastChainHeight(ctx, tx.Tx.Chain, tx.BlockHeight); nil != err {
			return sdk.ErrInternal(err.Error()).Result()
		}

		// add this chain to our list of supported chains
		chains, err := h.keeper.GetChains(ctx)
		if err != nil {
			return sdk.ErrInternal(err.Error()).Result()
		}
		chains = append(chains, tx.Tx.Chain)
		h.keeper.SetChains(ctx, chains)

		// add addresses to observing addresses. This is used to detect
		// active/inactive observing node accounts
		if err := h.keeper.AddObservingAddresses(ctx, txIn.Signers); err != nil {
			return sdk.ErrInternal(err.Error()).Result()
		}

		result := handler(ctx, m)
		if !result.IsOK() {
			if err := refundTx(ctx, tx, h.txOutStore, h.keeper, true); err != nil {
				return sdk.ErrInternal(err.Error()).Result()
			}
		}
	}
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
