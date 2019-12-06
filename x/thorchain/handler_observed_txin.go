package thorchain

import (
	"encoding/json"
	"fmt"

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

func (h ObservedTxInHandler) Validate(ctx sdk.Context, msg MsgObservedTxIn, version semver.Version) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.ValidateV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

func (h ObservedTxInHandler) ValidateV1(ctx sdk.Context, msg MsgObservedTxIn) error {
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

func (h ObservedTxInHandler) Handle(ctx sdk.Context, msg MsgObservedTxIn, version semver.Version) error {
	ctx.Logger().Info("handleMsgObservedTxIn request", "Tx:", msg.Txs[0].String())
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.HandleV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

func (h ObservedTxInHandler) InboundFailure(ctx sdk.Context, tx ObservedTx) error {
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

// Handle a message to observe inbound tx
func (h ObservedTxInHandler) HandleV1(ctx sdk.Context, msg MsgObservedTxIn) error {
	activeNodeAccounts, err := h.keeper.ListActiveNodeAccounts(ctx)
	if nil != err {
		return wrapError(ctx, err, "fail to get list of active node accounts")
	}

	handler := NewHandler(h.keeper, h.poolAddrMgr, h.txOutStore, h.validatorMgr)

	for _, tx := range msg.Txs {
		voter, err := h.keeper.GetObservedTxVoter(ctx, tx.Tx.ID)
		if err != nil {
			return err
		}
		preConsensus := voter.HasConensus(activeNodeAccounts)
		voter.Add(tx, msg.Signer)
		postConsensus := voter.HasConensus(activeNodeAccounts)
		h.keeper.SetObservedTxVoter(ctx, voter)

		// Check to see if we have enough identical observations to process the transaction
		if preConsensus == false && postConsensus == true && voter.Height == 0 {
			voter.Height = ctx.BlockHeight()
			h.keeper.SetObservedTxVoter(ctx, voter)
			txIn := voter.GetTx(activeNodeAccounts) // get consensus tx, in case our for loop is incorrect

			if ok := isCurrentVaultPubKey(ctx, h.keeper, h.poolAddrMgr, txIn); !ok {
				if err := refundTx(ctx, txIn, h.txOutStore, h.keeper, false); err != nil {
					return err
				}
				continue
			}

			m, err := processOneTxIn(ctx, h.keeper, txIn, msg.Signer)
			if nil != err || tx.Tx.Chain.IsEmpty() {
				ctx.Logger().Error("fail to process txIn", "error", err, "txhash", tx.Tx.ID.String())
				// Detect if the txIn is to the thorchain network or from the
				// thorchain network
				if err := h.InboundFailure(ctx, txIn); err != nil {
					return err
				}
				continue
			}

			if err := h.keeper.SetLastChainHeight(ctx, tx.Tx.Chain, txIn.BlockHeight); nil != err {
				return err
			}

			// add this chain to our list of supported chains
			chains, err := h.keeper.GetChains(ctx)
			if err != nil {
				return err
			}
			chains = append(chains, tx.Tx.Chain)
			h.keeper.SetChains(ctx, chains)

			// add addresses to observing addresses. This is used to detect
			// active/inactive observing node accounts
			if err := h.keeper.AddObservingAddresses(ctx, txIn.Signers); err != nil {
				fmt.Printf("ERR5: %s\n", err)
				return err
			}

			result := handler(ctx, m)
			if !result.IsOK() {
				if err := refundTx(ctx, txIn, h.txOutStore, h.keeper, true); err != nil {
					fmt.Printf("ERR6: %s\n", err)
					return err
				}
			}
		}
	}

	fmt.Println("DONE.")
	return nil
}
