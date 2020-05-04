package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type DummyGasManager struct {
}

func NewDummyGasManager() *DummyGasManager {
	return &DummyGasManager{}
}

func (m *DummyGasManager) BeginBlock()                                                        {}
func (m *DummyGasManager) EndBlock(ctx sdk.Context, keeper Keeper, eventManager EventManager) {}
func (m *DummyGasManager) AddGasAsset(gas common.Gas)                                         {}
func (m *DummyGasManager) GetGas() common.Gas                                                 { return nil }
func (m *DummyGasManager) ProcessGas(ctx sdk.Context, keeper Keeper)                          {}

type DummyVersionedGasMgr struct {
}

func NewDummyVersionedGasMgr() *DummyVersionedGasMgr {
	return &DummyVersionedGasMgr{}
}

func (m *DummyVersionedGasMgr) GetGasManager(ctx sdk.Context, version semver.Version) (GasManager, error) {
	return NewDummyGasManager(), nil
}
