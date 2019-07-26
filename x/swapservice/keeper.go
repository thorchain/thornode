package swapservice

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"

	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

const poolIndexKey = `poolindexkey`

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
	if !strings.HasPrefix(accID, "acc-") {
		accID = fmt.Sprintf("acc-%s", accID)
	}
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
	if !strings.HasPrefix(accID, "acc-") {
		accID = fmt.Sprintf("acc-%s", accID)
	}
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(accID), k.cdc.MustMarshalBinaryBare(accstruct))
}

// SetAccData - sets the value string that a acc ID resolves to
func (k Keeper) SetAccData(ctx sdk.Context, accID string, name, ticker, amount string) {
	if !strings.HasPrefix(accID, "acc-") {
		accID = fmt.Sprintf("acc-%s", accID)
	}
	accstruct := k.GetAccStruct(ctx, accID)
	found := false
	ticker = strings.ToUpper(ticker)
	for i, record := range accstruct.Holdings {
		if record.Ticker == ticker {
			accstruct.Holdings[i].Amount = amount
			found = true
			break
		}
	}
	if !found {
		record := Holding{
			Ticker: ticker,
			Amount: amount,
		}
		accstruct.Holdings = append(accstruct.Holdings, record)
	}
	k.SetAccStruct(ctx, accID, accstruct)
}

func (k Keeper) GetAccData(ctx sdk.Context, accID, ticker string) string {
	if !strings.HasPrefix(accID, "acc-") {
		accID = fmt.Sprintf("acc-%s", accID)
	}
	accstruct := k.GetAccStruct(ctx, accID)
	ticker = strings.ToUpper(ticker)
	for _, record := range accstruct.Holdings {
		if record.Ticker == ticker {
			return record.Amount
		}
	}
	return ""
}

// Gets the entire StakeStruct metadata struct for a stake ID
func (k Keeper) GetStakeStruct(ctx sdk.Context, stakeID string) StakeStruct {
	if !strings.HasPrefix(stakeID, "stake-") {
		stakeID = fmt.Sprintf("stake-%s", stakeID)
	}
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(stakeID)) {
		return NewStakeStruct()
	}
	bz := store.Get([]byte(stakeID))
	var stakestruct StakeStruct
	k.cdc.MustUnmarshalBinaryBare(bz, &stakestruct)
	return stakestruct
}

// Get stake data for a specific user
func (k Keeper) GetStakeData(ctx sdk.Context, stakeID, name string) AccStake {
	if !strings.HasPrefix(stakeID, "stake-") {
		stakeID = fmt.Sprintf("stake-%s", stakeID)
	}
	stakestruct := k.GetStakeStruct(ctx, stakeID)
	for _, record := range stakestruct.Stakes {
		if record.Name == name {
			return record
		}
	}
	return AccStake{
		Name:  name,
		Rune:  "0",
		Token: "0",
	}
}

// Sets the entire StakeStruct metadata struct for a stake ID
func (k Keeper) SetStakeStruct(ctx sdk.Context, stakeID string, stakestruct StakeStruct) {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(stakeID), k.cdc.MustMarshalBinaryBare(stakestruct))
}

// SetStakeData - sets the value string that a stake ID resolves to
func (k Keeper) SetStakeData(ctx sdk.Context, stakeID string, name, atom, token string) {
	parts := strings.Split(stakeID, "-")
	stakestruct := k.GetStakeStruct(ctx, stakeID)
	stakestruct.Ticker = parts[1]
	found := false
	for i, record := range stakestruct.Stakes {
		if record.Name == name {
			stakestruct.Stakes[i].Rune = atom
			stakestruct.Stakes[i].Token = token
			found = true
			break
		}
	}
	if !found {
		record := AccStake{
			Name:  name,
			Rune:  atom,
			Token: token,
		}
		stakestruct.Stakes = append(stakestruct.Stakes, record)
	}
	k.SetStakeStruct(ctx, stakeID, stakestruct)
}

