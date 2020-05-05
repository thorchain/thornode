package thorchain

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
)

type KeeperSwapQueue interface {
	SetSwapQueueItem(ctx sdk.Context, msg MsgSwap) error
	GetSwapQueueIterator(ctx sdk.Context) sdk.Iterator
	GetSwapQueueItem(ctx sdk.Context, txID common.TxID) (MsgSwap, error)
	RemoveSwapQueueItem(ctx sdk.Context, txID common.TxID)
}

// SetSwapQueueItem - writes a swap item to the kvstore
func (k KVStore) SetSwapQueueItem(ctx sdk.Context, msg MsgSwap) error {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixSwapQueueItem, msg.Tx.ID.String())
	buf, err := k.cdc.MarshalBinaryBare(msg)
	if err != nil {
		return dbError(ctx, "fail to marshal swap item to binary", err)
	}
	store.Set([]byte(key), buf)
	return nil
}

// GetSwapQueueIterator iterate tx out
func (k KVStore) GetSwapQueueIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixSwapQueueItem))
}

// GetSwapQueueItem - write the given swap queue item information to key values tore
func (k KVStore) GetSwapQueueItem(ctx sdk.Context, txID common.TxID) (MsgSwap, error) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixSwapQueueItem, txID.String())
	if !store.Has([]byte(key)) {
		return MsgSwap{}, errors.New("not found")
	}
	var msg MsgSwap
	buf := store.Get([]byte(key))
	if err := k.cdc.UnmarshalBinaryBare(buf, &msg); err != nil {
		return msg, dbError(ctx, "fail to unmarshal swap queue item", err)
	}
	return msg, nil
}

// RemoveSwapQueueItem - removes a swap item to the kvstore
func (k KVStore) RemoveSwapQueueItem(ctx sdk.Context, txID common.TxID) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixSwapQueueItem, txID.String())
	store.Delete([]byte(key))
}
