package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type DummyGasManager struct {
}

func NewDummyGasManager() *DummyGasManager {
	return &DummyGasManager{}
}

func (m *DummyGasManager) BeginBlock() {
}

func (m *DummyGasManager) EndBlock(ctx sdk.Context, keeper Keeper) {
}

func (m *DummyGasManager) AddGasAsset(gas common.Gas) {
}

func (m *DummyGasManager) AddRune(asset common.Asset, amt sdk.Uint) {
}
