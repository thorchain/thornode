package thorchain

import (
	"errors"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperEvents interface {
	GetEvent(ctx sdk.Context, eventID int64) (Event, error)
	GetEventsIterator(ctx sdk.Context) sdk.Iterator
	GetNextEventID(ctx sdk.Context) (int64, error)
	UpsertEvent(ctx sdk.Context, event Event)
	GetPendingEventID(ctx sdk.Context, txID common.TxID) (int64, error)
	GetCurrentEventID(ctx sdk.Context) (int64, error)
	SetCurrentEventID(ctx sdk.Context, eventID int64)
	GetAllPendingEvnets(ctx sdk.Context) (Events, error)
}

var ErrEventNotFound = errors.New("event not found")

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
func (k KVStore) UpsertEvent(ctx sdk.Context, event Event) {
	key := k.GetKey(ctx, prefixEvents, strconv.FormatInt(event.ID, 10))
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&event))
	if event.Status == EventPending {
		k.setEventPending(ctx, event)
		return
	}
	k.removeEventPending(ctx, event)
}

func (k KVStore) removeEventPending(ctx sdk.Context, event Event) {
	key := k.GetKey(ctx, prefixPendingEvents, event.InTx.ID.String())
	store := ctx.KVStore(k.storeKey)
	store.Delete([]byte(key))
}

// setEventPending store the pending event use InTx hash as the key
func (k KVStore) setEventPending(ctx sdk.Context, event Event) {
	if event.Status != EventPending {
		return
	}
	ctx.Logger().Info(fmt.Sprintf("event id(%d): %s", event.ID, event.InTx.ID))
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
	currentEventID, err := k.GetCurrentEventID(ctx)
	if nil != err {
		return currentEventID, err
	}
	nextEventID = currentEventID + 1
	k.SetCurrentEventID(ctx, nextEventID)
	return currentEventID, nil
}

// GetCurrentEventID get the current event id in data store without increasing it
func (k KVStore) GetCurrentEventID(ctx sdk.Context) (int64, error) {
	var currentEventID int64
	key := k.GetKey(ctx, prefixCurrentEventID, "")
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return 0, nil
	}
	buf := store.Get([]byte(key))
	if err := k.cdc.UnmarshalBinaryBare(buf, &currentEventID); err != nil {
		return 0, dbError(ctx, "Unmarshal: current event id", err)
	}
	return currentEventID, nil

}

// SetCurrentEventID set the current event id in kv store
func (k KVStore) SetCurrentEventID(ctx sdk.Context, eventID int64) {
	key := k.GetKey(ctx, prefixCurrentEventID, "")
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&eventID))
}

// GetAllPendingEvents all events in pending status
func (k KVStore) GetAllPendingEvnets(ctx sdk.Context) (Events, error) {
	key := k.GetKey(ctx, prefixPendingEvents, "")
	store := ctx.KVStore(k.storeKey)
	var events Events
	iter := sdk.KVStorePrefixIterator(store, []byte(key))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var eventID int64
		if err := k.Cdc().UnmarshalBinaryBare(iter.Value(), &eventID); nil != err {
			return nil, fmt.Errorf("fail to unmarshal event id: %w", err)
		}
		event, err := k.GetEvent(ctx, eventID)
		if nil != err {
			return nil, fmt.Errorf("fail to get event: %w", err)
		}
		events = append(events, event)
	}
	return events, nil
}
