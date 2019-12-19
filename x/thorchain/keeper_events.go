package thorchain

import (
	"errors"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperEvents interface {
	GetEvents(ctx sdk.Context) (Events, error)
	SetEvents(ctx sdk.Context, events Events)
	GetEventsIterator(ctx sdk.Context) sdk.Iterator
	GetNextEventID(ctx sdk.Context) (int64, error)
	AddEvent(ctx sdk.Context, event Event)
}

var ErrEventNotFound = errors.New("event not found")

// GetEvents retrieve  events
func (k KVStore) GetEvents(ctx sdk.Context) (Events, error) {
	events := make(Events, 0)
	key := k.GetKey(ctx, prefixEvents, "")
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

// SetEvents write  events
func (k KVStore) SetEvents(ctx sdk.Context, events Events) {
	key := k.GetKey(ctx, prefixEvents, "")
	store := ctx.KVStore(k.storeKey)
	if len(events) == 0 {
		return
	} else {
		store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&events))
	}
}

// AddEvents will add the given events to data store
func (k KVStore) AddEvents(ctx sdk.Context, events Events) {
	for _, item := range events {
		k.AddEvent(ctx, item)
	}
}

// GetEventByID will retrieve event with the given id from data store
func (k KVStore) GetEvent(ctx sdk.Context, eventID int64) (Event, error) {
	key := k.GetKey(ctx, prefixEvents, strconv.FormatInt(eventID, 10))
	store := ctx.KVStore(k.storeKey)
	buf := store.Get([]byte(key))
	var e Event
	if err := k.Cdc().UnmarshalBinaryBare(buf, &e); nil != err {
		return Event{}, fmt.Errorf("fail to unmarshal event: %w", err)
	}
	return e, nil
}

// AddEvent add one event to data store
func (k KVStore) AddEvent(ctx sdk.Context, event Event) {
	key := k.GetKey(ctx, prefixEvents, strconv.FormatInt(event.ID, 10))
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&event))
	if event.Status == EventPending {
		k.SetEventPending(ctx, event)
	}
}

// SetEventPending store the pending event use InTx hash as the key
func (k KVStore) SetEventPending(ctx sdk.Context, event Event) {
	if event.Status != EventPending {
		return
	}
	key := k.GetKey(ctx, prefixPendingEvents, event.InTx.ID.String())
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&event.ID))
}

// GetPendingEventID we store the event in pending status using it's in tx hash
func (k KVStore) GetPendingEventID(ctx sdk.Context, txID common.TxID) (int64, error) {
	key := k.GetKey(ctx, prefixPendingEvents, txID.String())
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return 0, ErrEventNotFound
	}
	buf := store.Get([]byte(key))
	var eventID int64
	if err := k.Cdc().UnmarshalBinaryBare(buf, &eventID); nil != err {
		return 0, fmt.Errorf("fail to unmarshal event id: %w", err)
	}
	return eventID, nil
}

// GetCompleteEventIterator iterate complete events
func (k KVStore) GetEventsIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixEvents))
}

// GetNextEventID will increase the event id in key value store
func (k KVStore) GetNextEventID(ctx sdk.Context) (int64, error) {
	var currentEventID, nextEventID int64
	key := k.GetKey(ctx, prefixLastEventID, "")
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		currentEventID = 0
		nextEventID = currentEventID + 1
		store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&nextEventID))
		return currentEventID, nil
	}
	buf := store.Get([]byte(key))
	if err := k.cdc.UnmarshalBinaryBare(buf, &currentEventID); err != nil {
		return 0, dbError(ctx, "Unmarshal: last event id", err)
	}
	nextEventID = currentEventID + 1
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&nextEventID))
	return currentEventID, nil
}
