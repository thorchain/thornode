package swapservice

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/log"
	"gitlab.com/thorchain/bepswap/common"
)

type dbPrefix string

const (
	prefixTxIn              dbPrefix = "tx_"
	prefixPool              dbPrefix = "pool_"
	prefixTxOut             dbPrefix = "txout_"
	prefixTrustAccount      dbPrefix = "trustaccount_"
	prefixPoolStaker        dbPrefix = "poolstaker_"
	prefixStakerPool        dbPrefix = "stakerpool_"
	prefixAdmin             dbPrefix = "admin_"
	prefixTxInIndex         dbPrefix = "txinIndex_"
	prefixInCompleteEvents  dbPrefix = "incomplete_events"
	prefixCompleteEvent     dbPrefix = "complete_event_"
	prefixLastEventID       dbPrefix = "last_event_id"
	prefixLastBinanceHeight dbPrefix = "last_binance_height"
	prefixLastSignedHeight  dbPrefix = "last_signed_height"
)

const poolIndexKey = "poolindexkey"

func getKey(prefix dbPrefix, key string) string {
	return fmt.Sprintf("%s%s", prefix, strings.ToUpper(key))
}

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

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", ModuleName))
}

func (k Keeper) SetLastSignedHeight(ctx sdk.Context, height sdk.Uint) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixLastSignedHeight, "")
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(height))
}

func (k Keeper) GetLastSignedHeight(ctx sdk.Context) (height sdk.Uint) {
	key := getKey(prefixLastSignedHeight, "")
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return sdk.ZeroUint()
	}
	bz := store.Get([]byte(key))
	k.cdc.MustUnmarshalBinaryBare(bz, &height)
	return
}

func (k Keeper) SetLastBinanceHeight(ctx sdk.Context, height sdk.Uint) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixLastBinanceHeight, "")
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(height))
}

func (k Keeper) GetLastBinanceHeight(ctx sdk.Context) (height sdk.Uint) {
	key := getKey(prefixLastBinanceHeight, "")
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return sdk.ZeroUint()
	}
	bz := store.Get([]byte(key))
	k.cdc.MustUnmarshalBinaryBare(bz, &height)
	return
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
	pool.ExpiryUtc = k.GetAdminConfigPoolExpiry(ctx, common.NoBnbAddress)

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

// ListTrustAccounts - gets a list of all trust accounts
func (k Keeper) ListTrustAccounts(ctx sdk.Context) TrustAccounts {
	var trustAccounts []TrustAccount
	taIterator := k.GetTrustAccountIterator(ctx)
	defer taIterator.Close()
	for ; taIterator.Valid(); taIterator.Next() {
		var ta TrustAccount
		k.cdc.MustUnmarshalBinaryBare(taIterator.Value(), &ta)
		trustAccounts = append(trustAccounts, ta)
	}
	return trustAccounts
}

// ListActiveTrustAccounts - get a list of active trust accounts
func (k Keeper) ListActiveTrustAccounts(ctx sdk.Context) TrustAccounts {
	all := k.ListTrustAccounts(ctx)
	trusts := make(TrustAccounts, 0)

	// Count the votes for min coins needed to be an active trusted account
	// We ignore any vote that is more coins than they themselves have. This
	// ensures that we can never kick out all validators with a super high
	// number, and that validators who are not active, still cannot vote
	counter := make(map[string]int)
	var total int
	for _, trust := range all {
		minCoins := k.GetAdminConfigMinStakerCoins(ctx, trust.AdminAddress)
		if k.coinKeeper.HasCoins(ctx, trust.ObserverAddress, minCoins) {
			counter[minCoins.String()] += 1
			total += 1
		}
	}

	// discover the majority min coins vote, and add any trust accounts that
	// meet that requirement to trusts
	for min, count := range counter {
		if HasMajority(count, total) {
			minCoins, _ := sdk.ParseCoins(min)
			for _, trust := range all {
				if k.coinKeeper.HasCoins(ctx, trust.ObserverAddress, minCoins) {
					trusts = append(trusts, trust)
				}
			}
			break
		}
	}
	return trusts
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
		ctx.Logger().Info("IsTrustAccountBnb", "bnb1", addr.String(), "bnb2", ta.AdminAddress)
		if ta.AdminAddress.Equals(addr) {
			return true
		}
	}

	return false
}

// SetTrustAccount save the given trust account into data store
func (k Keeper) SetTrustAccount(ctx sdk.Context, ta TrustAccount) {
	ctx.Logger().Debug("SetTrustAccount", "trust account", ta.String())
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixTrustAccount, ta.ObserverAddress.String())
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
func (k Keeper) GetTxInIndex(ctx sdk.Context, height uint64) (TxInIndex, error) {
	key := getKey(prefixTxInIndex, strconv.FormatUint(height, 10))
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
func (k Keeper) SetTxInIndex(ctx sdk.Context, height uint64, index TxInIndex) {
	key := getKey(prefixTxInIndex, strconv.FormatUint(height, 10))
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&index))
}

// AddToTxInIndex will add the given txIn into the index
func (k Keeper) AddToTxInIndex(ctx sdk.Context, height uint64, id common.TxID) error {
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
	key := getKey(prefixTxOut, strconv.FormatUint(blockOut.Height, 10))
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(blockOut))
}

