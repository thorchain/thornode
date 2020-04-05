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

// GasManangerImp implement a GasManager which will store the gas related events happened in thorchain in memory
// emit GasEvent every block if there are any
type GasMgr struct {
	gasEvent *EventGas
}

// NewGasMgr create a new instance of GasManager
func NewGasMgr() *GasMgr {
	return &GasMgr{
		gasEvent: NewEventGas(),
	}
}

// BeginBlock when a new block created , update the internal EventGas to new one
func (gm *GasMgr) BeginBlock() {
	gm.gasEvent = NewEventGas()
}

// AddGasAsset to the EventGas
func (gm *GasMgr) AddGasAsset(gas common.Gas) {
	for _, g := range gas {
		if g.IsEmpty() {
			continue
		}
		gasPool := GasPool{
			Asset:    g.Asset,
			AssetAmt: g.Amount,
			RuneAmt:  sdk.ZeroUint(),
		}
		gm.gasEvent.UpsertGasPool(gasPool)
	}
}

//  AddRune to the gas event
func (gm *GasMgr) AddRune(asset common.Asset, amt sdk.Uint) {
	gasPool := GasPool{
		Asset:    asset,
		AssetAmt: sdk.ZeroUint(),
		RuneAmt:  amt,
	}
	gm.gasEvent.UpsertGasPool(gasPool)
}

// EndBlock emit the events
func (gm *GasMgr) EndBlock(ctx sdk.Context, keeper Keeper) {
	if len(gm.gasEvent.Pools) == 0 {
		return
	}
	buf, err := json.Marshal(gm.gasEvent)
	if err != nil {
		ctx.Logger().Error("fail to marshal gas event", "error", err)
		return
	}
	evt := NewEvent(gm.gasEvent.Type(), ctx.BlockHeight(),
		common.Tx{ID: common.BlankTxID},
		buf, EventSuccess)
	if err := keeper.UpsertEvent(ctx, evt); err != nil {
		ctx.Logger().Error("fail to upsert event", "error", err)
	}
}
