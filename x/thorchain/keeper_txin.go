package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperObservedTx interface {
	SetObservedTxVoter(ctx sdk.Context, tx ObservedTxVoter)
	GetObservedTxVoterIterator(ctx sdk.Context) sdk.Iterator
	GetObservedTxVoter(ctx sdk.Context, hash common.TxID) (ObservedTxVoter, error)
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
func (k KVStore) GetObservedTxVoter(ctx sdk.Context, hash common.TxID) (ObservedTxVoter, error) {
	key := k.GetKey(ctx, prefixObservedTx, hash.String())

	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return ObservedTxVoter{TxID: hash}, nil
	}

	bz := store.Get([]byte(key))
	var record ObservedTxVoter
	if err := k.cdc.UnmarshalBinaryBare(bz, &record); err != nil {
		return ObservedTxVoter{}, dbError(ctx, "Unmarshal: observed tx voter", err)
	}
	return record, nil
}
