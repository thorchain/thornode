package thorchain

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type GasManager interface {
	BeginBlock()
	EndBlock(ctx sdk.Context, keeper Keeper)
	AddGasAsset(gas common.Gas)
	AddRune(asset common.Asset, amt sdk.Uint)
}

type GasManagerImp struct {
	gasEvent *EventGas
}

// NewGasManagerImp create a new instance of GasManager
func NewGasManagerImp() *GasManagerImp {
	return &GasManagerImp{
		gasEvent: NewEventGas(),
	}
}

// BeginBlock when a new block created , update the internal EventGas to new one
func (gm *GasManagerImp) BeginBlock() {
	gm.gasEvent = NewEventGas()
}

// AddGasAsset to the EventGas
func (gm *GasManagerImp) AddGasAsset(gas common.Gas) {
	for _, g := range gas {
		gasPool := GasPool{
			Asset:    g.Asset,
			AssetAmt: g.Amount,
			RuneAmt:  sdk.ZeroUint(),
		}
		gm.gasEvent.UpsertGasPool(gasPool)
	}
}

//  AddRune to the gas event
func (gm *GasManagerImp) AddRune(asset common.Asset, amt sdk.Uint) {
	gasPool := GasPool{
		Asset:    asset,
		AssetAmt: sdk.ZeroUint(),
		RuneAmt:  amt,
	}
	gm.gasEvent.UpsertGasPool(gasPool)
}

// EndBlock emit the events
func (gm *GasManagerImp) EndBlock(ctx sdk.Context, keeper Keeper) {
	buf, err := json.Marshal(gm.gasEvent)
	if err != nil {
		ctx.Logger().Error("fail to marshal gas event", "error", err)
	}
	evt := NewEvent(gm.gasEvent.Type(), ctx.BlockHeight(),
		common.Tx{ID: common.BlankTxID},
		buf, EventSuccess)
	if err := keeper.UpsertEvent(ctx, evt); err != nil {
		ctx.Logger().Error("fail to upsert event", "error", err)
	}
}
