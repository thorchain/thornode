package swapservice

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/log"
	"gitlab.com/thorchain/bepswap/common"
)

type dbPrefix string

const (
	prefixTxIn             dbPrefix = "tx_"
	prefixPool             dbPrefix = "pool_"
	prefixTxOut            dbPrefix = "txout_"
	prefixTrustAccount     dbPrefix = "trustaccount_"
	prefixPoolStaker       dbPrefix = "poolstaker_"
	prefixStakerPool       dbPrefix = "stakerpool_"
	prefixAdmin            dbPrefix = "admin_"
	prefixTxInIndex        dbPrefix = "txinIndex_"
	prefixInCompleteEvents dbPrefix = "incomplete_events"
	prefixCompleteEvent    dbPrefix = "complete_event_"
	prefixLastEventID      dbPrefix = "last_event_id"
)

const poolIndexKey = "poolindexkey"

func getKey(prefix dbPrefix, key string) string {
	return fmt.Sprintf("%s%s", prefix, strings.ToUpper(key))
}

// Keeper maintains the link to data storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	storeKey sdk.StoreKey // Unexposed key to access store from sdk.Context
	cdc      *codec.Codec // The wire codec for binary encoding/decoding.
}

// NewKeeper creates new instances of the swapservice Keeper
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return Keeper{
		storeKey: storeKey,
		cdc:      cdc,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", ModuleName))
}

// GetPool get the entire Pool metadata struct for a pool ID
func (k Keeper) GetPool(ctx sdk.Context, ticker common.Ticker) Pool {
	key := getKey(prefixPool, ticker.String())
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return NewPool()
	}
	bz := store.Get([]byte(key))
	var pool Pool
	k.cdc.MustUnmarshalBinaryBare(bz, &pool)
	pool.PoolAddress = k.GetAdminConfigPoolAddress(ctx, common.NoBnbAddress)
	return pool
}

// Sets the entire Pool metadata struct for a pool ID
func (k Keeper) SetPool(ctx sdk.Context, pool Pool) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixPool, pool.Ticker.String())
	if !store.Has([]byte(key)) {
		if err := k.AddToPoolIndex(ctx, pool.Ticker); nil != err {
			ctx.Logger().Error("fail to add ticker to pool index", "ticker", pool.Ticker, "error", err)
		}
	}
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(pool))
}

func (k Keeper) GetPoolBalances(ctx sdk.Context, ticker, ticker2 common.Ticker) (sdk.Uint, sdk.Uint) {
	pool := k.GetPool(ctx, ticker)
	if common.IsRune(ticker2) {
		return pool.BalanceRune, pool.BalanceToken
	}
	return pool.BalanceToken, pool.BalanceRune
}

// SetPoolData - sets the value string that a pool ID resolves to
func (k Keeper) SetPoolData(ctx sdk.Context, ticker common.Ticker, ps PoolStatus) {
	pool := k.GetPool(ctx, ticker)
	pool.Status = ps
	pool.Ticker = ticker
	k.SetPool(ctx, pool)
}

// GetPoolDataIterator only iterate pool data
func (k Keeper) GetPoolDataIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixPool))
}

// PoolExist check whether the given pool exist in the datastore
func (k Keeper) PoolExist(ctx sdk.Context, ticker common.Ticker) bool {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixPool, ticker.String())
	return store.Has([]byte(key))
}

// GetPoolIndex retrieve pool index from the data store
func (k Keeper) GetPoolIndex(ctx sdk.Context) (PoolIndex, error) {
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(poolIndexKey)) {
		return PoolIndex{}, nil
	}
	buf := store.Get([]byte(poolIndexKey))
	var pi PoolIndex
	if err := k.cdc.UnmarshalBinaryBare(buf, &pi); nil != err {
		ctx.Logger().Error(fmt.Sprintf("fail to unmarshal poolindex,err: %s", err))
		return PoolIndex{}, errors.Wrap(err, "fail to unmarshal poolindex")
	}
	return pi, nil
}

