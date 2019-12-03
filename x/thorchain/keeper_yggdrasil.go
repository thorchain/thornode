package thorchain

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperYggdrasil interface {
	GetYggdrasilIterator(ctx sdk.Context) sdk.Iterator
	YggdrasilExists(ctx sdk.Context, pk common.PubKey) bool
	FindPubKeyOfAddress(ctx sdk.Context, addr common.Address, chain common.Chain) (common.PubKey, error)
	SetYggdrasil(ctx sdk.Context, ygg Yggdrasil) error
	GetYggdrasil(ctx sdk.Context, pk common.PubKey) (Yggdrasil, error)
	HasValidYggdrasilPools(ctx sdk.Context) (bool, error)
}

// GetYggdrasilIterator only iterate yggdrasil pools
func (k KVStore) GetYggdrasilIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixYggdrasilPool))
}

// FindPubKeyOfAddress given an address to find out it's relevant pubkey
func (k KVStore) FindPubKeyOfAddress(ctx sdk.Context, addr common.Address, chain common.Chain) (common.PubKey, error) {
	iterator := k.GetYggdrasilIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var ygg Yggdrasil
		if err := k.cdc.UnmarshalBinaryBare(iterator.Value(), &ygg); nil != err {
			return common.EmptyPubKey, dbError(ctx, "fail to unmarshal yggdrasil", err)
		}
		address, err := ygg.PubKey.GetAddress(chain)
		if err != nil {
			return common.EmptyPubKey, err
		}
		if !address.IsEmpty() && address.Equals(addr) {
			return ygg.PubKey, nil
		}
	}
	return common.EmptyPubKey, nil
}

// SetYggdrasil save the Yggdrasil object to store
func (k KVStore) SetYggdrasil(ctx sdk.Context, ygg Yggdrasil) error {
	key := k.GetKey(ctx, prefixYggdrasilPool, ygg.PubKey.String())
	store := ctx.KVStore(k.storeKey)
	buf, err := k.cdc.MarshalBinaryBare(ygg)
	if nil != err {
		return dbError(ctx, "fail to marshal yggdrasil to binary", err)
	}
	store.Set([]byte(key), buf)
	return nil
}

// YggdrasilExists check whether the given pubkey is associated with a yggdrasil vault
func (k KVStore) YggdrasilExists(ctx sdk.Context, pk common.PubKey) bool {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixYggdrasilPool, pk.String())
	return store.Has([]byte(key))
}

var ErrYggdrasilNotFound = errors.New("yggdrasil not found")

// GetYggdrasil get Yggdrasil with the given pubkey from data store
func (k KVStore) GetYggdrasil(ctx sdk.Context, pk common.PubKey) (Yggdrasil, error) {
	var ygg Yggdrasil
	key := k.GetKey(ctx, prefixYggdrasilPool, pk.String())
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return ygg, fmt.Errorf("yggdrasil with pubkey(%s) doesn't exist: %w", pk, ErrYggdrasilNotFound)
	}
	buf := store.Get([]byte(key))
	if err := k.cdc.UnmarshalBinaryBare(buf, &ygg); nil != err {
		return ygg, dbError(ctx, "fail to unmarshal yggdrasil", err)
	}
	if ygg.PubKey.IsEmpty() {
		ygg.PubKey = pk
	}
	return ygg, nil
}

// HasValidYggdrasilPools check the datastore to see whether we have a valid yggdrasil pool
func (k KVStore) HasValidYggdrasilPools(ctx sdk.Context) (bool, error) {
	iterator := k.GetYggdrasilIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var ygg Yggdrasil
		if err := k.cdc.UnmarshalBinaryBare(iterator.Value(), &ygg); nil != err {
			return false, dbError(ctx, "fail to unmarshal yggdrasil", err)
		}
		if ygg.HasFunds() {
			return true, nil
		}
	}
	return false, nil
}
