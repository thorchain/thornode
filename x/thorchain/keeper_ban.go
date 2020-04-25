package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KeeperBanVoter interface {
	SetBanVoter(_ sdk.Context, _ BanVoter)
	GetBanVoter(_ sdk.Context, _ sdk.AccAddress) (BanVoter, error)
}

// SetBanVoter - save a ban voter object
func (k KVStore) SetBanVoter(ctx sdk.Context, ban BanVoter) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixBanVoter, ban.String())
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(ban))
}

// GetBanVoter - gets information of a tx hash
func (k KVStore) GetBanVoter(ctx sdk.Context, addr sdk.AccAddress) (BanVoter, error) {
	ban := NewBanVoter(addr)
	key := k.GetKey(ctx, prefixBanVoter, ban.String())

	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return ban, nil
	}

	bz := store.Get([]byte(key))
	var record BanVoter
	if err := k.cdc.UnmarshalBinaryBare(bz, &record); err != nil {
		return ban, dbError(ctx, "Unmarshal: ban voter", err)
	}
	return record, nil
}