// SetPoolIndex write a pool index into datastore
func (k Keeper) SetPoolIndex(ctx sdk.Context, pi PoolIndex) {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(poolIndexKey), k.cdc.MustMarshalBinaryBare(&pi))
}

// AddToPoolIndex will add the given ticker into the poolindex
func (k Keeper) AddToPoolIndex(ctx sdk.Context, ticker common.Ticker) error {
	pi, err := k.GetPoolIndex(ctx)
	if nil != err {
		return err
	}
	for _, item := range pi {
		if item.Equals(ticker) {
			// already in the pool index , don't need to add
			return nil
		}
	}
	pi = append(pi, ticker)
	k.SetPoolIndex(ctx, pi)
	return nil
}

// RemoveFromPoolIndex remove the given ticker from the poolIndex
func (k Keeper) RemoveFromPoolIndex(ctx sdk.Context, ticker common.Ticker) error {
	pi, err := k.GetPoolIndex(ctx)
	if nil != err {
		return err
	}
	var newPI PoolIndex
	for _, item := range pi {
		if !item.Equals(ticker) {
			newPI = append(newPI, item)
		}
	}
	k.SetPoolIndex(ctx, pi)
	return nil
}

// GetPoolStaker retrieve poolStaker from the data store
func (k Keeper) GetPoolStaker(ctx sdk.Context, ticker common.Ticker) (PoolStaker, error) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixPoolStaker, ticker.String())
	if !store.Has([]byte(key)) {
		ctx.Logger().Info("NotExist", "poolstakerkey", key)
		return NewPoolStaker(ticker, sdk.ZeroUint()), nil
	}
	var ps PoolStaker
	buf := store.Get([]byte(key))
	if err := k.cdc.UnmarshalBinaryBare(buf, &ps); nil != err {
		ctx.Logger().Error("fail to unmarshal poolstaker", err)
		return PoolStaker{}, err
	}
	return ps, nil
}

// SetPoolStaker store the poolstaker to datastore
func (k Keeper) SetPoolStaker(ctx sdk.Context, ticker common.Ticker, ps PoolStaker) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixPoolStaker, ticker.String())
	ctx.Logger().Info(fmt.Sprintf("key:%s ,pool staker:%s", key, ps))
	result := k.cdc.MustMarshalBinaryBare(ps)
	store.Set([]byte(key), result)
}

// GetStakerPool get the stakerpool from key value store
func (k Keeper) GetStakerPool(ctx sdk.Context, stakerID common.BnbAddress) (StakerPool, error) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixStakerPool, stakerID.String())
	ctx.Logger().Info("get staker pool", "stakerpoolkey", key)
	if !store.Has([]byte(key)) {
		return NewStakerPool(stakerID), nil
	}
	var ps StakerPool
	buf := store.Get([]byte(key))
	if err := k.cdc.UnmarshalBinaryBare(buf, &ps); nil != err {
		ctx.Logger().Error("fail to unmarshal stakerpool", err)
		return StakerPool{}, errors.Wrap(err, "fail to unmarshal stakerpool")
	}
	return ps, nil
}

// SetStakerPool save the given stakerpool object to key value store
func (k Keeper) SetStakerPool(ctx sdk.Context, stakerID common.BnbAddress, sp StakerPool) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixStakerPool, stakerID.String())
	ctx.Logger().Info(fmt.Sprintf("key:%s ,stakerpool:%s", key, sp))
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(sp))
}

// TotalTrustAccounts counts the number of trust accounts
func (k Keeper) TotalTrustAccounts(ctx sdk.Context) (count int) {
	taIterator := k.GetTrustAccountIterator(ctx)
	defer taIterator.Close()
	for ; taIterator.Valid(); taIterator.Next() {
		count += 1
	}
	return
}

// IsTrustAccount check whether the account is trust , and can send tx
func (k Keeper) IsTrustAccount(ctx sdk.Context, addr sdk.AccAddress) bool {
	ctx.Logger().Debug("IsTrustAccount", "account address", addr.String())
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixTrustAccount, addr.String())
	return store.Has([]byte(key))
}

