package swapservice

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/log"
	"gitlab.com/thorchain/bepswap/common"
)

type dbPrefix string

const (
	prefixTxIn              dbPrefix = "tx_"
	prefixPool              dbPrefix = "pool_"
	prefixTxOut             dbPrefix = "txout_"
	prefixPoolStaker        dbPrefix = "poolstaker_"
	prefixStakerPool        dbPrefix = "stakerpool_"
	prefixAdmin             dbPrefix = "admin_"
	prefixTxInIndex         dbPrefix = "txinIndex_"
	prefixInCompleteEvents  dbPrefix = "incomplete_events"
	prefixCompleteEvent     dbPrefix = "complete_event_"
	prefixLastEventID       dbPrefix = "last_event_id"
	prefixLastBinanceHeight dbPrefix = "last_binance_height"
	prefixLastSignedHeight  dbPrefix = "last_signed_height"
	prefixNodeAccount       dbPrefix = "node_account_"
	prefixActiveObserver    dbPrefix = "active_observer_"
	prefixPoolAddresses     dbPrefix = "pooladdresses"
	prefixValidatorMeta     dbPrefix = "validator_meta"
)

const poolIndexKey = "poolindexkey"

func getKey(prefix dbPrefix, key string) string {
	return fmt.Sprintf("%s%s", prefix, strings.ToUpper(key))
}

// Keeper maintains the link to data storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	coinKeeper   bank.Keeper
	supplyKeeper supply.Keeper
	storeKey     sdk.StoreKey // Unexposed key to access store from sdk.Context
	cdc          *codec.Codec // The wire codec for binary encoding/decoding.
}

// NewKeeper creates new instances of the swapservice Keeper
func NewKeeper(coinKeeper bank.Keeper, supplyKeeper supply.Keeper, storeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return Keeper{
		coinKeeper:   coinKeeper,
		supplyKeeper: supplyKeeper,
		storeKey:     storeKey,
		cdc:          cdc,
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

func (k Keeper) SetLastBinanceHeight(ctx sdk.Context, height sdk.Uint) error {
	currentHeight := k.GetLastBinanceHeight(ctx)
	if currentHeight.GT(height) {
		return errors.Errorf("current block height :%s is larger than %s , block height can't go backward ", currentHeight, height)
	}
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixLastBinanceHeight, "")
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(height))
	return nil
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

// GetPoolStakerIterator iterate pool stakers
func (k Keeper) GetPoolStakerIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixPoolStaker))
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

// GetStakerPoolIterator iterate stakers pools
func (k Keeper) GetStakerPoolIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixStakerPool))
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

// TotalNodeAccounts counts the number of trust accounts
func (k Keeper) TotalNodeAccounts(ctx sdk.Context) (count int) {
	nodes, _ := k.ListActiveNodeAccounts(ctx)
	return len(nodes)
}

// TotalActiveNodeAccount count the number of active node account
func (k Keeper) TotalActiveNodeAccount(ctx sdk.Context) (int, error) {
	activeNodes, err := k.ListActiveNodeAccounts(ctx)
	return len(activeNodes), err
}

// ListNodeAccounts - gets a list of all trust accounts
func (k Keeper) ListNodeAccounts(ctx sdk.Context) (NodeAccounts, error) {
	nodeAccounts := make(NodeAccounts, 0)
	naIterator := k.GetNodeAccountIterator(ctx)
	defer naIterator.Close()
	for ; naIterator.Valid(); naIterator.Next() {
		var na NodeAccount
		if err := k.cdc.UnmarshalBinaryBare(naIterator.Value(), &na); nil != err {
			return nil, errors.Wrap(err, "fail to unmarshal node account")
		}
		nodeAccounts = append(nodeAccounts, na)
	}
	return nodeAccounts, nil
}

// ListNodeAccountsByStatus - get a list of node accounts with the given status
// if status = NodeUnknown, then it return everything
func (k Keeper) ListNodeAccountsByStatus(ctx sdk.Context, status NodeStatus) (NodeAccounts, error) {
	nodeAccounts := make(NodeAccounts, 0)
	allNodeAccounts, err := k.ListNodeAccounts(ctx)
	if nil != err {
		return nodeAccounts, fmt.Errorf("fail to get all node accounts, %w", err)
	}
	for _, item := range allNodeAccounts {
		if item.Status == status {
			nodeAccounts = append(nodeAccounts, item)
		}
	}
	return nodeAccounts, nil
}

