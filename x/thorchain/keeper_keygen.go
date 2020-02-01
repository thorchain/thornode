package thorchain

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KeeperKeygen interface {
	SetKeygenBlock(ctx sdk.Context, keygenBlock KeygenBlock) error
	GetKeygenBlockIterator(ctx sdk.Context) sdk.Iterator
	GetKeygenBlock(ctx sdk.Context, height int64) (KeygenBlock, error)
}

// SetKeygenBlock save the KeygenBlock to kv store
func (k KVStore) SetKeygenBlock(ctx sdk.Context, keygen KeygenBlock) error {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixKeygen, strconv.FormatInt(keygen.Height, 10))
	buf, err := k.cdc.MarshalBinaryBare(keygen)
	if err != nil {
		return dbError(ctx, "fail to marshal keygen block", err)
	}
	store.Set([]byte(key), buf)
	return nil
}

// GetKeygenBlockIterator return an iterator
func (k KVStore) GetKeygenBlockIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixKeygen))
}

// GetKeygenBlock from a given height
func (k KVStore) GetKeygenBlock(ctx sdk.Context, height int64) (KeygenBlock, error) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixKeygen, strconv.FormatInt(height, 10))
	if !store.Has([]byte(key)) {
		return NewKeygenBlock(height), nil
	}
	buf := store.Get([]byte(key))
	var keygenBlock KeygenBlock
	if err := k.cdc.UnmarshalBinaryBare(buf, &keygenBlock); err != nil {
		return KeygenBlock{}, dbError(ctx, "fail to unmarshal keygen block", err)
	}
	return keygenBlock, nil
}
