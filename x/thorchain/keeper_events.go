package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/common"
)

type KeeperEvents interface {
	GetIncompleteEvents(ctx sdk.Context) (Events, error)
	SetIncompleteEvents(ctx sdk.Context, events Events)
	AddIncompleteEvents(ctx sdk.Context, event Event)
	GetCompleteEventIterator(ctx sdk.Context) sdk.Iterator
	GetCompletedEvent(ctx sdk.Context, id int64) (Event, error)
	SetCompletedEvent(ctx sdk.Context, event Event)
	CompleteEvents(ctx sdk.Context, in []common.TxID, out common.Tx)
	GetLastEventID(ctx sdk.Context) int64
	SetLastEventID(ctx sdk.Context, id int64)
}

// GetIncompleteEvents retrieve incomplete events
func (k KVStore) GetIncompleteEvents(ctx sdk.Context) (Events, error) {
	key := k.GetKey(ctx, prefixInCompleteEvents, "")
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return Events{}, nil
	}
	buf := store.Get([]byte(key))
	var events Events
	if err := k.cdc.UnmarshalBinaryBare(buf, &events); nil != err {
		ctx.Logger().Error(fmt.Sprintf("fail to unmarshal incomplete events, err: %s", err))
		return Events{}, errors.Wrap(err, "fail to unmarshal incomplete events")
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
func (k KVStore) AddIncompleteEvents(ctx sdk.Context, event Event) {
	events, _ := k.GetIncompleteEvents(ctx)
	events = append(events, event)
	k.SetIncompleteEvents(ctx, events)
}

// GetCompleteEventIterator iterate complete events
func (k KVStore) GetCompleteEventIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixCompleteEvent))
}

// GetCompletedEvent retrieve completed event
func (k KVStore) GetCompletedEvent(ctx sdk.Context, id int64) (Event, error) {
	key := k.GetKey(ctx, prefixCompleteEvent, fmt.Sprintf("%d", id))
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return Event{}, nil
	}
	buf := store.Get([]byte(key))
	var event Event
	if err := k.cdc.UnmarshalBinaryBare(buf, &event); nil != err {
		ctx.Logger().Error(fmt.Sprintf("fail to unmarshal complete event, err: %s", err))
		return Event{}, errors.Wrap(err, "fail to unmarshal complete event")
	}
	return event, nil
}

// SetCompletedEvent write a completed event
func (k KVStore) SetCompletedEvent(ctx sdk.Context, event Event) {
	key := k.GetKey(ctx, prefixCompleteEvent, fmt.Sprintf("%d", event.ID))
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&event))
}

// CompleteEvent
func (k KVStore) CompleteEvents(ctx sdk.Context, in []common.TxID, out common.Tx) {
	lastEventID := k.GetLastEventID(ctx)

	incomplete, _ := k.GetIncompleteEvents(ctx)

	for _, txID := range in {
		var evts Events
		evts, incomplete = incomplete.PopByInHash(txID)
		for _, evt := range evts {
			if !evt.Empty() {
				voter := k.GetTxInVoter(ctx, txID)
				evt.OutTx = append(evt.OutTx, out)
				// Check if we've seen enough OutTx to the number expected to
				// have seen by the voter.
				// Sometimes we can have voter.NumOuts be zero, for example,
				// when someone is staking there are no out txs.
				// Due to Asgard being the backup, in case a Yggdrasil pool isn't
				// signing tx, don't count them as two out tx, merge them into one.
				var dups int
				for i, tx1 := range voter.OutTxs {
					for j, tx2 := range voter.OutTxs {
						if i == j {
							continue
						}
						if tx1.Coin.Equals(tx2.Coin) || tx1.Memo == tx2.Memo {
							dups += 1
						}
					}
				}
				// if there are any dups, they will be counted twice, so divide by two
				dups = dups / 2

				if len(evt.OutTx) >= (len(voter.OutTxs) - dups) {
					lastEventID += 1
					evt.ID = lastEventID
					k.SetCompletedEvent(ctx, evt)
				} else {
					// since we have more out event, add event back to
					// incomplete evts
					incomplete = append(incomplete, evt)
				}
			}
		}
	}

	// save new list of incomplete events
	k.SetIncompleteEvents(ctx, incomplete)

	k.SetLastEventID(ctx, lastEventID)
}

// GetLastEventID get last event id
func (k KVStore) GetLastEventID(ctx sdk.Context) int64 {
	var lastEventID int64
	key := k.GetKey(ctx, prefixLastEventID, "")
	store := ctx.KVStore(k.storeKey)
	if store.Has([]byte(key)) {
		buf := store.Get([]byte(key))
		_ = k.cdc.UnmarshalBinaryBare(buf, &lastEventID)
	}
	return lastEventID
}

// SetLastEventID write a last event id
func (k KVStore) SetLastEventID(ctx sdk.Context, id int64) {
	key := k.GetKey(ctx, prefixLastEventID, "")
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&id))
}
