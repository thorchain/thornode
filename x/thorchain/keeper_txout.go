package thorchain

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

type KeeperTxOut interface {
	SetTxOut(ctx sdk.Context, blockOut *TxOut)
	GetTxOutIterator(ctx sdk.Context) sdk.Iterator
	GetTxOut(ctx sdk.Context, height uint64) (*TxOut, error)
}

// SetTxOut - write the given txout information to key values tore
func (k KVStore) SetTxOut(ctx sdk.Context, blockOut *TxOut) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixTxOut, strconv.FormatUint(blockOut.Height, 10))
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(blockOut))
}

// GetTxOutIterator iterate tx out
func (k KVStore) GetTxOutIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixTxOut))
}

// GetTxOut - write the given txout information to key values tore
func (k KVStore) GetTxOut(ctx sdk.Context, height uint64) (*TxOut, error) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixTxOut, strconv.FormatUint(height, 10))
	if !store.Has([]byte(key)) {
		return NewTxOut(height), nil
	}
	buf := store.Get([]byte(key))
	var txOut TxOut
	if err := k.cdc.UnmarshalBinaryBare(buf, &txOut); nil != err {
		return nil, errors.Wrap(err, "fail to unmarshal tx out")
	}
	return &txOut, nil
}
