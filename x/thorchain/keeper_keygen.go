package thorchain

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KeeperKeygen interface {
	SetKeygens(ctx sdk.Context, blockOut Keygens) error
	GetKeygensIterator(ctx sdk.Context) sdk.Iterator
	GetKeygens(ctx sdk.Context, height uint64) (Keygens, error)
}

// SetKeygen - write the given txout information to key values tore
func (k KVStore) SetKeygens(ctx sdk.Context, keygens Keygens) error {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixKeygen, strconv.FormatUint(keygens.Height, 10))
	buf, err := k.cdc.MarshalBinaryBare(keygens)
	if nil != err {
		return dbError(ctx, "fail to marshal tx out to binary", err)
	}
	store.Set([]byte(key), buf)
	return nil
}

// GetKeygenIterator iterate tx out
func (k KVStore) GetKeygensIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixKeygen))
}

// GetKeygen - write the given txout information to key values tore
func (k KVStore) GetKeygens(ctx sdk.Context, height uint64) (Keygens, error) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixKeygen, strconv.FormatUint(height, 10))
	if !store.Has([]byte(key)) {
		return NewKeygens(height), nil
	}
	buf := store.Get([]byte(key))
	var keygens Keygens
	if err := k.cdc.UnmarshalBinaryBare(buf, &keygens); nil != err {
		return Keygens{}, dbError(ctx, "fail to unmarshal tx out", err)
	}
	return keygens, nil
}
