package swapservice

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

const poolIndexKey = `internal-pool-indexes`

// Keeper maintains the link to data storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	coinKeeper bank.Keeper
	storeKey   sdk.StoreKey // Unexposed key to access store from sdk.Context
	cdc        *codec.Codec // The wire codec for binary encoding/decoding.
}

// NewKeeper creates new instances of the swapservice Keeper
func NewKeeper(coinKeeper bank.Keeper, storeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return Keeper{
		coinKeeper: coinKeeper,
		storeKey:   storeKey,
		cdc:        cdc,
	}
}

// Gets the entire PoolStruct metadata struct for a pool ID
func (k Keeper) GetPoolStruct(ctx sdk.Context, poolID string) PoolStruct {
	if !strings.HasPrefix(poolID, "pool-") {
		poolID = fmt.Sprintf("pool-%s", poolID)
	}
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(poolID)) {
		return NewPoolStruct()
	}
	bz := store.Get([]byte(poolID))
	var poolstruct PoolStruct
	k.cdc.MustUnmarshalBinaryBare(bz, &poolstruct)
	if poolstruct.BalanceRune == "" {
		poolstruct.BalanceRune = "0"
	}
	if poolstruct.BalanceToken == "" {
		poolstruct.BalanceToken = "0"
	}
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
		return poolstruct.BalanceRune, poolstruct.BalanceToken
	}
	return poolstruct.BalanceToken, poolstruct.BalanceRune
}

// SetPoolData - sets the value string that a pool ID resolves to
func (k Keeper) SetPoolData(ctx sdk.Context, poolID string, tokenName, ticker, balanceAtom, balanceToken string) {
	poolstruct := k.GetPoolStruct(ctx, poolID)
	poolstruct.TokenName = tokenName
	poolstruct.Ticker = strings.ToUpper(ticker)
	poolstruct.BalanceRune = balanceAtom
	poolstruct.BalanceToken = balanceToken
	k.SetPoolStruct(ctx, poolID, poolstruct)
}

// SetBalances - sets the current balances of a pool
func (k Keeper) SetBalances(ctx sdk.Context, poolID, r, token string) {
	poolstruct := k.GetPoolStruct(ctx, poolID)
	poolstruct.BalanceRune = r
	poolstruct.BalanceToken = token
	k.SetPoolStruct(ctx, poolID, poolstruct)
}

// Get an iterator over all pool IDs in which the keys are the pool IDs and the values are the poolstruct
func (k Keeper) GetDatasIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, nil)
}

// GetPoolIndex retrieve pool index from the data store
func (k Keeper) GetPoolIndex(ctx sdk.Context) types.PoolIndex {
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(poolIndexKey)) {
		return types.PoolIndex{}
	}
	buf := store.Get([]byte(poolIndexKey))
	var pi types.PoolIndex
	if err := k.cdc.UnmarshalBinaryBare(buf, &pi); nil != err {
		log.Errorf("fail to unmarshal poolindex,err: %s", err)
	}
	return pi
}

// SetPoolIndex write a pool index into datastore
func (k Keeper) SetPoolIndex(ctx sdk.Context, pi types.PoolIndex) {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(poolIndexKey), k.cdc.MustMarshalBinaryBare(&pi))
}

// AddToPoolIndex will add the given poolID into the poolindex
func (k Keeper) AddToPoolIndex(ctx sdk.Context, poolID string) {
	pi := k.GetPoolIndex(ctx)
	for _, item := range pi {
		if strings.EqualFold(item, poolID) {
			// already in the pool index , don't need to add
			return
		}
	}
	pi = append(pi, strings.ToUpper(poolID))
	k.SetPoolIndex(ctx, pi)
}

// RemoveFromPoolIndex remove the given poolID from the poolIndex
func (k Keeper) RemoveFromPoolIndex(ctx sdk.Context, poolID string) {
	pi := k.GetPoolIndex(ctx)
	var newPI types.PoolIndex
	for _, item := range pi {
		if !strings.EqualFold(item, poolID) {
			newPI = append(newPI, item)
		}
	}
	k.SetPoolIndex(ctx, pi)
}

// GetPoolStaker retrieve poolStaker from the data store
func (k Keeper) GetPoolStaker(ctx sdk.Context, poolID string) (types.PoolStaker, error) {
	store := ctx.KVStore(k.storeKey)
	poolStakerKey := "poolstaker-" + poolID
	if !store.Has([]byte(poolStakerKey)) {
		return types.NewPoolStaker(poolID, "0"), nil
	}
	var ps types.PoolStaker
	buf := store.Get([]byte(poolStakerKey))
	if err := k.cdc.UnmarshalBinaryBare(buf, &ps); nil != err {
		log.Errorf("fail to unmarshal poolstaker,err: %s", err)
		return types.PoolStaker{}, err
	}
	return ps, nil
}

// SetPoolStaker store the poolstaker to datastore
func (k Keeper) SetPoolStaker(ctx sdk.Context, poolID string, ps types.PoolStaker) {
	store := ctx.KVStore(k.storeKey)
	poolStakerKey := "poolstaker-" + poolID
	store.Set([]byte(poolStakerKey), k.cdc.MustMarshalBinaryBare(&ps))
}

// AddStaker will add a staker into the PoolStaker
func (k Keeper) AddStaker(ctx sdk.Context, poolID, totalUnits, stakerID, stakerUnits string) error {
	ps, err := k.GetPoolStaker(ctx, poolID)
	if nil != err {
		return errors.Wrap(err, "fail to get poolstake from data store")
	}
	ps.TotalUnits = totalUnits
	ps.Stakers[stakerID] = stakerUnits
	k.SetPoolStaker(ctx, poolID, ps)
	return nil
}
