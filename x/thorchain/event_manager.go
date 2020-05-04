package thorchain

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// EventManager define methods need to be support to manage events
type EventManager interface {
	BeginBlock(ctx sdk.Context)
	EndBlock(ctx sdk.Context, keeper Keeper)
	GetBlockEvents(ctx sdk.Context, keeper Keeper, height int64) (*BlockEvents, error)
	CompleteEvents(ctx sdk.Context, keeper Keeper, height int64, txID common.TxID, txs common.Txs, eventStatus EventStatus)
	AddEvent(event Event)
	EmitPoolEvent(ctx sdk.Context, keeper Keeper, txIn common.TxID, status EventStatus, poolEvt EventPool) error
	EmitErrataEvent(ctx sdk.Context, keeper Keeper, txIn common.TxID, errataEvent EventErrata) error
	EmitGasEvent(ctx sdk.Context, keeper Keeper, gasEvent *EventGas) error
}

// EventMgr implement EventManager interface
type EventMgr struct {
	blockEvents *BlockEvents
}

// NewEventMgr create a new instance of EventMgr
func NewEventMgr() *EventMgr {
	return &EventMgr{}
}

// BeginBlock is going to create a brand new BlockEvents
func (m *EventMgr) BeginBlock(ctx sdk.Context) {
	m.blockEvents = NewBlockEvents(ctx.BlockHeight())
}

// EndBlock will write the block event to storage
func (m *EventMgr) EndBlock(ctx sdk.Context, keeper Keeper) {
	keeper.SetBlockEvents(ctx, m.blockEvents)
}

// GetBlockEvents return the instance of block events on the given height
func (m *EventMgr) GetBlockEvents(ctx sdk.Context, keeper Keeper, height int64) (*BlockEvents, error) {
	return keeper.GetBlockEvents(ctx, height)
}

// CompleteEvents Mark an event in the given block height to the given status
func (m *EventMgr) CompleteEvents(ctx sdk.Context, keeper Keeper, height int64, txID common.TxID, txs common.Txs, eventStatus EventStatus) {
}

// AddEvent add an event to block event
func (m *EventMgr) AddEvent(event Event) {
	m.blockEvents.AddEvent(event)
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
	m.AddEvent(evt)
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
	m.AddEvent(evt)
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
	m.blockEvents.AddEvent(evt)
	return nil
}
