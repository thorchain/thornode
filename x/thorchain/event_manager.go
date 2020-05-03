package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// EventManager define methods need to be support to manage events
type EventManager interface {
	BeginBlock(ctx sdk.Context)
	EndBlock(ctx sdk.Context, keeper Keeper)
	GetBlockEvents(ctx sdk.Context, keeper Keeper, height int64) (*BlockEvents, error)
	CompleteEvents(ctx sdk.Context, keeper Keeper, height int64, txID common.TxID, txs common.Txs, eventStatus EventStatus)
	AddEvent(ctx sdk.Context, event Event)
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
func (m *EventMgr) AddEvent(ctx sdk.Context, event Event) {
	m.blockEvents.AddEvent(event)
}
