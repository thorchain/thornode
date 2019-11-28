package thorchain

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/common"
)

type KeeperTxIn interface {
	SetTxInVoter(ctx sdk.Context, tx TxInVoter)
	GetTxInVoterIterator(ctx sdk.Context) sdk.Iterator
	GetTxInVoter(ctx sdk.Context, hash common.TxID) TxInVoter
	CheckTxHash(ctx sdk.Context, hash common.TxID) bool
	GetTxInIndexIterator(ctx sdk.Context) sdk.Iterator
	GetTxInIndex(ctx sdk.Context, height uint64) (TxInIndex, error)
	SetTxInIndex(ctx sdk.Context, height uint64, index TxInIndex)
	AddToTxInIndex(ctx sdk.Context, height uint64, id common.TxID) error
}

// SetTxInVoter - save a txin voter object
func (k KVStore) SetTxInVoter(ctx sdk.Context, tx TxInVoter) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixTxIn, tx.String())
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(tx))
}

// GetTxInVoterIterator iterate tx in voters
func (k KVStore) GetTxInVoterIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixTxIn))
}

// GetTxIn - gets information of a tx hash
func (k KVStore) GetTxInVoter(ctx sdk.Context, hash common.TxID) TxInVoter {
	key := k.GetKey(ctx, prefixTxIn, hash.String())

	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return TxInVoter{TxID: hash}
	}

	bz := store.Get([]byte(key))
	var record TxInVoter
	k.cdc.MustUnmarshalBinaryBare(bz, &record)
	return record
}

// CheckTxHash - check to see if we have already processed a specific tx
func (k KVStore) CheckTxHash(ctx sdk.Context, hash common.TxID) bool {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixTxIn, hash.String())
	return store.Has([]byte(key))
}

// GetTxInIndexIterator iterate tx in indexes
func (k KVStore) GetTxInIndexIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixTxInIndex))
}

// GetTxInIndex retrieve txIn by height
func (k KVStore) GetTxInIndex(ctx sdk.Context, height uint64) (TxInIndex, error) {
	key := k.GetKey(ctx, prefixTxInIndex, strconv.FormatUint(height, 10))
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return TxInIndex{}, nil
	}
	buf := store.Get([]byte(key))
	var index TxInIndex
	if err := k.cdc.UnmarshalBinaryBare(buf, &index); nil != err {
		ctx.Logger().Error(fmt.Sprintf("fail to unmarshal poolindex,err: %s", err))
		return TxInIndex{}, errors.Wrap(err, "fail to unmarshal poolindex")
	}
	return index, nil
}

// SetTxInIndex write a TxIn index into datastore
func (k KVStore) SetTxInIndex(ctx sdk.Context, height uint64, index TxInIndex) {
	key := k.GetKey(ctx, prefixTxInIndex, strconv.FormatUint(height, 10))
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&index))
}

// AddToTxInIndex will add the given txIn into the index
func (k KVStore) AddToTxInIndex(ctx sdk.Context, height uint64, id common.TxID) error {
	index, err := k.GetTxInIndex(ctx, height)
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
	k.SetTxInIndex(ctx, height, index)
	return nil
}
