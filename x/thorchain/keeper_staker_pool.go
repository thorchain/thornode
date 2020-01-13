package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/common"
)

type KeeperStakerPool interface {
	GetStakerPoolIterator(ctx sdk.Context) sdk.Iterator
	GetStakerPool(ctx sdk.Context, stakerID common.Address) (StakerPool, error)
	SetStakerPool(ctx sdk.Context, sp StakerPool)
}

// GetStakerPoolIterator iterate stakers pools
func (k KVStore) GetStakerPoolIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixStakerPool))
}

// GetStakerPool get the stakerpool from key value store
func (k KVStore) GetStakerPool(ctx sdk.Context, stakerID common.Address) (StakerPool, error) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixStakerPool, stakerID.String())
	ctx.Logger().Info("get staker pool", "stakerpoolkey", key)
	if !store.Has([]byte(key)) {
		return NewStakerPool(stakerID), nil
	}
	var ps StakerPool
	buf := store.Get([]byte(key))
	if err := k.cdc.UnmarshalBinaryBare(buf, &ps); nil != err {
		ctx.Logger().Error("fail to unmarshal stakerpool", "error", err)
		return StakerPool{}, errors.Wrap(err, "fail to unmarshal stakerpool")
	}
	return ps, nil
}

// SetStakerPool save the given stakerpool object to key value store
func (k KVStore) SetStakerPool(ctx sdk.Context, sp StakerPool) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixStakerPool, sp.RuneAddress.String())
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(sp))
}