// IsTrustAccountBnb check whether the account is trust , and can send tx
func (k Keeper) IsTrustAccountBnb(ctx sdk.Context, addr common.BnbAddress) bool {
	ctx.Logger().Debug("IsTrustAccountBnb", "bnb address", addr.String())

	taIterator := k.GetTrustAccountIterator(ctx)
	defer taIterator.Close()
	for ; taIterator.Valid(); taIterator.Next() {
		var ta TrustAccount
		k.cdc.MustUnmarshalBinaryBare(taIterator.Value(), &ta)
		ctx.Logger().Info("IsTrustAccountBnb", "bnb1", addr.String(), "bnb2", ta.BnbAddress)
		if ta.BnbAddress.Equals(addr) {
			return true
		}
	}

	return false
}

// SetTrustAccount save the given trust account into data store
func (k Keeper) SetTrustAccount(ctx sdk.Context, ta TrustAccount) {
	ctx.Logger().Debug("SetTrustAccount", "trust account", ta.String())
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixTrustAccount, ta.RuneAddress.String())
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(ta))
}

// GetTrustAccountIterator iterate trust accounts
func (k Keeper) GetTrustAccountIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixTrustAccount))
}

// SetTxHas - saving a given txhash to the KVStore
func (k Keeper) SetTxInVoter(ctx sdk.Context, tx TxInVoter) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixTxIn, tx.Key().String())
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(tx))
}

// GetTxIn - gets information of a tx hash
func (k Keeper) GetTxInVoter(ctx sdk.Context, hash common.TxID) TxInVoter {
	key := getKey(prefixTxIn, hash.String())

	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return TxInVoter{TxID: hash}
	}

	bz := store.Get([]byte(key))
	var record TxInVoter
	k.cdc.MustUnmarshalBinaryBare(bz, &record)
	return record
}

// CheckTxHash - check to see if we have already processed a specific tx
func (k Keeper) CheckTxHash(ctx sdk.Context, hash common.TxID) bool {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixTxIn, hash.String())
	return store.Has([]byte(key))
}

// GetTxInIndex retrieve txIn by height
func (k Keeper) GetTxInIndex(ctx sdk.Context, height int64) (TxInIndex, error) {
	key := getKey(prefixTxInIndex, strconv.FormatInt(height, 10))
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return TxInIndex{}, nil
	}
	buf := store.Get([]byte(key))
	var index TxInIndex
	if err := k.cdc.UnmarshalBinaryBare(buf, &index); nil != err {
		ctx.Logger().Error(fmt.Sprintf("fail to unmarshal poolindex,err: %s", err))
		return TxInIndex{}, errors.Wrap(err, "fail to unmarshal poolindex")
	}
	return index, nil
}

// SetTxInIndex write a TxIn index into datastore
func (k Keeper) SetTxInIndex(ctx sdk.Context, height int64, index TxInIndex) {
	key := getKey(prefixTxInIndex, strconv.FormatInt(height, 10))
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&index))
}

// AddToTxInIndex will add the given txIn into the index
func (k Keeper) AddToTxInIndex(ctx sdk.Context, height int64, id common.TxID) error {
	index, err := k.GetTxInIndex(ctx, height)
	if nil != err {
		return err
	}
	for _, item := range index {
		if item.Equals(id) {
			// already in the index , don't need to add
			return nil
		}
	}
	index = append(index, id)
	k.SetTxInIndex(ctx, height, index)
	return nil
}

// SetTxOut - write the given txout information to key values tore
func (k Keeper) SetTxOut(ctx sdk.Context, blockOut *TxOut) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixTxOut, strconv.FormatInt(blockOut.Height, 10))
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(blockOut))
}

// GetTxOut - write the given txout information to key values tore
func (k Keeper) GetTxOut(ctx sdk.Context, height int64) (*TxOut, error) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixTxOut, strconv.FormatInt(height, 10))
	if !store.Has([]byte(key)) {
		return NewTxOut(height), nil
	}
	buf := store.Get([]byte(key))
	var txOut TxOut
	if err := k.cdc.UnmarshalBinaryBare(buf, &txOut); nil != err {
		return nil, errors.Wrap(err, "fail to unmarshal tx out")
	}
	return &txOut, nil
}