// ListActiveNodeAccounts - get a list of active trust accounts
func (k Keeper) ListActiveNodeAccounts(ctx sdk.Context) (NodeAccounts, error) {
	return k.ListNodeAccountsByStatus(ctx, NodeActive)
}

// IsWhitelistedAccount check whether the given account is white listed
func (k Keeper) IsWhitelistedNode(ctx sdk.Context, addr sdk.AccAddress) bool {
	ctx.Logger().Debug("IsWhitelistedAccount", "account address", addr.String())
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixNodeAccount, addr.String())
	return store.Has([]byte(key))
}

// GetNodeAccount try to get node account with the given address from db
func (k Keeper) GetNodeAccount(ctx sdk.Context, addr sdk.AccAddress) (NodeAccount, error) {
	ctx.Logger().Debug("GetNodeAccount", "node account", addr.String())
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixNodeAccount, addr.String())
	payload := store.Get([]byte(key))
	var na NodeAccount
	if err := k.cdc.UnmarshalBinaryBare(payload, &na); nil != err {
		return na, errors.Wrap(err, "fail to unmarshal node account")
	}
	return na, nil
}

// GetNodeAccountByObserver
func (k Keeper) GetNodeAccountByObserver(ctx sdk.Context, addr sdk.AccAddress) (NodeAccount, error) {
	ctx.Logger().Debug("GetNodeAccountByObserver", "observer address", addr.String())
	var na NodeAccount
	nodeAccounts, err := k.ListNodeAccounts(ctx)
	if nil != err {
		return na, errors.Wrap(err, "fail to get all node accounts")
	}
	for _, item := range nodeAccounts {
		if item.Accounts.ObserverBEPAddress.Equals(addr) {
			return item, nil
		}
	}
	return na, nil
}

// SetNodeAccount save the given node account into datastore
func (k Keeper) SetNodeAccount(ctx sdk.Context, na NodeAccount) {
	ctx.Logger().Debug("SetNodeAccount", "node account", na.String())
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixNodeAccount, na.NodeAddress.String())
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(na))

	// When a node is in active status, we need to add the observer address to active
	// if it is not , then we could remove them
	if na.Status == NodeActive {
		k.SetActiveObserver(ctx, na.Accounts.ObserverBEPAddress)
	} else {
		k.RemoveActiveObserver(ctx, na.Accounts.ObserverBEPAddress)
	}
}

func (k Keeper) EnsureTrustAccountUnique(ctx sdk.Context, account TrustAccount) error {
	iter := k.GetNodeAccountIterator(ctx)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var na NodeAccount
		if err := k.cdc.UnmarshalBinaryBare(iter.Value(), &na); nil != err {
			return errors.Wrap(err, "fail to unmarshal node account")
		}
		if na.Accounts.ValidatorBEPConsPubKey == account.ValidatorBEPConsPubKey {
			return errors.Errorf("%s already exist", account.ValidatorBEPConsPubKey)
		}
		if na.Accounts.SignerBNBAddress.Equals(account.SignerBNBAddress) {
			return errors.Errorf("%s already exist", account.SignerBNBAddress)
		}
		if na.Accounts.ObserverBEPAddress.Equals(account.ObserverBEPAddress) {
			return errors.Errorf("%s already exist", account.ObserverBEPAddress)
		}
	}

	return nil
}

// GetTrustAccountIterator iterate trust accounts
func (k Keeper) GetNodeAccountIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixNodeAccount))
}

// SetActiveObserver set the given addr as an active observer address
func (k Keeper) SetActiveObserver(ctx sdk.Context, addr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixActiveObserver, addr.String())
	ctx.Logger().Info("set_active_observer", "key", key)
	store.Set([]byte(key), addr.Bytes())
}

// RemoveActiveObserver remove the given address from active observer
func (k Keeper) RemoveActiveObserver(ctx sdk.Context, addr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixActiveObserver, addr.String())
	store.Delete([]byte(key))
}

// IsActiveObserver check the given account address, whether they are active
func (k Keeper) IsActiveObserver(ctx sdk.Context, addr sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixActiveObserver, addr.String())
	ctx.Logger().Info("is_active_observer", "key", key)
	return store.Has([]byte(key))
}

// SetTxHas - saving a given txhash to the KVStore
func (k Keeper) SetTxInVoter(ctx sdk.Context, tx TxInVoter) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixTxIn, tx.Key().String())
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(tx))
}

