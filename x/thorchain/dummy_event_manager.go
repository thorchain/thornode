package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// DummyEventMgr used for test purpose , and it implement EventManager interface
type DummyEventMgr struct {
}

func NewDummyEventMgr() *DummyEventMgr {
	return &DummyEventMgr{}
}

func (m *DummyEventMgr) BeginBlock(ctx sdk.Context) {
}

func (m *DummyEventMgr) EndBlock(ctx sdk.Context, keeper Keeper) {
}

func (m *DummyEventMgr) GetBlockEvents(ctx sdk.Context, keeper Keeper, height int64) (*BlockEvents, error) {
	return nil, nil
}

func (m *DummyEventMgr) CompleteEvents(ctx sdk.Context, keeper Keeper, height int64, txID common.TxID, txs common.Txs, eventStatus EventStatus) {
}

func (m *DummyEventMgr) AddEvent(ctx sdk.Context, event Event) {
}

type DummyVersionedEventMgr struct{}

func NewDummyVersionedEventMgr() *DummyVersionedEventMgr {
	return &DummyVersionedEventMgr{}
}

func (m *DummyVersionedEventMgr) GetEventManager(ctx sdk.Context, version semver.Version) (EventManager, error) {
	return NewDummyEventMgr(), nil
}
