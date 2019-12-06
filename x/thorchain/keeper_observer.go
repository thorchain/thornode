package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KeeperObserver interface {
	SetActiveObserver(ctx sdk.Context, addr sdk.AccAddress)
	RemoveActiveObserver(ctx sdk.Context, addr sdk.AccAddress)
	IsActiveObserver(ctx sdk.Context, addr sdk.AccAddress) bool
	GetObservingAddresses(ctx sdk.Context) ([]sdk.AccAddress, error)
	AddObservingAddresses(ctx sdk.Context, inAddresses []sdk.AccAddress) error
	ClearObservingAddresses(ctx sdk.Context)
}

// SetActiveObserver set the given addr as an active observer address
func (k KVStore) SetActiveObserver(ctx sdk.Context, addr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixActiveObserver, addr.String())
	ctx.Logger().Info("set_active_observer", "key", key)
	store.Set([]byte(key), addr.Bytes())
}

// RemoveActiveObserver remove the given address from active observer
func (k KVStore) RemoveActiveObserver(ctx sdk.Context, addr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixActiveObserver, addr.String())
	store.Delete([]byte(key))
}

// IsActiveObserver check the given account address, whether they are active
func (k KVStore) IsActiveObserver(ctx sdk.Context, addr sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixActiveObserver, addr.String())
	ctx.Logger().Info("is_active_observer", "key", key)
	return store.Has([]byte(key))
}

// GetObservingAddresses - get list of observed addresses. This is a list of
// addresses that have recently contributed via observing a tx that got 2/3rds
// majority
func (k KVStore) GetObservingAddresses(ctx sdk.Context) ([]sdk.AccAddress, error) {
	key := k.GetKey(ctx, prefixObservingAddresses, "")

	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return make([]sdk.AccAddress, 0), nil
	}

	bz := store.Get([]byte(key))
	var addresses []sdk.AccAddress
	if err := k.cdc.UnmarshalBinaryBare(bz, &addresses); err != nil {
		return nil, dbError(ctx, "Unmarshal: observer", err)
	}
	return addresses, nil
}

// AddObservingAddresses - add a list of addresses that have been helpful in
// getting enough observations to process an inbound tx.
func (k KVStore) AddObservingAddresses(ctx sdk.Context, inAddresses []sdk.AccAddress) error {
	// combine addresses
	curr, err := k.GetObservingAddresses(ctx)
	if err != nil {
		return err
	}
	all := append(curr, inAddresses...)

	// ensure uniqueness
	uniq := make([]sdk.AccAddress, 0, len(all))
	m := make(map[string]bool)
	for _, val := range all {
		if _, ok := m[val.String()]; !ok {
			m[val.String()] = true
			uniq = append(uniq, val)
		}
	}

	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixObservingAddresses, "")
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(uniq))
	return nil
}

// ClearObservingAddresses - clear all observing addresses
func (k KVStore) ClearObservingAddresses(ctx sdk.Context) {
	key := k.GetKey(ctx, prefixObservingAddresses, "")
	store := ctx.KVStore(k.storeKey)
	store.Delete([]byte(key))
}