// GetTxInVoterIterator iterate tx in voters
func (k Keeper) GetTxInVoterIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixTxIn))
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

// GetTxInIndexIterator iterate tx in indexes
func (k Keeper) GetTxInIndexIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixTxInIndex))
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

// GetTxOutIterator iterate tx out
func (k Keeper) GetTxOutIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixTxOut))
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
func (k Keeper) GetAdminConfigGSL(ctx sdk.Context, addr sdk.AccAddress) common.Amount {
	return k.GetAdminConfigAmountType(ctx, GSLKey, GSLKey.Default(), addr)
}

// GetAdminConfigStakerAmtInterval - get the config for StakerAmtInterval
func (k Keeper) GetAdminConfigStakerAmtInterval(ctx sdk.Context, addr sdk.AccAddress) common.Amount {
	return k.GetAdminConfigAmountType(ctx, StakerAmtIntervalKey, StakerAmtIntervalKey.Default(), addr)
}

// GetAdminConfigMinValidatorBond get the minimum bond to become a validator
func (k Keeper) GetAdminConfigMinValidatorBond(ctx sdk.Context, addr sdk.AccAddress) sdk.Uint {
	return k.GetAdminConfigUintType(ctx, MinValidatorBondKey, MinValidatorBondKey.Default(), addr)
}

// GetAdminConfigWhiteListGasToken
func (k Keeper) GetAdminConfigWhiteListGasToken(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.GetAdminConfigCoinsType(ctx, WhiteListGasTokenKey, WhiteListGasTokenKey.Default(), addr)
}

// GetAdminConfigMRRA get the config for minimum refund rune amount default to 1 rune
func (k Keeper) GetAdminConfigMRRA(ctx sdk.Context, addr sdk.AccAddress) sdk.Uint {
	return k.GetAdminConfigUintType(ctx, MRRAKey, MRRAKey.Default(), addr)
}

// GetAdminConfigMinStakerCoins - get the min amount of coins needed to be a staker
func (k Keeper) GetAdminConfigMinStakerCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.GetAdminConfigCoinsType(ctx, MinStakerCoinsKey, MinStakerCoinsKey.Default(), addr)
}

// GetAdminConfigBnbAddressType - get the config for TSL
func (k Keeper) GetAdminConfigBnbAddressType(ctx sdk.Context, key AdminConfigKey, dValue string, addr sdk.AccAddress) common.BnbAddress {
	value, _ := k.GetAdminConfigValue(ctx, key, addr)
	if value == "" {
		value = dValue
	}
	return common.BnbAddress(value)
}

// GetAdminConfigDesireValidatorSet
func (k Keeper) GetAdminConfigDesireValidatorSet(ctx sdk.Context, addr sdk.AccAddress) int64 {
	return k.GetAdminConfigInt64(ctx, DesireValidatorSetKey, DesireValidatorSetKey.Default(), addr)
}

// GetAdminConfigRotatePerBlockHeight get rotate per block height
func (k Keeper) GetAdminConfigRotatePerBlockHeight(ctx sdk.Context, addr sdk.AccAddress) int64 {
	return k.GetAdminConfigInt64(ctx, RotatePerBlockHeightKey, RotatePerBlockHeightKey.Default(), addr)
}

// GetAdminConfigValidatorsChangeWindow get validator change window
func (k Keeper) GetAdminConfigValidatorsChangeWindow(ctx sdk.Context, addr sdk.AccAddress) int64 {
	return k.GetAdminConfigInt64(ctx, ValidatorsChangeWindowKey, ValidatorsChangeWindowKey.Default(), addr)
}

