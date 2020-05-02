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
	ProcessGas(ctx sdk.Context, keeper Keeper)
	GetGas() common.Gas
}

// GasManangerImp implement a GasManager which will store the gas related events happened in thorchain in memory
// emit GasEvent every block if there are any
type GasMgr struct {
	gasEvent *EventGas
	gas      common.Gas
}

// NewGasMgr create a new instance of GasManager
func NewGasMgr() *GasMgr {
	return &GasMgr{
		gasEvent: NewEventGas(),
		gas:      common.Gas{},
	}
}

// BeginBlock when a new block created , update the internal EventGas to new one
func (gm *GasMgr) BeginBlock() {
	gm.gasEvent = NewEventGas()
	gm.gas = common.Gas{}
}

// AddGasAsset to the EventGas
func (gm *GasMgr) AddGasAsset(gas common.Gas) {
	gm.gas = gm.gas.Add(gas)
}

func (gm *GasMgr) GetGas() common.Gas {
	return gm.gas
}

// EndBlock emit the events
func (gm *GasMgr) EndBlock(ctx sdk.Context, keeper Keeper) {
	gm.ProcessGas(ctx, keeper)

	if len(gm.gasEvent.Pools) == 0 {
		return
	}

	buf, err := json.Marshal(gm.gasEvent)
	if err != nil {
		ctx.Logger().Error("fail to marshal gas event", "error", err)
		return
	}
	evt := NewEvent(gm.gasEvent.Type(), ctx.BlockHeight(), common.Tx{ID: common.BlankTxID}, buf, EventSuccess)
	if err := keeper.UpsertEvent(ctx, evt); err != nil {
		ctx.Logger().Error("fail to upsert event", "error", err)
	}
}

// ProcessGas to subsidise the pool with RUNE for the gas they have spent
func (gm *GasMgr) ProcessGas(ctx sdk.Context, keeper Keeper) {
	vault, err := keeper.GetVaultData(ctx)
	if err != nil {
		ctx.Logger().Error("fail to get vault data", "error", err)
		return
	}
	for _, gas := range gm.gas {
		// if the coin is zero amount, don't need to do anything
		if gas.Amount.IsZero() {
			continue
		}

		pool, err := keeper.GetPool(ctx, gas.Asset)
		if err != nil {
			ctx.Logger().Error("fail to get pool", "pool", gas.Asset, "error", err)
			continue
		}
		if err := pool.Valid(); err != nil {
			ctx.Logger().Error("fail to get pool", "pool", gas.Asset, "error", err)
			continue
		}
		runeGas := pool.AssetValueInRune(gas.Amount) // Convert to Rune (gas will never be RUNE)
		// If Rune owed now exceeds the Total Reserve, return it all
		if runeGas.GT(vault.TotalReserve) {
			continue // looks like we don't have enough in reserve to pay the fee
		}
		vault.TotalReserve = common.SafeSub(vault.TotalReserve, runeGas) // Deduct from the Reserve.
		pool.BalanceRune = pool.BalanceRune.Add(runeGas)                 // Add to the pool
		pool.BalanceAsset = common.SafeSub(pool.BalanceAsset, gas.Amount)

		if err := keeper.SetPool(ctx, pool); err != nil {
			ctx.Logger().Error("fail to set pool", "pool", gas.Asset, "error", err)
			continue
		}

		gasPool := GasPool{
			Asset:    gas.Asset,
			AssetAmt: gas.Amount,
			RuneAmt:  runeGas,
		}
		gm.gasEvent.UpsertGasPool(gasPool)
	}

	if err := keeper.SetVaultData(ctx, vault); err != nil {
		ctx.Logger().Error("fail to set vault data", "error", err)
	}
}
