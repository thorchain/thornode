package thorchain

import sdk "github.com/cosmos/cosmos-sdk/types"

type KeeperTssKeysignFail interface {
	SetTssKeysignFailVoter(_ sdk.Context, tss TssKeysignFailVoter)
	GetTssKeysignFailVoterIterator(_ sdk.Context) sdk.Iterator
	GetTssKeysignFailVoter(_ sdk.Context, _ string) (TssKeysignFailVoter, error)
}

// SetTssKeysignFailVoter - save a txin voter object
func (k KVStore) SetTssKeysignFailVoter(ctx sdk.Context, tss TssKeysignFailVoter) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixTss, tss.String())
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(tss))
}

// GetTssKeysignFailVoterIterator iterate tx in voters
func (k KVStore) GetTssKeysignFailVoterIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixTss))
}

// GetTss - gets information of a tx hash
func (k KVStore) GetTssKeysignFailVoter(ctx sdk.Context, id string) (TssKeysignFailVoter, error) {
	key := k.GetKey(ctx, prefixTss, id)

	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return TssKeysignFailVoter{ID: id}, nil
	}

	bz := store.Get([]byte(key))
	var record TssKeysignFailVoter
	if err := k.cdc.UnmarshalBinaryBare(bz, &record); err != nil {
		return TssKeysignFailVoter{}, dbError(ctx, "Unmarshal: tss voter", err)
	}
	return record, nil
}
