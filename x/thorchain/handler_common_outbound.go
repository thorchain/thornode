package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// CommonOutboundTxHandler is the place where those common logic can be shared between multiple different kind of outbound tx handler
// at the moment, handler_refund, and handler_outbound_tx are largely the same , only some small difference
type CommonOutboundTxHandler struct {
	keeper Keeper
}

// NewCommonOutboundTxHander create a new instance of the CommonOutboundTxHandler
func NewCommonOutboundTxHander(k Keeper) CommonOutboundTxHandler {
	return CommonOutboundTxHandler{keeper: k}
}

func (h CommonOutboundTxHandler) handle(ctx sdk.Context, tx ObservedTx, inTxID common.TxID, status EventStatus) sdk.Result {
	voter, err := h.keeper.GetObservedTxVoter(ctx, inTxID)
	if err != nil {
		ctx.Logger().Error(err.Error())
		return sdk.ErrInternal("fail to get observed tx voter").Result()
	}

	if voter.IsDone() || !voter.AddOutTx(tx.Tx) {
		// voter.IsDone will cover the scenario that more tx out than the actions
		// this will also cover the scenario that the hash in `outbound:xxxx` doesn't match any of the ObservedTxVoter as well
		// the outbound tx doesn't match against any of the action items
		// slash the node account for every coin they send out using this memo
		for _, c := range tx.Tx.Coins {
			if err := slashNodeAccount(ctx, h.keeper, tx.ObservedPubKey, c.Asset, c.Amount); err != nil {
				ctx.Logger().Error("fail to slash account for sending extra fund", "error", err)
				return sdk.ErrInternal("fail to slash account").Result()
			}
		}

		return sdk.Result{
			Code:      sdk.CodeOK,
			Codespace: DefaultCodespace,
		}
	}
	h.keeper.SetObservedTxVoter(ctx, voter)

	// complete events
	if voter.IsDone() {
		err := completeEvents(ctx, h.keeper, inTxID, voter.OutTxs, status)
		if err != nil {
			ctx.Logger().Error("unable to complete events", "error", err)
			return sdk.ErrInternal(err.Error()).Result()
		}
	}

	// update txOut record with our TxID that sent funds out of the pool
	txOut, err := h.keeper.GetTxOut(ctx, voter.Height)
	if err != nil {
		ctx.Logger().Error("unable to get txOut record", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	// Save TxOut back with the TxID only when the TxOut on the block height is
	// not empty
	if !txOut.IsEmpty() {
		processedCoins := common.Coins{}
		for i, txOutItem := range txOut.TxArray {
			if !tx.Tx.Coins.Contains(txOutItem.Coin) {
				continue
			}
			// withdraw , refund etc, one inbound tx might result two outbound txes, THORNode have to correlate outbound tx back to the
			// inbound, and also txitem , thus THORNode could record both outbound tx hash correctly
			// given every tx item will only have one coin in it , THORNode could use that to identify which txit
			if txOutItem.InHash.Equals(inTxID) &&
				txOutItem.OutHash.IsEmpty() {
				txOut.TxArray[i].OutHash = tx.Tx.ID
			}
			processedCoins = append(processedCoins, txOutItem.Coin)
		}
		// the following logic will handle the scenario that pool send out coins that not specific in the original tx out item
		// for example, the txout item says , send 1 RUNE to customer, however , it send 1 RUNE and 1 BNB as a result
		// in that case, thorchain will slash the node account for 1.5 BNB in RUNE value
		for _, c := range tx.Tx.Coins {
			if processedCoins.Contains(c) {
				continue
			}
			if err := slashNodeAccount(ctx, h.keeper, tx.ObservedPubKey, c.Asset, c.Amount); err != nil {
				ctx.Logger().Error("fail to slash account for sending out extra fund", "error", err)
				return sdk.ErrInternal("fail to slash account").Result()
			}
		}
		if err := h.keeper.SetTxOut(ctx, txOut); err != nil {
			ctx.Logger().Error("fail to save tx out", "error", err)
			return sdk.ErrInternal("fail to save tx out").Result()
		}
	}
	h.keeper.SetLastSignedHeight(ctx, voter.Height)

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