func (k Keeper) GetAdminConfigUintType(ctx sdk.Context, key AdminConfigKey, dValue string, addr sdk.AccAddress) sdk.Uint {
	value, _ := k.GetAdminConfigValue(ctx, key, addr)
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
func (k Keeper) GetAdminConfigAmountType(ctx sdk.Context, key AdminConfigKey, dValue string, addr sdk.AccAddress) common.Amount {
	value, _ := k.GetAdminConfigValue(ctx, key, addr)
	if value == "" {
		value = dValue
	}
	return common.Amount(value)
}

// GetAdminConfigCoinsType - get the config for TSL
func (k Keeper) GetAdminConfigCoinsType(ctx sdk.Context, key AdminConfigKey, dValue string, addr sdk.AccAddress) sdk.Coins {
	value, _ := k.GetAdminConfigValue(ctx, key, addr)
	if value == "" {
		value = dValue
	}
	coins, _ := sdk.ParseCoins(value)
	return coins
}

// GetAdminConfigInt64 - get the int64 config
func (k Keeper) GetAdminConfigInt64(ctx sdk.Context, key AdminConfigKey, dValue string, addr sdk.AccAddress) int64 {
	value, _ := k.GetAdminConfigValue(ctx, key, addr)
	if value == "" {
		value = dValue
	}
	result, _ := strconv.ParseInt(value, 10, 64)
	return result
}

// GetAdminConfigValue - gets the value of a given admin key
func (k Keeper) GetAdminConfigValue(ctx sdk.Context, kkey AdminConfigKey, addr sdk.AccAddress) (val string, err error) {
	getConfigValue := func(nodeAddr sdk.AccAddress) (string, error) {
		config := NewAdminConfig(kkey, "", nodeAddr)
		key := getKey(prefixAdmin, config.DbKey())
		store := ctx.KVStore(k.storeKey)
		if !store.Has([]byte(key)) {
			return kkey.Default(), nil
		}
		buf := store.Get([]byte(key))
		if err := k.cdc.UnmarshalBinaryBare(buf, &config); nil != err {
			ctx.Logger().Error(fmt.Sprintf("fail to unmarshal admin config, err: %s", err))
			return "", errors.Wrap(err, "fail to unmarshal admin config")
		}
		return config.Value, nil
	}
	// no specific bnb address given, look for consensus value
	if addr.Empty() {
		nodeAccounts, err := k.ListActiveNodeAccounts(ctx)
		if nil != err {
			return "", errors.Wrap(err, "fail to get active node accounts")
		}
		counter := make(map[string]int)
		for _, node := range nodeAccounts {
			config, err := getConfigValue(node.NodeAddress)
			if err != nil {
				return "", err
			}
			counter[config] += 1
		}

		for k, v := range counter {
			if HasMajority(v, len(nodeAccounts)) {
				return k, nil
			}
		}
	} else {
		// lookup admin config set by specific bnb address
		val, err = getConfigValue(addr)
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

// GetCompleteEventIterator iterate complete events
func (k Keeper) GetCompleteEventIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixCompleteEvent))
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
	lastEventID := k.GetLastEventID(ctx)
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
	k.SetLastEventID(ctx, lastEventID)
}

// GetLastEventID get last event id
func (k Keeper) GetLastEventID(ctx sdk.Context) common.Amount {
	var lastEventID common.Amount
	key := getKey(prefixLastEventID, "")
	store := ctx.KVStore(k.storeKey)
	if store.Has([]byte(key)) {
		buf := store.Get([]byte(key))
		_ = k.cdc.UnmarshalBinaryBare(buf, &lastEventID)
	}
	return lastEventID
}

// SetLastEventID write a last event id
func (k Keeper) SetLastEventID(ctx sdk.Context, id common.Amount) {
	key := getKey(prefixLastEventID, "")
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&id))
}

// SetPoolAddresses save the pool address to key value store
func (k Keeper) SetPoolAddresses(ctx sdk.Context, addresses PoolAddresses) {
	key := getKey(prefixPoolAddresses, "")
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(addresses))
}

// GetPoolAddresses get current pool addresses
func (k Keeper) GetPoolAddresses(ctx sdk.Context) PoolAddresses {
	var addr PoolAddresses
	key := getKey(prefixPoolAddresses, "")
	store := ctx.KVStore(k.storeKey)
	if store.Has([]byte(key)) {
		buf := store.Get([]byte(key))
		_ = k.cdc.UnmarshalBinaryBare(buf, &addr)
	}
	return addr
}

func (k Keeper) SetValidatorMeta(ctx sdk.Context, meta ValidatorMeta) {
	key := getKey(prefixValidatorMeta, "")
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(meta))
}

func (k Keeper) GetValidatorMeta(ctx sdk.Context) ValidatorMeta {
	var meta ValidatorMeta
	key := getKey(prefixValidatorMeta, "")
	store := ctx.KVStore(k.storeKey)
	if store.Has([]byte(key)) {
		buf := store.Get([]byte(key))
		_ = k.cdc.UnmarshalBinaryBare(buf, &meta)
	}
	return meta
}
