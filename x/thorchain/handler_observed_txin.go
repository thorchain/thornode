package thorchain

import (
	"encoding/json"
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

type ObservedTxInHandler struct {
	keeper       Keeper
	txOutStore   TxOutStore
	validatorMgr ValidatorManager
	vaultMgr     VaultManager
}

func NewObservedTxInHandler(keeper Keeper, txOutStore TxOutStore, validatorMgr ValidatorManager, vaultMgr VaultManager) ObservedTxInHandler {
	return ObservedTxInHandler{
		keeper:       keeper,
		txOutStore:   txOutStore,
		validatorMgr: validatorMgr,
		vaultMgr:     vaultMgr,
	}
}

func (h ObservedTxInHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, _ constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgObservedTxIn)
	if !ok {
		return errInvalidMessage.Result()
	}
	isNewSigner, err := h.validate(ctx, msg, version)
	if err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	if isNewSigner {
		return sdk.Result{
			Code:      sdk.CodeOK,
			Codespace: DefaultCodespace,
		}
	}
	return h.handle(ctx, msg, version)
}

func (h ObservedTxInHandler) validate(ctx sdk.Context, msg MsgObservedTxIn, version semver.Version) (bool, error) {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return false, badVersion
	}
}

func (h ObservedTxInHandler) validateV1(ctx sdk.Context, msg MsgObservedTxIn) (bool, error) {
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return false, err
	}

	if !isSignedByActiveObserver(ctx, h.keeper, msg.GetSigners()) {
		signers := msg.GetSigners()
		for _, signer := range signers {
			newSigner, err := h.signedByNewObserver(ctx, signer)
			if nil != err {
				ctx.Logger().Error("fail to determinate whether the tx is signed by a new observer", err)
				return false, notAuthorized
			}

			// if this tx is signed by a new observer , we have to return a success code
			if newSigner {
				return true, nil
			}
		}
		ctx.Logger().Error(notAuthorized.Error())
		return false, notAuthorized
	}

	return false, nil
}

// when THORChain observe a tx is signed by new observer, who's node account still in standby status, THORChain need to mark their observer is alive.
// by doing that, it also need to return a success code, otherwise the change will not be saved to key value store.
func (h ObservedTxInHandler) signedByNewObserver(ctx sdk.Context, addr sdk.AccAddress) (bool, error) {
	nodeAcct, err := h.keeper.GetNodeAccount(ctx, addr)
	if nil != err {
		return false, fmt.Errorf("fail to get node account(%s): %w", addr.String(), err)
	}
	if nodeAcct.Status != NodeStandby {
		return false, fmt.Errorf("node account (%s) is in status(%s) not standby yet", addr, nodeAcct.Status)
	}
	nodeAcct.ObserverActive = true
	err = h.keeper.SetNodeAccount(ctx, nodeAcct)
	if nil == err {
		return true, nil
	}
	return false, fmt.Errorf("fail to save node account(%s): %w", addr, err)

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

	handler := NewHandler(h.keeper, h.txOutStore, h.validatorMgr, h.vaultMgr)

	for _, tx := range msg.Txs {

		// check we are sending to a valid vault
		if !h.keeper.VaultExists(ctx, tx.ObservedPubKey) {
			fmt.Printf("Observed Pubkey: %s\n", tx.ObservedPubKey)
			return sdk.ErrInternal("Observed Tx Pubkey is not associated with a valid vault").Result()
		}

		voter, err := h.keeper.GetObservedTxVoter(ctx, tx.Tx.ID)
		if err != nil {
			return sdk.ErrInternal(err.Error()).Result()
		}

		voter, ok := h.preflight(ctx, voter, activeNodeAccounts, tx, msg.Signer)
		if !ok {
			continue
		}

		txIn := voter.GetTx(activeNodeAccounts)

		if ok := isCurrentVaultPubKey(ctx, h.keeper, tx); !ok {
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
