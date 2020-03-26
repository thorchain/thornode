package thorchain

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KeeperTxOut interface {
	SetTxOut(ctx sdk.Context, blockOut *TxOut) error
	AppendTxOut(ctx sdk.Context, height int64, item *TxOutItem) error
	GetTxOutIterator(ctx sdk.Context) sdk.Iterator
	GetTxOut(ctx sdk.Context, height int64) (*TxOut, error)
}

// AppendTxOut - append a given item to txOut
func (k KVStore) AppendTxOut(ctx sdk.Context, height int64, item *TxOutItem) error {
	block, err := k.GetTxOut(ctx, height)
	if err != nil {
		return err
	}
	block.TxArray = append(block.TxArray, item)
	return k.SetTxOut(ctx, block)
}

// SetTxOut - write the given txout information to key values tore
func (k KVStore) SetTxOut(ctx sdk.Context, blockOut *TxOut) error {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixTxOut, strconv.FormatInt(blockOut.Height, 10))
	buf, err := k.cdc.MarshalBinaryBare(blockOut)
	if err != nil {
		return dbError(ctx, "fail to marshal tx out to binary", err)
	}
	store.Set([]byte(key), buf)
	return nil
}

// GetTxOutIterator iterate tx out
func (k KVStore) GetTxOutIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixTxOut))
}

// GetTxOut - write the given txout information to key values tore
func (k KVStore) GetTxOut(ctx sdk.Context, height int64) (*TxOut, error) {
	txOut := NewTxOut(height)
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixTxOut, strconv.FormatInt(height, 10))
	if !store.Has([]byte(key)) {
		return txOut, nil
	}
	buf := store.Get([]byte(key))
	if err := k.cdc.UnmarshalBinaryBare(buf, txOut); err != nil {
		return txOut, dbError(ctx, "fail to unmarshal tx out", err)
	}
	return txOut, nil
}
