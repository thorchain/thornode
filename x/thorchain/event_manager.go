package thorchain

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// EventManager define methods need to be support to manage events
type EventManager interface {
	CompleteEvents(ctx sdk.Context, keeper Keeper, height int64, txID common.TxID, txs common.Txs, eventStatus EventStatus)
	EmitPoolEvent(ctx sdk.Context, keeper Keeper, txIn common.TxID, status EventStatus, poolEvt EventPool) error
	EmitErrataEvent(ctx sdk.Context, keeper Keeper, txIn common.TxID, errataEvent EventErrata) error
	EmitGasEvent(ctx sdk.Context, keeper Keeper, gasEvent *EventGas) error
	EmitStakeEvent(ctx sdk.Context, keeper Keeper, inTx common.Tx, stakeEvent EventStake) error
	EmitRewardEvent(ctx sdk.Context, keeper Keeper, rewardEvt EventRewards) error
	EmitReserveEvent(ctx sdk.Context, keeper Keeper, reserveEvent EventReserve) error
	EmitUnstakeEvent(ctx sdk.Context, keeper Keeper, unstakeEvt EventUnstake) error
	EmitSwapEvent(ctx sdk.Context, keeper Keeper, swap EventSwap) error
	EmitRefundEvent(ctx sdk.Context, keeper Keeper, refundEvt EventRefund, status EventStatus) error
	EmitBondEvent(ctx sdk.Context, keeper Keeper, bondEvent EventBond) error
	EmitAddEvent(ctx sdk.Context, keeper Keeper, addEvt EventAdd) error
	EmitFeeEvent(ctx sdk.Context, keeper Keeper, feeEvent EventFee) error
	EmitSlashEvent(ctx sdk.Context, keeper Keeper, slashEvt EventSlash) error
	EmitOutboundEvent(ctx sdk.Context, outbound EventOutbound) error
}

// EventMgr implement EventManager interface
type EventMgr struct {
}

// NewEventMgr create a new instance of EventMgr
func NewEventMgr() *EventMgr {
	return &EventMgr{}
}

// CompleteEvents Mark an event in the given block height to the given status
func (m *EventMgr) CompleteEvents(ctx sdk.Context, keeper Keeper, height int64, txID common.TxID, txs common.Txs, eventStatus EventStatus) {
}

// EmitPoolEvent is going to save a pool event to storage
func (m *EventMgr) EmitPoolEvent(ctx sdk.Context, keeper Keeper, txIn common.TxID, status EventStatus, poolEvt EventPool) error {
	bytes, err := json.Marshal(poolEvt)
	if err != nil {
		return fmt.Errorf("fail to marshal pool event: %w", err)
	}

	tx := common.Tx{
		ID: txIn,
	}
	evt := NewEvent(poolEvt.Type(), ctx.BlockHeight(), tx, bytes, status)
	if err := keeper.UpsertEvent(ctx, evt); err != nil {
		return fmt.Errorf("fail to save pool status change event: %w", err)
	}
	events, err := poolEvt.Events()
	if err != nil {
		return fmt.Errorf("fail to get pool events: %w", err)
	}
	ctx.EventManager().EmitEvents(events)

	return nil
}

// EmitErrataEvent generate an errata event
func (m *EventMgr) EmitErrataEvent(ctx sdk.Context, keeper Keeper, txIn common.TxID, errataEvent EventErrata) error {
	errataBuf, err := json.Marshal(errataEvent)
	if err != nil {
		ctx.Logger().Error("fail to marshal errata event to buf", "error", err)
		return fmt.Errorf("fail to marshal errata event to json: %w", err)
	}
	evt := NewEvent(
		errataEvent.Type(),
		ctx.BlockHeight(),
		common.Tx{ID: txIn},
		errataBuf,
		EventSuccess,
	)
	if err := keeper.UpsertEvent(ctx, evt); err != nil {
		ctx.Logger().Error("fail to save errata event", "error", err)
		return fmt.Errorf("fail to save errata event: %w", err)
	}
	events, err := errataEvent.Events()
	if err != nil {
		return fmt.Errorf("fail to emit standard event: %w", err)
	}
	ctx.EventManager().EmitEvents(events)
	return nil
}

func (m *EventMgr) EmitGasEvent(ctx sdk.Context, keeper Keeper, gasEvent *EventGas) error {
	if gasEvent == nil {
		return nil
	}
	buf, err := json.Marshal(gasEvent)
	if err != nil {
		ctx.Logger().Error("fail to marshal gas event", "error", err)
		return fmt.Errorf("fail to marshal gas event to json: %w", err)
	}
	evt := NewEvent(gasEvent.Type(), ctx.BlockHeight(), common.Tx{ID: common.BlankTxID}, buf, EventSuccess)
	if err := keeper.UpsertEvent(ctx, evt); err != nil {
		ctx.Logger().Error("fail to upsert event", "error", err)
		return fmt.Errorf("fail to save gas event: %w", err)
	}
	events, err := gasEvent.Events()
	if err != nil {
		return fmt.Errorf("fail to get events: %w", err)
	}
	ctx.EventManager().EmitEvents(events)

	return nil
}

