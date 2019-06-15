package swapservice

import (
	"log"
	"strings"

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

// Gets the entire AccStruct metadata struct for a acc ID
func (k Keeper) GetAccStruct(ctx sdk.Context, accID string) AccStruct {
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(accID)) {
		return NewAccStruct()
	}
	bz := store.Get([]byte(accID))
	var accstruct AccStruct
	k.cdc.MustUnmarshalBinaryBare(bz, &accstruct)
	return accstruct
}

// Sets the entire AccStruct metadata struct for a acc ID
func (k Keeper) SetAccStruct(ctx sdk.Context, accID string, accstruct AccStruct) {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(accID), k.cdc.MustMarshalBinaryBare(accstruct))
}

// SetAccData - sets the value string that a acc ID resolves to
func (k Keeper) SetAccData(ctx sdk.Context, accID string, name, atom, btc string) {
	accstruct := k.GetAccStruct(ctx, accID)
	accstruct.Name = strings.ToLower(name)
	accstruct.ATOM = atom
	accstruct.BTC = btc
	k.SetAccStruct(ctx, accID, accstruct)
}

// Gets the entire PoolStruct metadata struct for a pool ID
func (k Keeper) GetPoolStruct(ctx sdk.Context, poolID string) PoolStruct {
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(poolID)) {
		return NewPoolStruct()
	}
	bz := store.Get([]byte(poolID))
	var poolstruct PoolStruct
	k.cdc.MustUnmarshalBinaryBare(bz, &poolstruct)
	return poolstruct
}

// Sets the entire PoolStruct metadata struct for a pool ID
func (k Keeper) SetPoolStruct(ctx sdk.Context, poolID string, poolstruct PoolStruct) {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(poolID), k.cdc.MustMarshalBinaryBare(poolstruct))
}

// GetPool - gets the balances of a pool. Specifying ticker dictates which
// balance is return in 0 vs 1 spot.
func (k Keeper) GetPoolData(ctx sdk.Context, poolID, ticker string) (string, string) {
	poolstruct := k.GetPoolStruct(ctx, poolID)
	if strings.ToUpper(ticker) == "ATOM" {
		return poolstruct.BalanceAtom, poolstruct.BalanceToken
	}
	return poolstruct.BalanceToken, poolstruct.BalanceAtom
}

// SetPoolData - sets the value string that a pool ID resolves to
func (k Keeper) SetPoolData(ctx sdk.Context, poolID string, tokenName, ticker, balanceAtom, balanceToken string) {
	poolstruct := k.GetPoolStruct(ctx, poolID)
	poolstruct.TokenName = tokenName
	poolstruct.Ticker = strings.ToUpper(ticker)
	poolstruct.BalanceAtom = balanceAtom
	poolstruct.BalanceToken = balanceToken
	log.Printf("Pool ID: %s", poolID)
	log.Printf("SetPoolData: %s", poolstruct)
	k.SetPoolStruct(ctx, poolID, poolstruct)
}

// SetBalances - sets the current balances of a pool
func (k Keeper) SetBalances(ctx sdk.Context, poolID, atom, token string) {
	poolstruct := k.GetPoolStruct(ctx, poolID)
	poolstruct.BalanceAtom = atom
	poolstruct.BalanceToken = token
	k.SetPoolStruct(ctx, poolID, poolstruct)
}

// Get an iterator over all pool IDs in which the keys are the pool IDs and the values are the poolstruct
func (k Keeper) GetPoolDatasIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, nil)
}
