package thorchain

import (
	"errors"
	stdErrors "errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperVault interface {
	GetVaultIterator(ctx sdk.Context) sdk.Iterator
	VaultExists(ctx sdk.Context, pk common.PubKey) bool
	SetVault(ctx sdk.Context, vault Vault) error
	GetVault(ctx sdk.Context, pk common.PubKey) (Vault, error)
	HasValidVaultPools(ctx sdk.Context) (bool, error)
	GetActiveAsgardVaults(ctx sdk.Context) (Vaults, error)
	DeleteVault(ctx sdk.Context, pk common.PubKey) error
}

// GetVaultIterator only iterate vault pools
func (k KVStore) GetVaultIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixVaultPool))
}

// SetVault save the Vault object to store
func (k KVStore) SetVault(ctx sdk.Context, vault Vault) error {
	key := k.GetKey(ctx, prefixVaultPool, vault.PubKey.String())
	store := ctx.KVStore(k.storeKey)
	buf, err := k.cdc.MarshalBinaryBare(vault)
	if nil != err {
		return dbError(ctx, "fail to marshal vault to binary", err)
	}
	if vault.IsAsgard() {
		if err := k.addAsgardIndex(ctx, vault.PubKey); err != nil {
			return err
		}
	}
	store.Set([]byte(key), buf)
	return nil
}

// VaultExists check whether the given pubkey is associated with a vault vault
func (k KVStore) VaultExists(ctx sdk.Context, pk common.PubKey) bool {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixVaultPool, pk.String())
	return store.Has([]byte(key))
}

var ErrVaultNotFound = errors.New("vault not found")

// GetVault get Vault with the given pubkey from data store
func (k KVStore) GetVault(ctx sdk.Context, pk common.PubKey) (Vault, error) {
	var vault Vault
	key := k.GetKey(ctx, prefixVaultPool, pk.String())
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		vault.BlockHeight = ctx.BlockHeight()
		vault.PubKey = pk
		return vault, fmt.Errorf("vault with pubkey(%s) doesn't exist: %w", pk, ErrVaultNotFound)
	}
	buf := store.Get([]byte(key))
	if err := k.cdc.UnmarshalBinaryBare(buf, &vault); nil != err {
		return vault, dbError(ctx, "fail to unmarshal vault", err)
	}
	if vault.PubKey.IsEmpty() {
		vault.PubKey = pk
	}
	return vault, nil
}

// HasValidVaultPools check the datastore to see whether we have a valid vault pool
func (k KVStore) HasValidVaultPools(ctx sdk.Context) (bool, error) {
	iterator := k.GetVaultIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var vault Vault
		if err := k.cdc.UnmarshalBinaryBare(iterator.Value(), &vault); nil != err {
			return false, dbError(ctx, "fail to unmarshal vault", err)
		}
		if vault.HasFunds() {
			return true, nil
		}
	}
	return false, nil
}

func (k KVStore) getAsgardIndex(ctx sdk.Context) ([]common.PubKey, error) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixVaultAsgardIndex, "")
	if !store.Has([]byte(key)) {
		return nil, nil
	}
	buf := store.Get([]byte(key))
	var pks []common.PubKey
	if err := k.cdc.UnmarshalBinaryBare(buf, &pks); nil != err {
		return nil, dbError(ctx, "fail to unmarshal asgard index", err)
	}
	return pks, nil
}

func (k KVStore) addAsgardIndex(ctx sdk.Context, pubkey common.PubKey) error {
	pks, err := k.getAsgardIndex(ctx)
	if err != nil {
		return err
	}
	for _, pk := range pks {
		if pk.Equals(pubkey) {
			return nil
		}
	}
	pks = append(pks, pubkey)
	key := k.GetKey(ctx, prefixVaultAsgardIndex, "")
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(pks))
	return nil
}

func (k KVStore) GetActiveAsgardVaults(ctx sdk.Context) (Vaults, error) {
	pks, err := k.getAsgardIndex(ctx)
	if err != nil {
		return nil, err
	}

	var asgards Vaults
	for _, pk := range pks {
		vault, err := k.GetVault(ctx, pk)
		if err != nil {
			return nil, err
		}
		if vault.IsAsgard() && vault.Status == ActiveVault {
			asgards = append(asgards, vault)
		}
	}

	return asgards, nil
}

func (k KVStore) DeleteVault(ctx sdk.Context, pubkey common.PubKey) error {
	vault, err := k.GetVault(ctx, pubkey)
	if err != nil {
		if !stdErrors.Is(err, ErrVaultNotFound) {
			return nil
		}
		return err
	}

	if vault.IsAsgard() {
		pks, err := k.getAsgardIndex(ctx)
		if err != nil {
			return err
		}

		var newPks []common.PubKey
		for _, pk := range pks {
			if !pk.Equals(pubkey) {
				newPks = append(newPks, pk)
			}
		}

		key := k.GetKey(ctx, prefixVaultAsgardIndex, "")
		store := ctx.KVStore(k.storeKey)
		store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(newPks))
	}
	return nil
}