// EmitStakeEvent add the stake event to block
func (m *EventMgr) EmitStakeEvent(ctx sdk.Context, keeper Keeper, inTx common.Tx, stakeEvent EventStake) error {
	stakeBytes, err := json.Marshal(stakeEvent)
	if err != nil {
		return fmt.Errorf("fail to marshal stake event to json: %w", err)
	}
	evt := NewEvent(
		stakeEvent.Type(),
		ctx.BlockHeight(),
		inTx,
		stakeBytes,
		EventSuccess,
	)
	// stake event doesn't need to have outbound
	tx := common.Tx{ID: common.BlankTxID}
	evt.OutTxs = common.Txs{tx}
	if err := keeper.UpsertEvent(ctx, evt); err != nil {
		return fmt.Errorf("fail to save stake event: %w", err)
	}
	events, err := stakeEvent.Events()
	if err != nil {
		return fmt.Errorf("fail to get events: %w", err)
	}
	ctx.EventManager().EmitEvents(events)
	return nil
}

// EmitRewardEvent save the reward event to keyvalue store and also use event manager
func (m *EventMgr) EmitRewardEvent(ctx sdk.Context, keeper Keeper, rewardEvt EventRewards) error {
	evtBytes, err := json.Marshal(rewardEvt)
	if err != nil {
		return fmt.Errorf("fail to marshal reward event to json: %w", err)
	}
	evt := NewEvent(
		rewardEvt.Type(),
		ctx.BlockHeight(),
		common.Tx{ID: common.BlankTxID},
		evtBytes,
		EventSuccess,
	)
	if err := keeper.UpsertEvent(ctx, evt); err != nil {
		return fmt.Errorf("fail to save event: %w", err)
	}
	events, err := rewardEvt.Events()
	if err != nil {
		return fmt.Errorf("fail to get events: %w", err)
	}
	ctx.EventManager().EmitEvents(events)
	return nil
}

func (m *EventMgr) EmitSwapEvent(ctx sdk.Context, keeper Keeper, swap EventSwap) error {
	buf, err := json.Marshal(swap)
	if err != nil {
		return fmt.Errorf("fail to marshal swap event to json: %w", err)
	}
	evt := NewEvent(swap.Type(), ctx.BlockHeight(), swap.InTx, buf, EventPending)
	// OutTxs is a temporary field that we used, as for now we need to keep backward compatibility so the
	// events change doesn't break midgard and smoke test, for double swap , we first swap the source asset to RUNE ,
	// and then from RUNE to target asset, so the first will be marked as success
	if !swap.OutTxs.IsEmpty() {
		evt.Status = EventSuccess
		evt.OutTxs = common.Txs{swap.OutTxs}
	}
	if err := keeper.UpsertEvent(ctx, evt); err != nil {
		return fmt.Errorf("fail to save swap event: %w", err)
	}
	events, err := swap.Events()
	if err != nil {
		return fmt.Errorf("fail to get events: %w", err)
	}
	ctx.EventManager().EmitEvents(events)
	return nil
}

// EmitReserveEvent emit reserve event both save it to local key value store , and also event manager
func (m *EventMgr) EmitReserveEvent(ctx sdk.Context, keeper Keeper, reserveEvent EventReserve) error {
	buf, err := json.Marshal(reserveEvent)
	if nil != err {
		return err
	}
	e := NewEvent(reserveEvent.Type(), ctx.BlockHeight(), reserveEvent.InTx, buf, EventSuccess)
	if err := keeper.UpsertEvent(ctx, e); err != nil {
		return fmt.Errorf("fail to save reserve event: %w", err)
	}
	events, err := reserveEvent.Events()
	if err != nil {
		return fmt.Errorf("fail to get events: %w", err)
	}
	ctx.EventManager().EmitEvents(events)
	return nil
}

// EmitRefundEvent emit refund event , save it to local key value store and also emit through event manager
func (m *EventMgr) EmitRefundEvent(ctx sdk.Context, keeper Keeper, refundEvt EventRefund, status EventStatus) error {
	buf, err := json.Marshal(refundEvt)
	if err != nil {
		return fmt.Errorf("fail to marshal refund event: %w", err)
	}
	event := NewEvent(refundEvt.Type(), ctx.BlockHeight(), refundEvt.InTx, buf, status)
	event.Fee = refundEvt.Fee
	if err := keeper.UpsertEvent(ctx, event); err != nil {
		return fmt.Errorf("fail to save refund event: %w", err)
	}
	events, err := refundEvt.Events()
	if err != nil {
		return fmt.Errorf("fail to get events: %w", err)
	}
	ctx.EventManager().EmitEvents(events)
	return nil
}

