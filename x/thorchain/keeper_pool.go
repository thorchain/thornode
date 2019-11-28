package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/common"
)

type KeeperPool interface {
	GetPool(ctx sdk.Context, asset common.Asset) Pool
	SetPool(ctx sdk.Context, pool Pool)
	GetPoolBalances(ctx sdk.Context, asset, asset2 common.Asset) (sdk.Uint, sdk.Uint)
	SetPoolData(ctx sdk.Context, asset common.Asset, ps PoolStatus)
	GetPoolDataIterator(ctx sdk.Context) sdk.Iterator
	EnableAPool(ctx sdk.Context)
	PoolExist(ctx sdk.Context, asset common.Asset) bool
	GetPoolIndex(ctx sdk.Context) (PoolIndex, error)
	SetPoolIndex(ctx sdk.Context, pi PoolIndex)
	AddToPoolIndex(ctx sdk.Context, asset common.Asset) error
	RemoveFromPoolIndex(ctx sdk.Context, asset common.Asset) error
}

// GetPool get the entire Pool metadata struct for a pool ID
func (k KVStore) GetPool(ctx sdk.Context, asset common.Asset) Pool {
	key := k.GetKey(ctx, prefixPool, asset.String())
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return NewPool()
	}
	bz := store.Get([]byte(key))
	var pool Pool
	k.cdc.MustUnmarshalBinaryBare(bz, &pool)

	return pool
}

// Sets the entire Pool metadata struct for a pool ID
func (k KVStore) SetPool(ctx sdk.Context, pool Pool) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixPool, pool.Asset.String())
	if !store.Has([]byte(key)) {
		if err := k.AddToPoolIndex(ctx, pool.Asset); nil != err {
			ctx.Logger().Error("fail to add asset to pool index", "asset", pool.Asset, "error", err)
		}
	}

	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(pool))
}

func (k KVStore) GetPoolBalances(ctx sdk.Context, asset, asset2 common.Asset) (sdk.Uint, sdk.Uint) {
	pool := k.GetPool(ctx, asset)
	if asset2.IsRune() {
		return pool.BalanceRune, pool.BalanceAsset
	}
	return pool.BalanceAsset, pool.BalanceRune
}

// SetPoolData - sets the value string that a pool ID resolves to
func (k KVStore) SetPoolData(ctx sdk.Context, asset common.Asset, ps PoolStatus) {
	pool := k.GetPool(ctx, asset)
	pool.Status = ps
	pool.Asset = asset
	k.SetPool(ctx, pool)
}

// GetPoolDataIterator only iterate pool data
func (k KVStore) GetPoolDataIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixPool))
}

// Picks the most "deserving" pool (by most staked rune) to be enabled and
// enables it
func (k KVStore) EnableAPool(ctx sdk.Context) {
	var pools []Pool
	iterator := k.GetPoolDataIterator(ctx)
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
		// ensure we don't enable a pool that doesn't have any rune or assets
		if pool.BalanceAsset.IsZero() || pool.BalanceRune.IsZero() {
			return
		}
		pool.Status = PoolEnabled
		k.SetPool(ctx, pool)

		eventPoolStatusWrapper(ctx, k, pool)
	}
}

// PoolExist check whether the given pool exist in the datastore
func (k KVStore) PoolExist(ctx sdk.Context, asset common.Asset) bool {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixPool, asset.String())
	return store.Has([]byte(key))
}

// GetPoolIndex retrieve pool index from the data store
func (k KVStore) GetPoolIndex(ctx sdk.Context) (PoolIndex, error) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixPoolIndex, "")
	if !store.Has([]byte(key)) {
		return PoolIndex{}, nil
	}
	buf := store.Get([]byte(key))
	var pi PoolIndex
	if err := k.cdc.UnmarshalBinaryBare(buf, &pi); nil != err {
		ctx.Logger().Error(fmt.Sprintf("fail to unmarshal poolindex,err: %s", err))
		return PoolIndex{}, errors.Wrap(err, "fail to unmarshal poolindex")
	}
	return pi, nil
}

// SetPoolIndex write a pool index into datastore
func (k KVStore) SetPoolIndex(ctx sdk.Context, pi PoolIndex) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixPoolIndex, "")
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&pi))
}

// AddToPoolIndex will add the given asset into the poolindex
func (k KVStore) AddToPoolIndex(ctx sdk.Context, asset common.Asset) error {
	pi, err := k.GetPoolIndex(ctx)
	if nil != err {
		return err
	}
	for _, item := range pi {
		if item.Equals(asset) {
			// already in the pool index , don't need to add
			return nil
		}
	}
	pi = append(pi, asset)
	k.SetPoolIndex(ctx, pi)
	return nil
}

// RemoveFromPoolIndex remove the given asset from the poolIndex
func (k KVStore) RemoveFromPoolIndex(ctx sdk.Context, asset common.Asset) error {
	pi, err := k.GetPoolIndex(ctx)
	if nil != err {
		return err
	}
	var newPI PoolIndex
	for _, item := range pi {
		if !item.Equals(asset) {
			newPI = append(newPI, item)
		}
	}
	k.SetPoolIndex(ctx, pi)
	return nil
}