// SetAdminConfig - saving a given admin config to the KVStore
func (k Keeper) SetAdminConfig(ctx sdk.Context, config AdminConfig) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixAdmin, config.DbKey())
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(config))
}

// GetAdminConfigGSL - get the config for GSL
func (k Keeper) GetAdminConfigGSL(ctx sdk.Context, bnb common.BnbAddress) common.Amount {
	return k.GetAdminConfigAmountType(ctx, GSLKey, "0.3", bnb)
}

// GetAdminConfigTSL - get the config for TSL
func (k Keeper) GetAdminConfigTSL(ctx sdk.Context, bnb common.BnbAddress) common.Amount {
	return k.GetAdminConfigAmountType(ctx, TSLKey, "0.1", bnb)
}

// GetAdminConfigStakerAmtInterval - get the config for StakerAmtInterval
func (k Keeper) GetAdminConfigStakerAmtInterval(ctx sdk.Context, bnb common.BnbAddress) common.Amount {
	return k.GetAdminConfigAmountType(ctx, StakerAmtIntervalKey, "100", bnb)
}

// GetAdminConfigPoolAddress - get the config for PoolAddress
func (k Keeper) GetAdminConfigPoolAddress(ctx sdk.Context, bnb common.BnbAddress) common.BnbAddress {
	return k.GetAdminConfigBnbAddressType(ctx, PoolAddressKey, "", bnb)
}

// GetAdminConfigMRRA get the config for minimum refund rune amount default to 1 rune
func (k Keeper) GetAdminConfigMRRA(ctx sdk.Context, bnb common.BnbAddress) sdk.Uint {
	return k.GetAdminConfigUintType(ctx, MRRAKey, sdk.NewUint(One), bnb)
}

// GetAdminConfigBnbAddressType - get the config for TSL
func (k Keeper) GetAdminConfigBnbAddressType(ctx sdk.Context, key AdminConfigKey, dValue common.BnbAddress, bnb common.BnbAddress) common.BnbAddress {
	value, _ := k.GetAdminConfigValue(ctx, key, bnb)
	if value == "" {
		return dValue // set default
	}
	return common.BnbAddress(value)
}

func (k Keeper) GetAdminConfigUintType(ctx sdk.Context, key AdminConfigKey, dValue sdk.Uint, bnb common.BnbAddress) sdk.Uint {
	value, _ := k.GetAdminConfigValue(ctx, key, bnb)
	if value == "" {
		return dValue // return default
	}
	amt, err := common.NewAmount(value)
	if nil != err {
		ctx.Logger().Error("fail to parse value to float", "value", value)
	}
	return amountToUint(amt)
}

// GetAdminConfigAmountType - get the config for TSL
func (k Keeper) GetAdminConfigAmountType(ctx sdk.Context, key AdminConfigKey, dValue common.Amount, bnb common.BnbAddress) common.Amount {
	value, _ := k.GetAdminConfigValue(ctx, key, bnb)
	if value == "" {
		return dValue // return default
	}
	return common.Amount(value)
}

// GetAdminConfigValue - gets the value of a given admin key
func (k Keeper) GetAdminConfigValue(ctx sdk.Context, kkey AdminConfigKey, bnb common.BnbAddress) (string, error) {

	// If we have a BNB address, lookup the admin configuration for that
	// individual
	if !bnb.IsEmpty() {
		config := NewAdminConfig(kkey, "", bnb)
		key := getKey(prefixAdmin, config.DbKey())
		store := ctx.KVStore(k.storeKey)
		if !store.Has([]byte(key)) {
			return "", nil
		}
		buf := store.Get([]byte(key))
		if err := k.cdc.UnmarshalBinaryBare(buf, &config); nil != err {
			ctx.Logger().Error(fmt.Sprintf("fail to unmarshal admin config, err: %s", err))
			return "", errors.Wrap(err, "fail to unmarshal admin config")
		}
		return config.Value, nil
	}

	// We don't have a bnb address, find the majority vote for admin config value
	counter := make(map[string]int)
	configIterator := k.GetAdminConfigIterator(ctx)
	defer configIterator.Close()
	for ; configIterator.Valid(); configIterator.Next() {
		var config AdminConfig
		k.cdc.MustUnmarshalBinaryBare(configIterator.Value(), &config)
		// check that its the key we are asking for and the bnbAddress is a
		// trust account
		if kkey == config.Key && k.IsTrustAccountBnb(ctx, config.Address) {
			counter[config.Value] += 1
		}
	}

	// find total number of trusted account
	totalTrusted := k.TotalTrustAccounts(ctx)

	// find the most voted on value
	var ans string
	var ansCount int
	for value, num := range counter {
		if num > ansCount {
			ans = value
			ansCount = num
		}
	}

	// check if we've hit majority
	if HasMajority(ansCount, totalTrusted) {
		return ans, nil
	}

	return "", nil
}

