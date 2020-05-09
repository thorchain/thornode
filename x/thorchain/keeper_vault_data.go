package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// KeeperVaultData func to access Vault in key value store
type KeeperVaultData interface {
	GetVaultData(ctx sdk.Context) (VaultData, error)
	SetVaultData(ctx sdk.Context, data VaultData) error
	UpdateVaultData(ctx sdk.Context, constAccessor constants.ConstantValues, gasManager GasManager, eventMgr EventManager) error
}

// GetVaultData retrieve vault data from key value store
func (k KVStore) GetVaultData(ctx sdk.Context) (VaultData, error) {
	data := NewVaultData()
	key := k.GetKey(ctx, prefixVaultData, "")
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return data, nil
	}
	buf := store.Get([]byte(key))
	if err := k.cdc.UnmarshalBinaryBare(buf, &data); err != nil {
		return data, dbError(ctx, "fail to unmarshal vault data", err)
	}

	return data, nil
}

// SetVaultData save the given vault data to key value store, it will overwrite existing vault
func (k KVStore) SetVaultData(ctx sdk.Context, data VaultData) error {
	key := k.GetKey(ctx, prefixVaultData, "")
	store := ctx.KVStore(k.storeKey)
	buf, err := k.cdc.MarshalBinaryBare(data)
	if err != nil {
		return fmt.Errorf("fail to marshal vault data: %w", err)
	}
	store.Set([]byte(key), buf)
	return nil
}

func (k KVStore) getEnabledPoolsAndTotalStakedRune(ctx sdk.Context) (Pools, sdk.Uint, error) {
	// First get active pools and total staked Rune
	totalStaked := sdk.ZeroUint()
	var pools Pools
	iterator := k.GetPoolIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var pool Pool
		if err := k.Cdc().UnmarshalBinaryBare(iterator.Value(), &pool); err != nil {
			return nil, sdk.ZeroUint(), fmt.Errorf("fail to unmarhsl pool: %w", err)
		}
		if pool.IsEnabled() && !pool.BalanceRune.IsZero() {
			totalStaked = totalStaked.Add(pool.BalanceRune)
			pools = append(pools, pool)
		}
	}
	return pools, totalStaked, nil
}

func (k KVStore) getTotalActiveBond(ctx sdk.Context) (sdk.Uint, error) {
	totalBonded := sdk.ZeroUint()
	nodes, err := k.ListActiveNodeAccounts(ctx)
	if err != nil {
		return sdk.ZeroUint(), fmt.Errorf("fail to get all active accounts: %w", err)
	}
	for _, node := range nodes {
		totalBonded = totalBonded.Add(node.Bond)
	}
	return totalBonded, nil
}

