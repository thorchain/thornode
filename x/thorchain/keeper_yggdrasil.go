package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
)

type KeeperYggdrasil interface {
	GetYggdrasilIterator(ctx sdk.Context) sdk.Iterator
	YggdrasilExists(ctx sdk.Context, pk common.PubKey) bool
	FindPubKeyOfAddress(ctx sdk.Context, addr common.Address, chain common.Chain) (common.PubKey, error)
	SetYggdrasil(ctx sdk.Context, ygg Yggdrasil)
	GetYggdrasil(ctx sdk.Context, pk common.PubKey) Yggdrasil
	HasValidYggdrasilPools(ctx sdk.Context) (bool, error)
}

// GetYggdrasilIterator only iterate yggdrasil pools
func (k KVStore) GetYggdrasilIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixYggdrasilPool))
}

func (k KVStore) FindPubKeyOfAddress(ctx sdk.Context, addr common.Address, chain common.Chain) (common.PubKey, error) {
	iterator := k.GetYggdrasilIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var ygg Yggdrasil
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &ygg)
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

func (k KVStore) SetYggdrasil(ctx sdk.Context, ygg Yggdrasil) {
	key := k.GetKey(ctx, prefixYggdrasilPool, ygg.PubKey.String())
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(ygg))
}

// YggdrasilExists check whether the given pubkey is associated with a
// yggdrasil vault
func (k KVStore) YggdrasilExists(ctx sdk.Context, pk common.PubKey) bool {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixYggdrasilPool, pk.String())
	return store.Has([]byte(key))
}

func (k KVStore) GetYggdrasil(ctx sdk.Context, pk common.PubKey) Yggdrasil {
	var ygg Yggdrasil
	key := k.GetKey(ctx, prefixYggdrasilPool, pk.String())
	store := ctx.KVStore(k.storeKey)
	if store.Has([]byte(key)) {
		buf := store.Get([]byte(key))
		_ = k.cdc.UnmarshalBinaryBare(buf, &ygg)
	}
	if ygg.PubKey.IsEmpty() {
		ygg.PubKey = pk
	}
	return ygg
}
func (k KVStore) HasValidYggdrasilPools(ctx sdk.Context) (bool, error) {
	iterator := k.GetYggdrasilIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var ygg Yggdrasil
		if err := k.cdc.UnmarshalBinaryBare(iterator.Value(), &ygg); nil != err {
			return false, fmt.Errorf("fail to unmarshal yggdrasil: %w", err)
		}
		if ygg.HasFunds() {
			return true, nil
		}
	}
	return false, nil
}
