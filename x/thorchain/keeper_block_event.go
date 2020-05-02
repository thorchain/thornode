package thorchain

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// KeeperBlockEvent define the methods required to store BlockEvent into kvstore
type KeeperBlockEvent interface {
	GetBlockEvents(ctx sdk.Context, height int64) (*BlockEvents, error)
	GetBlockEventsIterator(ctx sdk.Context) sdk.Iterator
	SetBlockEvents(ctx sdk.Context, blockEvents *BlockEvents)
}

// GetBlockEvents read block events from KV store
func (k KVStore) GetBlockEvents(ctx sdk.Context, height int64) (*BlockEvents, error) {
	key := k.GetKey(ctx, prefixBlockEvents, strconv.FormatInt(height, 10))
	store := ctx.KVStore(k.storeKey)
	// doesn't exist
	if !store.Has([]byte(key)) {
		return nil, nil
	}
	buf := store.Get([]byte(key))
	var e BlockEvents
	if err := k.cdc.UnmarshalBinaryBare(buf, &e); err != nil {
		return nil, fmt.Errorf("fail to unmarshal block events: %w", err)
	}
	return &e, nil
}

// SetBlockEvents save the block events into KV store
func (k KVStore) SetBlockEvents(ctx sdk.Context, events *BlockEvents) {
	key := k.GetKey(ctx, prefixBlockEvents, strconv.FormatInt(events.Height, 10))
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(events))
}

// GetBlockEventsIterator
func (k KVStore) GetBlockEventsIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixBlockEvents))
}
