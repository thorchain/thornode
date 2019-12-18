package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperPoolStaker interface {
	GetPoolStakerIterator(ctx sdk.Context) sdk.Iterator
	GetPoolStaker(ctx sdk.Context, asset common.Asset) (PoolStaker, error)
	SetPoolStaker(ctx sdk.Context, ps PoolStaker)
}

// GetPoolStakerIterator iterate pool stakers
func (k KVStore) GetPoolStakerIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixPoolStaker))
}

// GetPoolStaker retrieve poolStaker from the data store
func (k KVStore) GetPoolStaker(ctx sdk.Context, asset common.Asset) (PoolStaker, error) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixPoolStaker, asset.String())
	if !store.Has([]byte(key)) {
		ctx.Logger().Info("NotExist", "poolstakerkey", key)
		return NewPoolStaker(asset, sdk.ZeroUint()), nil
	}
	var ps PoolStaker
	buf := store.Get([]byte(key))
	if err := k.cdc.UnmarshalBinaryBare(buf, &ps); nil != err {
		ctx.Logger().Error("fail to unmarshal poolstaker", err)
		return PoolStaker{}, err
	}
	return ps, nil
}

// SetPoolStaker store the poolstaker to datastore
func (k KVStore) SetPoolStaker(ctx sdk.Context, ps PoolStaker) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixPoolStaker, ps.Asset.String())
	ctx.Logger().Info(fmt.Sprintf("key:%s ,pool staker:%s", key, ps))
	result := k.cdc.MustMarshalBinaryBare(ps)
	store.Set([]byte(key), result)
}
