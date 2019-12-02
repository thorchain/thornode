package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperPool interface {
	GetPoolIterator(ctx sdk.Context) sdk.Iterator
	GetPool(ctx sdk.Context, asset common.Asset) (Pool, error)
	SetPool(ctx sdk.Context, pool Pool) error
	EnableAPool(ctx sdk.Context)
	PoolExist(ctx sdk.Context, asset common.Asset) bool
}

// GetPoolIterator iterate pools
func (k KVStore) GetPoolIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixPool))
}

// GetPool get the entire Pool metadata struct for a pool ID
func (k KVStore) GetPool(ctx sdk.Context, asset common.Asset) (Pool, error) {
	key := k.GetKey(ctx, prefixPool, asset.String())
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return NewPool(), nil
	}
	buf := store.Get([]byte(key))
	var pool Pool
	if err := k.cdc.UnmarshalBinaryBare(buf, &pool); err != nil {
		return NewPool(), dbError(ctx, "Unmarshal: pool", err)
	}
	return pool, nil
}

// Sets the entire Pool metadata struct for a pool ID
func (k KVStore) SetPool(ctx sdk.Context, pool Pool) error {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixPool, pool.Asset.String())

	prePool, err := k.GetPool(ctx, pool.Asset)
	if err != nil {
		return err
	}

	if prePool.Status != pool.Status {
		eventPoolStatusWrapper(ctx, k, pool)
	}
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(pool))
	return nil
}

// Picks the most "deserving" pool (by most staked rune) to be enabled and
// enables it
func (k KVStore) EnableAPool(ctx sdk.Context) {
	var pools []Pool
	iterator := k.GetPoolIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var pool Pool
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &pool)
		if pool.Status == PoolBootstrap {
			pools = append(pools, pool)
		}
	}

	if len(pools) > 0 {
		pool := pools[0]
		for _, p := range pools {
			if pool.BalanceRune.LT(p.BalanceRune) {
				pool = p
			}
		}
		// ensure THORNode don't enable a pool that doesn't have any rune or assets
		if pool.BalanceAsset.IsZero() || pool.BalanceRune.IsZero() {
			return
		}
		pool.Status = PoolEnabled
		k.SetPool(ctx, pool)

	}
}

// PoolExist check whether the given pool exist in the datastore
func (k KVStore) PoolExist(ctx sdk.Context, asset common.Asset) bool {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixPool, asset.String())
	return store.Has([]byte(key))
}
