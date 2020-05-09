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

func (m *DummyEventMgr) CompleteEvents(ctx sdk.Context, keeper Keeper, height int64, txID common.TxID, txs common.Txs, eventStatus EventStatus) {
}

func (m *DummyEventMgr) EmitPoolEvent(ctx sdk.Context, keeper Keeper, txIn common.TxID, status EventStatus, poolEvt EventPool) error {
	return nil
}

func (m *DummyEventMgr) EmitErrataEvent(ctx sdk.Context, keeper Keeper, txIn common.TxID, errataEvent EventErrata) error {
	return nil
}

func (m *DummyEventMgr) EmitGasEvent(ctx sdk.Context, keeper Keeper, gasEvent *EventGas) error {
	return nil
}

func (m *DummyEventMgr) EmitStakeEvent(ctx sdk.Context, keeper Keeper, inTx common.Tx, stakeEvent EventStake) error {
	return nil
}

func (m *DummyEventMgr) EmitReserveEvent(ctx sdk.Context, keeper Keeper, reserveEvent EventReserve) error {
	return nil
}

type DummyVersionedEventMgr struct{}

func NewDummyVersionedEventMgr() *DummyVersionedEventMgr {
	return &DummyVersionedEventMgr{}
}

func (m *DummyVersionedEventMgr) GetEventManager(ctx sdk.Context, version semver.Version) (EventManager, error) {
	return NewDummyEventMgr(), nil
}
