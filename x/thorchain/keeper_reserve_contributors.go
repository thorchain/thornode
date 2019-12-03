package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KeeperReserveContributors interface {
	GetReservesContributors(ctx sdk.Context) ReserveContributors
	SetReserveContributors(ctx sdk.Context, contribs ReserveContributors)
	AddFeeToReserve(ctx sdk.Context, fee sdk.Uint) error
}

func (k KVStore) AddFeeToReserve(ctx sdk.Context, fee sdk.Uint) error {
	vault, err := k.GetVaultData(ctx)
	if nil != err {
		return fmt.Errorf("fail to get vault: %w", err)
	}
	vault.TotalReserve = vault.TotalReserve.Add(fee)
	return k.SetVaultData(ctx, vault)
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
