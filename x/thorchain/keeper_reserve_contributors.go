package thorchain

import sdk "github.com/cosmos/cosmos-sdk/types"

type KeeperReserveContributors interface {
	GetReservesContributors(ctx sdk.Context) ReserveContributors
	SetReserveContributors(ctx sdk.Context, contribs ReserveContributors)
}

func (k KVStore) GetReservesContributors(ctx sdk.Context) ReserveContributors {
	contribs := make(ReserveContributors, 0)
	key := k.GetKey(ctx, prefixReserves, "")
	store := ctx.KVStore(k.storeKey)
	if store.Has([]byte(key)) {
		buf := store.Get([]byte(key))
		_ = k.cdc.UnmarshalBinaryBare(buf, &contribs)
	}
	return contribs
}

func (k KVStore) SetReserveContributors(ctx sdk.Context, contribs ReserveContributors) {
	key := k.GetKey(ctx, prefixReserves, "")
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(contribs))
}
