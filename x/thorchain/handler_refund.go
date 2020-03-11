package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// RefundHandler a handle to process tx that had refund memo
// usually this type or tx is because Thorchain fail to process the tx, which result in a refund, signer honour the tx and refund customer accordingly
type RefundHandler struct {
	keeper Keeper
	ch     CommonOutboundTxHandler
}

// NewRefundHandler create a new refund handler
func NewRefundHandler(keeper Keeper) RefundHandler {
	return RefundHandler{
		keeper: keeper,
		ch:     NewCommonOutboundTxHander(keeper),
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

	if !isSignedByActiveObserver(ctx, h.keeper, msg.GetSigners()) {
		return sdk.ErrUnauthorized("msg is not signed by an active node account")
	}
	return nil
}

func (h RefundHandler) handle(ctx sdk.Context, msg MsgRefundTx, version semver.Version) sdk.Result {
	inTxID := msg.InTxID
	tx := msg.Tx
	evetIDs, err := h.keeper.GetEventsIDByTxHash(ctx, msg.Tx.Tx.ID)
	if err != nil {
		return h.ch.handle(ctx, msg.Tx, msg.InTxID, EventRefund)
	}
	if len(evetIDs) > 0 {
		event, err := h.keeper.GetEvent(ctx, evetIDs[0])
		if err != nil {
			ctx.Logger().Error(err.Error())
			return sdk.ErrInternal("fail to get observed tx voter").Result()
		}
		if len(event.OutTxs) == 0 || len(event.Fee.Coins)==0 {
			return h.ch.handle(ctx, msg.Tx, msg.InTxID, EventRefund)
		}
	} else {
		return h.ch.handle(ctx, msg.Tx, msg.InTxID, EventRefund)
	}
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
				txOutItem.OutHash.IsEmpty() &&
				tx.Tx.Coins.Contains(txOutItem.Coin) {
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
	}
	h.keeper.SetLastSignedHeight(ctx, voter.Height)
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
	//return h.ch.handle(ctx, msg.Tx, msg.InTxID, EventRefund)
}
