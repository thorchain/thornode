package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KeeperMimir interface {
	GetMimir(_ sdk.Context, key string) (int64, error)
	SetMimir(_ sdk.Context, key string, value int64)
	GetMimirIterator(ctx sdk.Context) sdk.Iterator
}

func (k KVStore) GetMimir(ctx sdk.Context, key string) (int64, error) {
	key = k.GetKey(ctx, prefixMimir, key)
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return -1, nil
	}
	var value int64
	buf := store.Get([]byte(key))
	err := k.cdc.UnmarshalBinaryBare(buf, &value)
	if err != nil {
		return -1, dbError(ctx, "Unmarshal: mimir attr", err)
	}
	return value, nil
}

func (k KVStore) SetMimir(ctx sdk.Context, key string, value int64) {
	store := ctx.KVStore(k.storeKey)
	key = k.GetKey(ctx, prefixMimir, key)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(value))
}

// GetMimirIterator iterate gas units
func (k KVStore) GetMimirIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixMimir))
}
