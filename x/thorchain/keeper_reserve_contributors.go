package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KeeperReserveContributors interface {
	GetReservesContributors(ctx sdk.Context) (ReserveContributors, error)
	SetReserveContributors(ctx sdk.Context, contributors ReserveContributors) error
	AddFeeToReserve(ctx sdk.Context, fee sdk.Uint) error
}

// AddFeeToReserve add fee to reserve
func (k KVStore) AddFeeToReserve(ctx sdk.Context, fee sdk.Uint) error {
	vault, err := k.GetVaultData(ctx)
	if nil != err {
		return fmt.Errorf("fail to get vault: %w", err)
	}
	vault.TotalReserve = vault.TotalReserve.Add(fee)
	return k.SetVaultData(ctx, vault)
}

func (k KVStore) GetReservesContributors(ctx sdk.Context) (ReserveContributors, error) {
	contributors := make(ReserveContributors, 0)
	key := k.GetKey(ctx, prefixReserves, "")
	store := ctx.KVStore(k.storeKey)
	if store.Has([]byte(key)) {
		buf := store.Get([]byte(key))
		if err := k.cdc.UnmarshalBinaryBare(buf, &contributors); nil != err {
			return nil, dbError(ctx, "fail to unmarshal reserve contributors", err)
		}
	}
	return contributors, nil
}

func (k KVStore) SetReserveContributors(ctx sdk.Context, contributors ReserveContributors) error {
	key := k.GetKey(ctx, prefixReserves, "")
	store := ctx.KVStore(k.storeKey)
	buf, err := k.cdc.MarshalBinaryBare(contributors)
	if nil != err {
		return dbError(ctx, "fail to marshal reserve contributors to binary", err)
	}
	store.Set([]byte(key), buf)
	return nil
}
