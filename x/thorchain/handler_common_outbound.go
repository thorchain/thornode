package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// CommonOutboundTxHandler is the place where those common logic can be shared between multiple different kind of outbound tx handler
// at the moment, handler_refund, and handler_outbound_tx are largely the same , only some small difference
type CommonOutboundTxHandler struct {
	keeper                Keeper
	versionedEventManager VersionedEventManager
}

// NewCommonOutboundTxHandler create a new instance of the CommonOutboundTxHandler
func NewCommonOutboundTxHandler(k Keeper, versionedEventManager VersionedEventManager) CommonOutboundTxHandler {
	return CommonOutboundTxHandler{
		keeper:                k,
		versionedEventManager: versionedEventManager,
	}
}

func (h CommonOutboundTxHandler) slash(ctx sdk.Context, version semver.Version, tx ObservedTx) error {
	var returnErr error
	slasher, err := NewSlasher(h.keeper, version, h.versionedEventManager)
	if err != nil {
		return fmt.Errorf("fail to create new slasher,error:%w", err)
	}
	for _, c := range tx.Tx.Coins {
		if err := slasher.SlashNodeAccount(ctx, tx.ObservedPubKey, c.Asset, c.Amount); err != nil {
			ctx.Logger().Error("fail to slash account", "error", err)
			returnErr = err
		}
	}
	return returnErr
}

func (h CommonOutboundTxHandler) handle(ctx sdk.Context, version semver.Version, tx ObservedTx, inTxID common.TxID, status EventStatus) sdk.Result {
	voter, err := h.keeper.GetObservedTxVoter(ctx, inTxID)
	if err != nil {
		ctx.Logger().Error("fail to get observed tx voter", "error", err)
		return sdk.ErrInternal("fail to get observed tx voter").Result()
	}

	if voter.Height > 0 {
		voter.AddOutTx(tx.Tx)
		h.keeper.SetObservedTxVoter(ctx, voter)
	}

	// update txOut record with our TxID that sent funds out of the pool
	txOut, err := h.keeper.GetTxOut(ctx, voter.Height)
	if err != nil {
		ctx.Logger().Error("unable to get txOut record", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	// Save TxOut back with the TxID only when the TxOut on the block height is
	// not empty
	shouldSlash := true
	for i, txOutItem := range txOut.TxArray {
		// withdraw , refund etc, one inbound tx might result two outbound
		// txes, THORNode have to correlate outbound tx back to the
		// inbound, and also txitem , thus THORNode could record both
		// outbound tx hash correctly given every tx item will only have
		// one coin in it , THORNode could use that to identify which tx it
		// is
		if txOutItem.InHash.Equals(inTxID) &&
			txOutItem.OutHash.IsEmpty() &&
			tx.Tx.Coins.Equals(common.Coins{txOutItem.Coin}) &&
			tx.Tx.Chain.Equals(txOutItem.Chain) &&
			tx.Tx.ToAddress.Equals(txOutItem.ToAddress) &&
			tx.ObservedPubKey.Equals(txOutItem.VaultPubKey) {

			txOut.TxArray[i].OutHash = tx.Tx.ID
			shouldSlash = false

			if err := h.keeper.SetTxOut(ctx, txOut); err != nil {
				ctx.Logger().Error("fail to save tx out", "error", err)
			}
			break
		}
	}

	if shouldSlash {
		if err := h.slash(ctx, version, tx); err != nil {
			return sdk.ErrInternal("fail to slash account").Result()
		}
	}

	h.keeper.SetLastSignedHeight(ctx, voter.Height)

	// complete events
	if voter.IsDone() {
		err := completeEvents(ctx, h.keeper, inTxID, voter.OutTxs, status)
		if err != nil {
			ctx.Logger().Error("unable to complete events", "error", err)
			return sdk.ErrInternal(err.Error()).Result()
		}
		eventMgr, err := h.versionedEventManager.GetEventManager(ctx, version)
		if err != nil {
			ctx.Logger().Error("fail to get event manager", "error", err)
			return errFailGetEventManager.Result()
		}
		for _, item := range voter.OutTxs {
			if err := eventMgr.EmitOutboundEvent(ctx, NewEventOutbound(inTxID, item)); err != nil {
				ctx.Logger().Error("fail to emit outbound event", "error", err)
				return sdk.ErrInternal("fail to emit outbound event").Result()
			}
		}
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
