package thorchain

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/common"
)

type KeeperObservedTx interface {
	SetObservedTxVoter(ctx sdk.Context, tx ObservedTxVoter)
	GetObservedTxVoterIterator(ctx sdk.Context) sdk.Iterator
	GetObservedTxVoter(ctx sdk.Context, hash common.TxID) ObservedTxVoter
	GetObservedTxIndexIterator(ctx sdk.Context) sdk.Iterator
	GetObservedTxIndex(ctx sdk.Context, height uint64) (ObservedTxIndex, error)
	SetObservedTxIndex(ctx sdk.Context, height uint64, index ObservedTxIndex)
	AddToObservedTxIndex(ctx sdk.Context, height uint64, id common.TxID) error
}

// SetObservedTxVoter - save a txin voter object
func (k KVStore) SetObservedTxVoter(ctx sdk.Context, tx ObservedTxVoter) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixObservedTx, tx.String())
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(tx))
}

// GetObservedTxVoterIterator iterate tx in voters
func (k KVStore) GetObservedTxVoterIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixObservedTx))
}

// GetObservedTx - gets information of a tx hash
func (k KVStore) GetObservedTxVoter(ctx sdk.Context, hash common.TxID) ObservedTxVoter {
	key := k.GetKey(ctx, prefixObservedTx, hash.String())

	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return ObservedTxVoter{TxID: hash}
	}

	bz := store.Get([]byte(key))
	var record ObservedTxVoter
	k.cdc.MustUnmarshalBinaryBare(bz, &record)
	return record
}

// GetObservedTxIndexIterator iterate tx in indexes
func (k KVStore) GetObservedTxIndexIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixObservedTxIndex))
}

// GetObservedTxIndex retrieve txIn by height
func (k KVStore) GetObservedTxIndex(ctx sdk.Context, height uint64) (ObservedTxIndex, error) {
	key := k.GetKey(ctx, prefixObservedTxIndex, strconv.FormatUint(height, 10))
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return ObservedTxIndex{}, nil
	}
	buf := store.Get([]byte(key))
	var index ObservedTxIndex
	if err := k.cdc.UnmarshalBinaryBare(buf, &index); nil != err {
		ctx.Logger().Error(fmt.Sprintf("fail to unmarshal observed tx index,err: %s", err))
		return ObservedTxIndex{}, errors.Wrap(err, "fail to unmarshal observed tx index")
	}
	return index, nil
}

// SetObservedTxIndex write a ObservedTx index into datastore
func (k KVStore) SetObservedTxIndex(ctx sdk.Context, height uint64, index ObservedTxIndex) {
	key := k.GetKey(ctx, prefixObservedTxIndex, strconv.FormatUint(height, 10))
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&index))
}

// AddToObservedTxIndex will add the given txIn into the index
func (k KVStore) AddToObservedTxIndex(ctx sdk.Context, height uint64, id common.TxID) error {
	index, err := k.GetObservedTxIndex(ctx, height)
	if nil != err {
		return err
	}
	for _, item := range index {
		if item.Equals(id) {
			// already in the index , don't need to add
			return nil
		}
	}
	index = append(index, id)
	k.SetObservedTxIndex(ctx, height, index)
	return nil
}
