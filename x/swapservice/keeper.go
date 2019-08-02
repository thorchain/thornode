package swapservice

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

type prefixKey string

const (
	poolKey prefixKey = "pool"
	txKey   prefixKey = "tx"
)

// Keeper maintains the link to data storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	coinKeeper bank.Keeper
	storeKey   sdk.StoreKey // Unexposed key to access store from sdk.Context
	cdc        *codec.Codec // The wire codec for binary encoding/decoding.
}

func getKey(k string, prefix prefixKey) string {
	if prefix == poolKey {
		k = strings.ToUpper(k)
	}
	return fmt.Sprintf("%s_%s", prefix, k)
}

// NewKeeper creates new instances of the swapservice Keeper
func NewKeeper(coinKeeper bank.Keeper, storeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	fmt.Println(storeKey)
	return Keeper{
		coinKeeper: coinKeeper,
		storeKey:   storeKey,
		cdc:        cdc,
	}
}

func (k Keeper) DoesExist(ctx sdk.Context, key string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has([]byte(key))
}

func (k Keeper) PoolDoesExist(ctx sdk.Context, key string) bool {
	key = getKey(key, poolKey)
	return k.DoesExist(ctx, key)
}

// Get a pool
func (k Keeper) GetPool(ctx sdk.Context, ticker string) Pool {
	key := getKey(ticker, poolKey)
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return Pool{}
	}
	bz := store.Get([]byte(key))
	var pool Pool
	k.cdc.MustUnmarshalBinaryBare(bz, &pool)

	return pool
}

// Set a pool
func (k Keeper) SetPool(ctx sdk.Context, pool Pool) {
	key := getKey(pool.Key(), poolKey)
	if pool.Empty() {
		return // cannot write an empty pool
	}
	if k.PoolDoesExist(ctx, pool.TokenTicker) {
		return // cannot overwrite a pool that already exists
	}

	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(pool))
}

// Check if tx exists
func (k Keeper) TxDoesExist(ctx sdk.Context, key string) bool {
	key = getKey(key, txKey)
	return k.DoesExist(ctx, key)
}

// Set tx
func (k Keeper) SetStakeTx(ctx sdk.Context, tx StakeTx) {
	key := getKey(tx.Key(), poolKey)
	if tx.Empty() {
		return // cannot write an empty pool
	}
	if k.TxDoesExist(ctx, tx.Key()) {
		return // cannot overwrite a pool that already exists
	}

	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(tx))
}
