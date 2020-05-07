package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type GasManager interface {
	BeginBlock()
	EndBlock(ctx sdk.Context, keeper Keeper, eventManager EventManager)
	AddGasAsset(gas common.Gas)
	ProcessGas(ctx sdk.Context, keeper Keeper)
	GetGas() common.Gas
}

// GasManangerImp implement a GasManager which will store the gas related events happened in thorchain in memory
// emit GasEvent every block if there are any
type GasMgr struct {
	gasEvent *EventGas
	gas      common.Gas
	gasCount map[common.Asset]int64
}

// NewGasMgr create a new instance of GasManager
func NewGasMgr() *GasMgr {
	return &GasMgr{
		gasEvent: NewEventGas(),
		gas:      common.Gas{},
		gasCount: make(map[common.Asset]int64, 0),
	}
}

// BeginBlock when a new block created , update the internal EventGas to new one
func (gm *GasMgr) BeginBlock() {
	gm.gasEvent = NewEventGas()
	gm.gas = common.Gas{}
	gm.gasCount = make(map[common.Asset]int64, 0)
}

// AddGasAsset to the EventGas
func (gm *GasMgr) AddGasAsset(gas common.Gas) {
	gm.gas = gm.gas.Add(gas)
	for _, coin := range gas {
		gm.gasCount[coin.Asset] += 1
	}
}

func (gm *GasMgr) GetGas() common.Gas {
	return gm.gas
}

// EndBlock emit the events
func (gm *GasMgr) EndBlock(ctx sdk.Context, keeper Keeper, eventManager EventManager) {
	gm.ProcessGas(ctx, keeper)

	if len(gm.gasEvent.Pools) == 0 {
		return
	}

	if err := eventManager.EmitGasEvent(ctx, keeper, gm.gasEvent); nil != err {
		ctx.Logger().Error("fail to emit gas event", "error", err)
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
		if common.RuneAsset().Chain.Equals(common.THORChain) {
			if runeGas.LT(keeper.GetRuneBalaceOfModule(ctx, ReserveName)) {
				coin := common.NewCoin(common.RuneNative, runeGas)
				if err := keeper.SendFromModuleToModule(ctx, ReserveName, AsgardName, coin); err != nil {
					ctx.Logger().Error("fail to transfer funds from reserve to asgard", "pool", gas.Asset, "error", err)
					continue
				}
				pool.BalanceRune = pool.BalanceRune.Add(runeGas) // Add to the pool
			}
		} else {
			if runeGas.LT(vault.TotalReserve) {
				vault.TotalReserve = common.SafeSub(vault.TotalReserve, runeGas) // Deduct from the Reserve.
				pool.BalanceRune = pool.BalanceRune.Add(runeGas)                 // Add to the pool
			}
		}

		pool.BalanceAsset = common.SafeSub(pool.BalanceAsset, gas.Amount)

		if err := keeper.SetPool(ctx, pool); err != nil {
			ctx.Logger().Error("fail to set pool", "pool", gas.Asset, "error", err)
			continue
		}

		gasPool := GasPool{
			Asset:    gas.Asset,
			AssetAmt: gas.Amount,
			RuneAmt:  runeGas,
			Count:    gm.gasCount[gas.Asset],
		}
		gm.gasEvent.UpsertGasPool(gasPool)
	}

	if err := keeper.SetVaultData(ctx, vault); err != nil {
		ctx.Logger().Error("fail to set vault data", "error", err)
	}
}