// GetTxOut - write the given txout information to key values tore
func (k Keeper) GetTxOut(ctx sdk.Context, height uint64) (*TxOut, error) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixTxOut, strconv.FormatUint(height, 10))
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
	return k.GetAdminConfigAmountType(ctx, GSLKey, GSLKey.Default(), bnb)
}

// GetAdminConfigTSL - get the config for TSL
func (k Keeper) GetAdminConfigTSL(ctx sdk.Context, bnb common.BnbAddress) common.Amount {
	return k.GetAdminConfigAmountType(ctx, TSLKey, TSLKey.Default(), bnb)
}

// GetAdminConfigStakerAmtInterval - get the config for StakerAmtInterval
func (k Keeper) GetAdminConfigStakerAmtInterval(ctx sdk.Context, bnb common.BnbAddress) common.Amount {
	return k.GetAdminConfigAmountType(ctx, StakerAmtIntervalKey, StakerAmtIntervalKey.Default(), bnb)
}

// GetAdminConfigPoolAddress - get the config for PoolAddress
func (k Keeper) GetAdminConfigPoolAddress(ctx sdk.Context, bnb common.BnbAddress) common.BnbAddress {
	return k.GetAdminConfigBnbAddressType(ctx, PoolAddressKey, PoolAddressKey.Default(), bnb)
}

// GetAdminConfigPoolExpiry get the config for pool address expiry
func (k Keeper) GetAdminConfigPoolExpiry(ctx sdk.Context, bnb common.BnbAddress) time.Time {
	expiry, err := k.GetAdminConfigValue(ctx, PoolExpiryKey, bnb)
	if nil != err {
		ctx.Logger().Error("fail to get pool address expiry", "error", err)
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, expiry)
	if nil != err {
		ctx.Logger().Error("fail to parse pool address expiry", "error", err)
		return time.Time{}
	}
	return t
}

// GetAdminConfigMRRA get the config for minimum refund rune amount default to 1 rune
func (k Keeper) GetAdminConfigMRRA(ctx sdk.Context, bnb common.BnbAddress) sdk.Uint {
	return k.GetAdminConfigUintType(ctx, MRRAKey, MRRAKey.Default(), bnb)
}

// GetAdminConfigMinStakerCoins - get the min amount of coins needed to be a staker
func (k Keeper) GetAdminConfigMinStakerCoins(ctx sdk.Context, bnb common.BnbAddress) sdk.Coins {
	return k.GetAdminConfigCoinsType(ctx, MinStakerCoinsKey, MinStakerCoinsKey.Default(), bnb)
}

// GetAdminConfigBnbAddressType - get the config for TSL
func (k Keeper) GetAdminConfigBnbAddressType(ctx sdk.Context, key AdminConfigKey, dValue string, bnb common.BnbAddress) common.BnbAddress {
	value, _ := k.GetAdminConfigValue(ctx, key, bnb)
	if value == "" {
		value = dValue
	}
	return common.BnbAddress(value)
}

func (k Keeper) GetAdminConfigUintType(ctx sdk.Context, key AdminConfigKey, dValue string, bnb common.BnbAddress) sdk.Uint {
	value, _ := k.GetAdminConfigValue(ctx, key, bnb)
	if value == "" {
		value = dValue
	}
	amt, err := common.NewAmount(value)
	if nil != err {
		ctx.Logger().Error("fail to parse value to float", "value", value)
	}
	return common.AmountToUint(amt)
}

// GetAdminConfigAmountType - get the config for TSL
func (k Keeper) GetAdminConfigAmountType(ctx sdk.Context, key AdminConfigKey, dValue string, bnb common.BnbAddress) common.Amount {
	value, _ := k.GetAdminConfigValue(ctx, key, bnb)
	if value == "" {
		value = dValue
	}
	return common.Amount(value)
}

// GetAdminConfigCoinsType - get the config for TSL
func (k Keeper) GetAdminConfigCoinsType(ctx sdk.Context, key AdminConfigKey, dValue string, bnb common.BnbAddress) sdk.Coins {
	value, _ := k.GetAdminConfigValue(ctx, key, bnb)
	if value == "" {
		value = dValue
	}
	coins, _ := sdk.ParseCoins(value)
	return coins
}

// GetAdminConfigValue - gets the value of a given admin key
func (k Keeper) GetAdminConfigValue(ctx sdk.Context, kkey AdminConfigKey, bnb common.BnbAddress) (val string, err error) {

	getConfigValue := func(bnb common.BnbAddress) (string, error) {
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

	// no specific bnb address given, look for consensus value
	if bnb.IsEmpty() {
		trustAccounts := k.ListActiveTrustAccounts(ctx)
		counter := make(map[string]int)
		for _, trust := range trustAccounts {
			config, err := getConfigValue(trust.AdminAddress)
			if err != nil {
				return "", err
			}
			counter[config] += 1
		}

		for k, v := range counter {
			if HasMajority(v, len(trustAccounts)) {
				return k, nil
			}
		}
	} else {
		// lookup admin config set by specific bnb address
		val, err = getConfigValue(bnb)
		if err != nil {
			return val, err
		}
	}

	if val == "" {
		val = kkey.Default()
	}

	return val, err
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
