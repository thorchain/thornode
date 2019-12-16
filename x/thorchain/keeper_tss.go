package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KeeperTss interface {
	SetTssVoter(_ sdk.Context, tss TssVoter)
	GetTssVoterIterator(_ sdk.Context) sdk.Iterator
	GetTssVoter(_ sdk.Context, _ string) (TssVoter, error)
}

// SetTssVoter - save a txin voter object
func (k KVStore) SetTssVoter(ctx sdk.Context, tss TssVoter) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixTss, tss.String())
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(tss))
}

// GetTssVoterIterator iterate tx in voters
func (k KVStore) GetTssVoterIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixTss))
}

// GetTss - gets information of a tx hash
func (k KVStore) GetTssVoter(ctx sdk.Context, id string) (TssVoter, error) {
	key := k.GetKey(ctx, prefixTss, id)

	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return TssVoter{ID: id}, nil
	}

	bz := store.Get([]byte(key))
	var record TssVoter
	if err := k.cdc.UnmarshalBinaryBare(bz, &record); err != nil {
		return TssVoter{}, dbError(ctx, "Unmarshal: tss voter", err)
	}
	return record, nil
}