// UpdateVaultData Update the vault data to reflect changing in this block
// TODO: there is way too much business logic her for a keeper function. Move
// to its own file/manager
func (k KVStore) UpdateVaultData(ctx sdk.Context, constAccessor constants.ConstantValues, gasManager GasManager, eventMgr EventManager) error {
	vaultData, err := k.GetVaultData(ctx)
	if err != nil {
		return fmt.Errorf("fail to get existing vault data: %w", err)
	}

	totalReserve := sdk.ZeroUint()
	if common.RuneAsset().Chain.Equals(common.THORChain) {
		totalReserve = k.GetRuneBalaceOfModule(ctx, ReserveName)
	} else {
		totalReserve = vaultData.TotalReserve
	}

	// when total reserve is zero , can't pay reward
	if totalReserve.IsZero() {
		return nil
	}
	currentHeight := uint64(ctx.BlockHeight())
	pools, totalStaked, err := k.getEnabledPoolsAndTotalStakedRune(ctx)
	if err != nil {
		return fmt.Errorf("fail to get enabled pools and total staked rune: %w", err)
	}

	// If no Rune is staked, then don't give out block rewards.
	if totalStaked.IsZero() {
		return nil // If no Rune is staked, then don't give out block rewards.
	}

	// get total liquidity fees
	totalLiquidityFees, err := k.GetTotalLiquidityFees(ctx, currentHeight)
	if err != nil {
		return fmt.Errorf("fail to get total liquidity fee: %w", err)
	}

	// NOTE: if we continue to have remaining gas to pay off (which is
	// extremely unlikely), ignore it for now (attempt to recover in the next
	// block). This should be OK as the asset amount in the pool has already
	// been deducted so the balances are correct. Just operating at a deficit.
	totalBonded, err := k.getTotalActiveBond(ctx)
	if err != nil {
		return fmt.Errorf("fail to get total active bond: %w", err)
	}
	emissionCurve := constAccessor.GetInt64Value(constants.EmissionCurve)
	blocksOerYear := constAccessor.GetInt64Value(constants.BlocksPerYear)
	bondReward, totalPoolRewards, stakerDeficit := calcBlockRewards(totalStaked, totalBonded, totalReserve, totalLiquidityFees, emissionCurve, blocksOerYear)

	// given bondReward and toolPoolRewards are both calculated base on totalReserve, thus it should always have enough to pay the bond reward

	// Move Rune from the Reserve to the Bond and Pool Rewards
	totalReserve = common.SafeSub(totalReserve, bondReward.Add(totalPoolRewards))
	if common.RuneAsset().Chain.Equals(common.THORChain) {
		coin := common.NewCoin(common.RuneNative, bondReward)
		if err := k.SendFromModuleToModule(ctx, ReserveName, BondName, coin); err != nil {
			ctx.Logger().Error("fail to transfer funds from reserve to bond", "error", err)
			return fmt.Errorf("fail to transfer funds from reserve to bond: %w", err)
		}
	} else {
		vaultData.TotalReserve = totalReserve
	}
	vaultData.BondRewardRune = vaultData.BondRewardRune.Add(bondReward) // Add here for individual Node collection later

	var evtPools []PoolAmt

	if !totalPoolRewards.IsZero() { // If Pool Rewards to hand out

		var rewardAmts []sdk.Uint
		// Pool Rewards are based on Fee Share
		for _, pool := range pools {
			fees, err := k.GetPoolLiquidityFees(ctx, currentHeight, pool.Asset)
			if err != nil {
				err = fmt.Errorf("fail to get fees: %w", err)
				ctx.Logger().Error(err.Error())
				return err
			}
			amt := common.GetShare(fees, totalLiquidityFees, totalPoolRewards)
			rewardAmts = append(rewardAmts, amt)
			evtPools = append(evtPools, PoolAmt{Asset: pool.Asset, Amount: int64(amt.Uint64())})
		}
		// Pay out
		if err := payPoolRewards(ctx, k, rewardAmts, pools); err != nil {
			return err
		}

	} else { // Else deduct pool deficit

		for _, pool := range pools {
			poolFees, err := k.GetPoolLiquidityFees(ctx, currentHeight, pool.Asset)
			if err != nil {
				return fmt.Errorf("fail to get liquidity fees for pool(%s): %w", pool.Asset, err)
			}
			if pool.BalanceRune.IsZero() || poolFees.IsZero() { // Safety checks
				continue
			}
			poolDeficit := calcPoolDeficit(stakerDeficit, totalLiquidityFees, poolFees)
			if common.RuneAsset().Chain.Equals(common.THORChain) {
				coin := common.NewCoin(common.RuneNative, poolDeficit)
				if err := k.SendFromModuleToModule(ctx, AsgardName, BondName, coin); err != nil {
					ctx.Logger().Error("fail to transfer funds from asgard to bond", "error", err)
					return fmt.Errorf("fail to transfer funds from asgard to bond: %w", err)
				}
			}
			pool.BalanceRune = common.SafeSub(pool.BalanceRune, poolDeficit)
			vaultData.BondRewardRune = vaultData.BondRewardRune.Add(poolDeficit)
			if err := k.SetPool(ctx, pool); err != nil {
				err = fmt.Errorf("fail to set pool: %w", err)
				ctx.Logger().Error(err.Error())
				return err
			}
			evtPools = append(evtPools, PoolAmt{
				Asset:  pool.Asset,
				Amount: 0 - int64(poolDeficit.Uint64()),
			})
		}
	}

	rewardEvt := NewEventRewards(bondReward, evtPools)
	if err := eventMgr.EmitRewardEvent(ctx, k, rewardEvt); err != nil {
		return fmt.Errorf("fail to emit reward event: %w", err)
	}
	i, err := getTotalActiveNodeWithBond(ctx, k)
	if err != nil {
		return fmt.Errorf("fail to get total active node account: %w", err)
	}
	vaultData.TotalBondUnits = vaultData.TotalBondUnits.Add(sdk.NewUint(uint64(i))) // Add 1 unit for each active Node

	return k.SetVaultData(ctx, vaultData)
}

func getTotalActiveNodeWithBond(ctx sdk.Context, k Keeper) (int64, error) {
	nas, err := k.ListActiveNodeAccounts(ctx)
	if err != nil {
		return 0, fmt.Errorf("fail to get active node accounts: %w", err)
	}
	var total int64
	for _, item := range nas {
		if !item.Bond.IsZero() {
			total++
		}
	}
	return total, nil
}

// Pays out Rewards
func payPoolRewards(ctx sdk.Context, k Keeper, poolRewards []sdk.Uint, pools Pools) error {
	for i, reward := range poolRewards {
		pools[i].BalanceRune = pools[i].BalanceRune.Add(reward)
		if err := k.SetPool(ctx, pools[i]); err != nil {
			err = fmt.Errorf("fail to set pool: %w", err)
			ctx.Logger().Error(err.Error())
			return err
		}
		if common.RuneAsset().Chain.Equals(common.THORChain) {
			coin := common.NewCoin(common.RuneNative, reward)
			if err := k.SendFromModuleToModule(ctx, ReserveName, AsgardName, coin); err != nil {
				ctx.Logger().Error("fail to transfer funds from reserve to asgard", "error", err)
				return fmt.Errorf("fail to transfer funds from reserve to asgard: %w", err)
			}
		}
	}
	return nil
}