// Gets the entire PoolStruct metadata struct for a pool ID
func (k Keeper) GetPoolStruct(ctx sdk.Context, poolID string) PoolStruct {
	if !strings.HasPrefix(poolID, types.PoolDataKeyPrefix) {
		poolID = types.PoolDataKeyPrefix + poolID
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
	if len(poolstruct.PoolUnits) == 0 {
		poolstruct.PoolUnits = "0"
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
func (k Keeper) GetPoolBalances(ctx sdk.Context, poolID, ticker string) (string, string) {
	poolstruct := k.GetPoolStruct(ctx, poolID)
	if strings.EqualFold(ticker, types.RuneTicker) {
		return poolstruct.BalanceRune, poolstruct.BalanceToken
	}
	return poolstruct.BalanceToken, poolstruct.BalanceRune
}

// SetPoolData - sets the value string that a pool ID resolves to
func (k Keeper) SetPoolData(ctx sdk.Context, poolID, tokenName, ticker, balanceRune, balanceToken, poolAddress string, ps types.PoolStatus) {
	poolstruct := k.GetPoolStruct(ctx, poolID)
	if len(poolstruct.PoolID) == 0 {
		poolstruct.PoolID = types.GetPoolNameFromTicker(ticker)
		poolstruct.PoolUnits = "0"
	}
	poolstruct.PoolAddress = poolAddress
	poolstruct.Status = ps.String()
	poolstruct.Ticker = strings.ToUpper(ticker)
	poolstruct.BalanceRune = balanceRune
	poolstruct.BalanceToken = balanceToken
	k.SetPoolStruct(ctx, poolID, poolstruct)
}

// SetBalances - sets the current balances of a pool
func (k Keeper) SetBalances(ctx sdk.Context, poolID, rune, token string) {
	poolstruct := k.GetPoolStruct(ctx, poolID)
	poolstruct.BalanceRune = rune
	poolstruct.BalanceToken = token
	k.SetPoolStruct(ctx, poolID, poolstruct)
}

// GetPoolStructDataIterator only iterate pool data
func (k Keeper) GetPoolStructDataIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(types.PoolDataKeyPrefix))
}

// TODO remove this method later
// Get an iterator over all pool IDs in which the keys are the pool IDs and the values are the poolstruct
func (k Keeper) GetDatasIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, nil)
}

// GetPoolIndex retrieve pool index from the data store
func (k Keeper) GetPoolIndex(ctx sdk.Context) (types.PoolIndex, error) {
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(poolIndexKey)) {
		return types.PoolIndex{}, nil
	}
	buf := store.Get([]byte(poolIndexKey))
	var pi types.PoolIndex
	if err := k.cdc.UnmarshalBinaryBare(buf, &pi); nil != err {
		log.Errorf("fail to unmarshal poolindex,err: %s", err)
		return types.PoolIndex{}, errors.Wrap(err, "fail to unmarshal poolindex")
	}
	return pi, nil
}

// SetPoolIndex write a pool index into datastore
func (k Keeper) SetPoolIndex(ctx sdk.Context, pi types.PoolIndex) {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(poolIndexKey), k.cdc.MustMarshalBinaryBare(&pi))
}

// AddToPoolIndex will add the given poolID into the poolindex
func (k Keeper) AddToPoolIndex(ctx sdk.Context, poolID string) error {
	pi, err := k.GetPoolIndex(ctx)
	if nil != err {
		return err
	}
	for _, item := range pi {
		if strings.EqualFold(item, poolID) {
			// already in the pool index , don't need to add
			return nil
		}
	}
	pi = append(pi, strings.ToUpper(poolID))
	k.SetPoolIndex(ctx, pi)
	return nil
}

// RemoveFromPoolIndex remove the given poolID from the poolIndex
func (k Keeper) RemoveFromPoolIndex(ctx sdk.Context, poolID string) error {
	pi, err := k.GetPoolIndex(ctx)
	if nil != err {
		return err
	}
	var newPI types.PoolIndex
	for _, item := range pi {
		if !strings.EqualFold(item, poolID) {
			newPI = append(newPI, item)
		}
	}
	k.SetPoolIndex(ctx, pi)
	return nil
}

// GetPoolStaker retrieve poolStaker from the data store
func (k Keeper) GetPoolStaker(ctx sdk.Context, poolID string) (types.PoolStaker, error) {
	store := ctx.KVStore(k.storeKey)
	poolStakerKey := types.PoolStakerKeyPrefix + poolID
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
	poolStakerKey := types.PoolStakerKeyPrefix + poolID
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
