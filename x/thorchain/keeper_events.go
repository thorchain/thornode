package thorchain

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KeeperEvents interface {
	GetIncompleteEvents(ctx sdk.Context) (Events, error)
	SetIncompleteEvents(ctx sdk.Context, events Events)
	AddIncompleteEvents(ctx sdk.Context, event Event) error
	GetCompleteEventIterator(ctx sdk.Context) sdk.Iterator
	GetCompletedEvent(ctx sdk.Context, id int64) (Event, error)
	SetCompletedEvent(ctx sdk.Context, event Event)
	GetLastEventID(ctx sdk.Context) (int64, error)
	SetLastEventID(ctx sdk.Context, id int64)
}

// GetIncompleteEvents retrieve incomplete events
func (k KVStore) GetIncompleteEvents(ctx sdk.Context) (Events, error) {
	events := make(Events, 0)
	key := k.GetKey(ctx, prefixInCompleteEvents, "")
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return events, nil
	}
	buf := store.Get([]byte(key))
	if err := k.cdc.UnmarshalBinaryBare(buf, &events); nil != err {
		return events, dbError(ctx, "Unmarshal: incomplete events", err)
	}
	return events, nil
}

// SetIncompleteEvents write incomplete events
func (k KVStore) SetIncompleteEvents(ctx sdk.Context, events Events) {
	key := k.GetKey(ctx, prefixInCompleteEvents, "")
	store := ctx.KVStore(k.storeKey)
	if len(events) == 0 {
		store.Delete([]byte(key))
	} else {
		store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&events))
	}
}

// AddIncompleteEvents append to incomplete events
func (k KVStore) AddIncompleteEvents(ctx sdk.Context, event Event) error {
	events, err := k.GetIncompleteEvents(ctx)
	if err != nil {
		return err
	}
	events = append(events, event)
	k.SetIncompleteEvents(ctx, events)
	return nil
}

// GetCompleteEventIterator iterate complete events
func (k KVStore) GetCompleteEventIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixCompleteEvent))
}

// GetCompletedEvent retrieve completed event
func (k KVStore) GetCompletedEvent(ctx sdk.Context, id int64) (Event, error) {
	key := k.GetKey(ctx, prefixCompleteEvent, strconv.FormatInt(id, 10))
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return Event{}, nil
	}
	buf := store.Get([]byte(key))
	var event Event
	if err := k.cdc.UnmarshalBinaryBare(buf, &event); nil != err {
		return event, dbError(ctx, "Unmarshal: complete events", err)
	}
	return event, nil
}

// SetCompletedEvent write a completed event
func (k KVStore) SetCompletedEvent(ctx sdk.Context, event Event) {
	key := k.GetKey(ctx, prefixCompleteEvent, strconv.FormatInt(event.ID, 10))
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&event))
}

// GetLastEventID get last event id
func (k KVStore) GetLastEventID(ctx sdk.Context) (int64, error) {
	var lastEventID int64
	key := k.GetKey(ctx, prefixLastEventID, "")
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return lastEventID, nil
	}
	buf := store.Get([]byte(key))
	if err := k.cdc.UnmarshalBinaryBare(buf, &lastEventID); err != nil {
		return lastEventID, dbError(ctx, "Unmarshal: last event id", err)
	}
	return lastEventID, nil
}

// SetLastEventID write a last event id
func (k KVStore) SetLastEventID(ctx sdk.Context, id int64) {
	key := k.GetKey(ctx, prefixLastEventID, "")
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&id))
}