// GetAdminConfigIterator iterate admin configs
func (k Keeper) GetAdminConfigIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixAdmin))
}

// GetIncompleteEvents retrieve incomplete events
func (k Keeper) GetIncompleteEvents(ctx sdk.Context) (Events, error) {
	key := getKey(prefixInCompleteEvents, "")
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return Events{}, nil
	}
	buf := store.Get([]byte(key))
	var events Events
	if err := k.cdc.UnmarshalBinaryBare(buf, &events); nil != err {
		ctx.Logger().Error(fmt.Sprintf("fail to unmarshal incomplete events, err: %s", err))
		return Events{}, errors.Wrap(err, "fail to unmarshal incomplete events")
	}
	return events, nil
}

// SetIncompleteEvents write incomplete events
func (k Keeper) SetIncompleteEvents(ctx sdk.Context, events Events) {
	key := getKey(prefixInCompleteEvents, "")
	store := ctx.KVStore(k.storeKey)
	if len(events) == 0 {
		store.Delete([]byte(key))
	} else {
		store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&events))
	}
}

// AddIncompleteEvents append to incomplete events
func (k Keeper) AddIncompleteEvents(ctx sdk.Context, event Event) {
	events, _ := k.GetIncompleteEvents(ctx)
	events = append(events, event)
	k.SetIncompleteEvents(ctx, events)
}

// GetCompletedEvent retrieve completed event
func (k Keeper) GetCompletedEvent(ctx sdk.Context, id int64) (Event, error) {
	key := getKey(prefixCompleteEvent, fmt.Sprintf("%d", id))
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return Event{}, nil
	}
	buf := store.Get([]byte(key))
	var event Event
	if err := k.cdc.UnmarshalBinaryBare(buf, &event); nil != err {
		ctx.Logger().Error(fmt.Sprintf("fail to unmarshal complete event, err: %s", err))
		return Event{}, errors.Wrap(err, "fail to unmarshal complete event")
	}
	return event, nil
}

// SetCompletedEvent write a completed event
func (k Keeper) SetCompletedEvent(ctx sdk.Context, event Event) {
	key := getKey(prefixCompleteEvent, fmt.Sprintf("%d", int64(event.ID.Float64())))
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&event))
}

// CompleteEvent
func (k Keeper) CompleteEvents(ctx sdk.Context, in []common.TxID, out common.TxID) {
	var lastEventID common.Amount
	key := getKey(prefixLastEventID, "")
	store := ctx.KVStore(k.storeKey)
	if store.Has([]byte(key)) {
		buf := store.Get([]byte(key))
		_ = k.cdc.UnmarshalBinaryBare(buf, &lastEventID)
	}
	eID := lastEventID.Float64()

	incomplete, _ := k.GetIncompleteEvents(ctx)

	for _, txID := range in {
		eID += 1
		var evts Events
		evts, incomplete = incomplete.PopByInHash(txID)
		for _, evt := range evts {
			if !evt.Empty() {
				evt.ID = common.NewAmountFromFloat(eID)
				evt.OutHash = out
				k.SetCompletedEvent(ctx, evt)
			}
		}
	}

	// save new list of incomplete events
	k.SetIncompleteEvents(ctx, incomplete)

	lastEventID = common.NewAmountFromFloat(eID)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(lastEventID))
}
