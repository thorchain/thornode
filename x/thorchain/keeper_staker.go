package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperStaker interface {
	GetStakerIterator(ctx sdk.Context, _ common.Asset) sdk.Iterator
	GetStaker(ctx sdk.Context, asset common.Asset, addr common.Address) (Staker, error)
	SetStaker(ctx sdk.Context, staker Staker)
	RemoveStaker(ctx sdk.Context, staker Staker)
}

// GetStakerIterator iterate stakers
func (k KVStore) GetStakerIterator(ctx sdk.Context, asset common.Asset) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixStaker, Staker{Asset: asset}.Key())
	return sdk.KVStorePrefixIterator(store, []byte(key))
}

// GetStaker retrieve staker from the data store
func (k KVStore) GetStaker(ctx sdk.Context, asset common.Asset, addr common.Address) (Staker, error) {
	store := ctx.KVStore(k.storeKey)
	staker := Staker{
		Asset:       asset,
		RuneAddress: addr,
		Units:       sdk.ZeroUint(),
		PendingRune: sdk.ZeroUint(),
	}
	key := k.GetKey(ctx, prefixStaker, staker.Key())
	if !store.Has([]byte(key)) {
		return staker, nil
	}
	buf := store.Get([]byte(key))
	if err := k.cdc.UnmarshalBinaryBare(buf, &staker); err != nil {
		return staker, err
	}
	return staker, nil
}

// SetStaker store the staker to kvstore
func (k KVStore) SetStaker(ctx sdk.Context, staker Staker) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixStaker, staker.Key())
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(staker))
}

// RemoveStaker remove the staker to kvstore
func (k KVStore) RemoveStaker(ctx sdk.Context, staker Staker) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixStaker, staker.Key())
	store.Delete([]byte(key))
}