func (m *EventMgr) EmitBondEvent(ctx sdk.Context, keeper Keeper, bondEvent EventBond) error {
	buf, err := json.Marshal(bondEvent)
	if err != nil {
		return fmt.Errorf("fail to marshal bond event: %w", err)
	}

	e := NewEvent(bondEvent.Type(), ctx.BlockHeight(), bondEvent.TxIn, buf, EventSuccess)
	if err := keeper.UpsertEvent(ctx, e); err != nil {
		return fmt.Errorf("fail to save bond event: %w", err)
	}
	events, err := bondEvent.Events()
	if err != nil {
		return fmt.Errorf("fail to get events: %w", err)
	}
	ctx.EventManager().EmitEvents(events)
	return nil
}

// EmitUnstakeEvent save unstake event to local key value store , and also add it to event manager
func (m *EventMgr) EmitUnstakeEvent(ctx sdk.Context, keeper Keeper, unstakeEvt EventUnstake) error {
	unstakeBytes, err := json.Marshal(unstakeEvt)
	if err != nil {
		return fmt.Errorf("fail to marshal unstake event: %w", err)
	}

	// unstake event is pending , once signer send the fund to customer successfully, then this should be marked as success
	evt := NewEvent(
		unstakeEvt.Type(),
		ctx.BlockHeight(),
		unstakeEvt.InTx,
		unstakeBytes,
		EventPending,
	)

	if err := keeper.UpsertEvent(ctx, evt); err != nil {
		return fmt.Errorf("fail to save unstake event: %w", err)
	}
	events, err := unstakeEvt.Events()
	if err != nil {
		return fmt.Errorf("fail to get events: %w", err)
	}
	ctx.EventManager().EmitEvents(events)
	return nil
}

// EmitAddEvent save add event to local key value store , and add it to event manager
func (m *EventMgr) EmitAddEvent(ctx sdk.Context, keeper Keeper, addEvt EventAdd) error {
	buf, err := json.Marshal(addEvt)
	if err != nil {
		return fmt.Errorf("fail to marshal add event: %w", err)
	}
	evt := NewEvent(
		addEvt.Type(),
		ctx.BlockHeight(),
		addEvt.InTx,
		buf,
		EventSuccess,
	)
	if err := keeper.UpsertEvent(ctx, evt); err != nil {
		return sdk.ErrInternal(fmt.Errorf("fail to save event: %w", err).Error())
	}
	events, err := addEvt.Events()
	if err != nil {
		return fmt.Errorf("fail to get events: %w", err)
	}
	ctx.EventManager().EmitEvents(events)
	return nil
}

// EmitSlashEvent
func (m *EventMgr) EmitSlashEvent(ctx sdk.Context, keeper Keeper, slashEvt EventSlash) error {
	slashBuf, err := json.Marshal(slashEvt)
	if err != nil {
		return fmt.Errorf("fail to marshal slash event to buf: %w", err)
	}
	event := NewEvent(
		slashEvt.Type(),
		ctx.BlockHeight(),
		common.Tx{ID: common.BlankTxID},
		slashBuf,
		EventSuccess,
	)
	if err := keeper.UpsertEvent(ctx, event); err != nil {
		return fmt.Errorf("fail to save event: %w", err)
	}
	events, err := slashEvt.Events()
	if err != nil {
		return fmt.Errorf("fail to get events: %w", err)
	}
	ctx.EventManager().EmitEvents(events)
	return nil
}

// EmitFeeEvent emit a fee event through event manager
func (m *EventMgr) EmitFeeEvent(ctx sdk.Context, keeper Keeper, feeEvent EventFee) error {
	if err := updateEventFee(ctx, keeper, feeEvent.TxID, feeEvent.Fee); err != nil {
		return fmt.Errorf("fail to update event fee: %w", err)
	}
	events, err := feeEvent.Events()
	if err != nil {
		return fmt.Errorf("fail to emit fee event: %w", err)
	}
	ctx.EventManager().EmitEvents(events)
	return nil
}

// EmitOutboundEvent emit an outbound event
func (m *EventMgr) EmitOutboundEvent(ctx sdk.Context, outbound EventOutbound) error {
	events, err := outbound.Events()
	if err != nil {
		return fmt.Errorf("fail to emit outbound event: %w", err)
	}
	ctx.EventManager().EmitEvents(events)
	return nil
}
