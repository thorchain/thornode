package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperPool interface {
	GetPoolIterator(ctx sdk.Context) sdk.Iterator
	GetPool(ctx sdk.Context, asset common.Asset) (Pool, error)
	SetPool(ctx sdk.Context, pool Pool) error
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

// PoolExist check whether the given pool exist in the datastore
func (k KVStore) PoolExist(ctx sdk.Context, asset common.Asset) bool {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixPool, asset.String())
	return store.Has([]byte(key))
}
