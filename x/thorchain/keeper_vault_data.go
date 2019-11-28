package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
)

type KeeperVaultData interface {
	GetVaultData(ctx sdk.Context) VaultData
	SetVaultData(ctx sdk.Context, data VaultData)
	UpdateVaultData(ctx sdk.Context)
}

func (k KVStore) GetVaultData(ctx sdk.Context) VaultData {
	data := NewVaultData()
	key := k.GetKey(ctx, prefixVaultData, "")
	store := ctx.KVStore(k.storeKey)
	if store.Has([]byte(key)) {
		buf := store.Get([]byte(key))
		_ = k.cdc.UnmarshalBinaryBare(buf, &data)
	}
	return data
}

func (k KVStore) SetVaultData(ctx sdk.Context, data VaultData) {
	key := k.GetKey(ctx, prefixVaultData, "")
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(data))
}

// Update the vault data to reflect changing in this block
func (k KVStore) UpdateVaultData(ctx sdk.Context) {
	vault := k.GetVaultData(ctx)
	currentHeight := uint64(ctx.BlockHeight())

	// First get active pools and total staked Rune
	totalRune := sdk.ZeroUint()
	assets, _ := k.GetPoolIndex(ctx)
	var pools []Pool
	for _, asset := range assets {
		pool := k.GetPool(ctx, asset)
		if pool.IsEnabled() && !pool.BalanceRune.IsZero() {
			totalRune = totalRune.Add(pool.BalanceRune)
			pools = append(pools, pool)
		}
	}

	if totalRune.IsZero() {
		return // If no Rune is staked, then don't give out block rewards.
	}

	// Then get fees and rewards
	totalFees, _ := k.GetTotalLiquidityFees(ctx, currentHeight)
	bondReward, totalPoolRewards, stakerDeficit := calcBlockRewards(vault.TotalReserve, totalFees)

	if !vault.TotalReserve.IsZero() {
		// Move Rune from the Reserve to the Bond and Pool Rewards
		if vault.TotalReserve.LT(totalPoolRewards) {
			vault.TotalReserve = sdk.ZeroUint()
		} else {
			vault.TotalReserve = common.SafeSub(common.SafeSub(vault.TotalReserve, bondReward), totalPoolRewards) // Subtract Bond and Pool rewards
		}
		vault.BondRewardRune = vault.BondRewardRune.Add(bondReward) // Add here for individual Node collection later
	}

	if !totalPoolRewards.IsZero() { // If Pool Rewards to hand out
		// First subsidise the gas that was consumed
		for _, coin := range vault.Gas {
			pool := k.GetPool(ctx, coin.Asset)
			runeGas := pool.AssetValueInRune(coin.Amount)
			pool.BalanceRune = pool.BalanceRune.Add(runeGas)
			k.SetPool(ctx, pool)
			totalPoolRewards = common.SafeSub(totalPoolRewards, runeGas)
		}

		// Then add pool rewards
		poolRewards := calcPoolRewards(totalPoolRewards, totalRune, pools)
		for i, reward := range poolRewards {
			pools[i].BalanceRune = pools[i].BalanceRune.Add(reward)
			k.SetPool(ctx, pools[i])
		}
	} else { // Else deduct pool deficit
		// Get total fees, then find individual pool deficits, then deduct
		totalFees, _ = k.GetTotalLiquidityFees(ctx, currentHeight)
		for _, pool := range pools {
			poolFees, _ := k.GetPoolLiquidityFees(ctx, currentHeight, pool.Asset)
			if !pool.BalanceRune.IsZero() || !poolFees.IsZero() { // Safety checks
				continue
			}
			poolDeficit := calcPoolDeficit(stakerDeficit, totalFees, poolFees)
			pool.BalanceRune = common.SafeSub(pool.BalanceRune, poolDeficit)
			k.SetPool(ctx, pool)
		}
	}

	i, _ := k.TotalActiveNodeAccount(ctx)
	vault.TotalBondUnits = vault.TotalBondUnits.Add(sdk.NewUint(uint64(i))) // Add 1 unit for each active Node

	k.SetVaultData(ctx, vault)
}
