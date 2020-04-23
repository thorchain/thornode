package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperGas interface {
	GetGas(_ sdk.Context, asset common.Asset) ([]sdk.Uint, error)
	SetGas(_ sdk.Context, asset common.Asset, units []sdk.Uint)
	GetGasIterator(ctx sdk.Context) sdk.Iterator
}

func (k KVStore) GetGas(ctx sdk.Context, asset common.Asset) ([]sdk.Uint, error) {
	key := k.GetKey(ctx, prefixGas, asset.String())
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return nil, nil
	}
	var gas []sdk.Uint
	buf := store.Get([]byte(key))
	err := k.cdc.UnmarshalBinaryBare(buf, &gas)
	if err != nil {
		return nil, dbError(ctx, "Unmarshal: gas", err)
	}
	return gas, nil
}

func (k KVStore) SetGas(ctx sdk.Context, asset common.Asset, units []sdk.Uint) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixGas, asset.String())
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(units))
}

// GetGasIterator iterate gas units
func (k KVStore) GetGasIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixGas))
}
