package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type DummyObserverManager struct {
}

func NewDummyObserverManager() *DummyObserverManager {
	return &DummyObserverManager{}
}

func (m *DummyObserverManager) BeginBlock()                                               {}
func (m *DummyObserverManager) EndBlock(ctx sdk.Context, keeper Keeper)                   {}
func (m *DummyObserverManager) AppendObserver(chain common.Chain, addrs []sdk.AccAddress) {}
func (m *DummyObserverManager) List() []sdk.AccAddress                                    { return nil }

type DummyVersionedObserverMgr struct {
}

func NewDummyVersionedObserverMgr() *DummyVersionedObserverMgr {
	return &DummyVersionedObserverMgr{}
}

func (m *DummyVersionedObserverMgr) GetObserverManager(ctx sdk.Context, version semver.Version) (ObserverManager, error) {
	return NewDummyObserverManager(), nil
}
