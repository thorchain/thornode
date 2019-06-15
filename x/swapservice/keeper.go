package swapservice

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/bank"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper maintains the link to data storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	coinKeeper bank.Keeper

	storeKey sdk.StoreKey // Unexposed key to access store from sdk.Context

	cdc *codec.Codec // The wire codec for binary encoding/decoding.
}

// NewKeeper creates new instances of the swapservice Keeper
func NewKeeper(coinKeeper bank.Keeper, storeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return Keeper{
		coinKeeper: coinKeeper,
		storeKey:   storeKey,
		cdc:        cdc,
	}
}

// Gets the entire PoolStruct metadata struct for a pooldata
func (k Keeper) GetPoolStruct(ctx sdk.Context, pooldata string) PoolStruct {
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(pooldata)) {
		return NewPoolStruct()
	}
	bz := store.Get([]byte(pooldata))
	var poolstruct PoolStruct
	k.cdc.MustUnmarshalBinaryBare(bz, &poolstruct)
	return poolstruct
}

// Sets the entire PoolStruct metadata struct for a pooldata
func (k Keeper) SetPoolStruct(ctx sdk.Context, pooldata string, poolstruct PoolStruct) {
	if poolstruct.Owner.Empty() {
		return
	}
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(pooldata), k.cdc.MustMarshalBinaryBare(poolstruct))
}

// ResolvePoolData - returns the string that the pooldata resolves to
func (k Keeper) ResolvePoolData(ctx sdk.Context, pooldata string) string {
	return k.GetPoolStruct(ctx, pooldata).Value
}

// SetPoolData - sets the value string that a pooldata resolves to
func (k Keeper) SetPoolData(ctx sdk.Context, pooldata string, value string) {
	poolstruct := k.GetPoolStruct(ctx, pooldata)
	poolstruct.Value = value
	k.SetPoolStruct(ctx, pooldata, poolstruct)
}

// HasOwner - returns whether or not the pooldata already has an owner
func (k Keeper) HasOwner(ctx sdk.Context, pooldata string) bool {
	return !k.GetPoolStruct(ctx, pooldata).Owner.Empty()
}

// GetOwner - get the current owner of a pooldata
func (k Keeper) GetOwner(ctx sdk.Context, pooldata string) sdk.AccAddress {
	return k.GetPoolStruct(ctx, pooldata).Owner
}

// SetOwner - sets the current owner of a pooldata
func (k Keeper) SetOwner(ctx sdk.Context, pooldata string, owner sdk.AccAddress) {
	poolstruct := k.GetPoolStruct(ctx, pooldata)
	poolstruct.Owner = owner
	k.SetPoolStruct(ctx, pooldata, poolstruct)
}

// GetPrice - gets the current price of a pooldata
func (k Keeper) GetPrice(ctx sdk.Context, pooldata string) sdk.Coins {
	return k.GetPoolStruct(ctx, pooldata).Price
}

// SetPrice - sets the current price of a pooldata
func (k Keeper) SetPrice(ctx sdk.Context, pooldata string, price sdk.Coins) {
	poolstruct := k.GetPoolStruct(ctx, pooldata)
	poolstruct.Price = price
	k.SetPoolStruct(ctx, pooldata, poolstruct)
}

// Get an iterator over all pooldatas in which the keys are the pooldatas and the values are the poolstruct
func (k Keeper) GetPoolDatasIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, nil)
}
