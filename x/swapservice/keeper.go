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
	prefixTxIn         dbPrefix = "tx_"
	prefixPool         dbPrefix = "pool_"
	prefixTxOut        dbPrefix = "txout_"
	prefixTrustAccount dbPrefix = "trustaccount_"
	prefixPoolStaker   dbPrefix = "poolstaker_"
	prefixStakerPool   dbPrefix = "stakerpool_"
	prefixAdmin        dbPrefix = "admin_"
	prefixTxInIndex    dbPrefix = "txinIndex_"
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
	if pool.BalanceRune.IsEmpty() {
		pool.BalanceRune = common.ZeroAmount
	}
	if pool.BalanceToken.IsEmpty() {
		pool.BalanceToken = common.ZeroAmount
	}
	if pool.PoolUnits.IsEmpty() {
		pool.PoolUnits = common.ZeroAmount
	}
	pool.PoolAddress = k.GetAdminConfigPoolAddress(ctx)

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

func (k Keeper) GetPoolBalances(ctx sdk.Context, ticker, ticker2 common.Ticker) (common.Amount, common.Amount) {
	pool := k.GetPool(ctx, ticker)
	if common.IsRune(ticker2) {
		return pool.BalanceRune, pool.BalanceToken
	}
	return pool.BalanceToken, pool.BalanceRune
}

// SetPoolData - sets the value string that a pool ID resolves to
func (k Keeper) SetPoolData(ctx sdk.Context, ticker common.Ticker, ps PoolStatus) {
	pool := k.GetPool(ctx, ticker)
	if pool.PoolUnits == "" {
		pool.PoolUnits = "0"
	}
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
		return NewPoolStaker(ticker, "0"), nil
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
func (k Keeper) SetTxIn(ctx sdk.Context, tx TxIn) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixTxIn, tx.Key().String())
	if !store.Has([]byte(key)) {
		if err := k.AddToTxInIndex(ctx, ctx.BlockHeight(), tx.Key()); nil != err {
			ctx.Logger().Error("fail to add tx id to txin index", "txid", tx.Key(), "error", err)
		}
	}
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(tx))
}

// GetTxIn - gets information of a tx hash
func (k Keeper) GetTxIn(ctx sdk.Context, hash common.TxID) TxIn {
	key := getKey(prefixTxIn, hash.String())

	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return TxIn{}
	}

	bz := store.Get([]byte(key))
	var record TxIn
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
	key := getKey(prefixAdmin, config.Key.String())
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(config))
}

// GetAdminConfigGSL - get the config for GSL
func (k Keeper) GetAdminConfigGSL(ctx sdk.Context) common.Amount {
	return k.GetAdminConfigAmountType(ctx, GSLKey, "0.3")
}

// GetAdminConfigTSL - get the config for TSL
func (k Keeper) GetAdminConfigTSL(ctx sdk.Context) common.Amount {
	return k.GetAdminConfigAmountType(ctx, TSLKey, "0.1")
}

// GetAdminConfigStakerAmtInterval - get the config for StakerAmtInterval
func (k Keeper) GetAdminConfigStakerAmtInterval(ctx sdk.Context) common.Amount {
	return k.GetAdminConfigAmountType(ctx, StakerAmtIntervalKey, "100")
}

// GetAdminConfigPoolAddress - get the config for PoolAddress
func (k Keeper) GetAdminConfigPoolAddress(ctx sdk.Context) common.BnbAddress {
	return k.GetAdminConfigBnbAddressType(ctx, PoolAddressKey, "")
}

// GetAdminConfigMRRA get the config for minimum refund rune amount default to 1 rune
func (k Keeper) GetAdminConfigMRRA(ctx sdk.Context) common.Amount {
	return k.GetAdminConfigAmountType(ctx, MRRAKey, "1")

}

// GetAdminConfigBnbAddressType - get the config for TSL
func (k Keeper) GetAdminConfigBnbAddressType(ctx sdk.Context, key AdminConfigKey, dValue string) common.BnbAddress {
	config := k.GetAdminConfig(ctx, key)
	if config.Value == "" {
		config.Value = dValue // set default
	}
	return common.BnbAddress(config.Value)
}

// GetAdminConfigAmountType - get the config for TSL
func (k Keeper) GetAdminConfigAmountType(ctx sdk.Context, key AdminConfigKey, dValue string) common.Amount {
	config := k.GetAdminConfig(ctx, key)
	if config.Value == "" {
		config.Value = dValue // set default
	}
	return common.Amount(config.Value)
}

// GetAdminConfig - gets information of a tx hash
func (k Keeper) GetAdminConfig(ctx sdk.Context, kkey AdminConfigKey) AdminConfig {
	key := getKey(prefixAdmin, kkey.String())

	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return AdminConfig{}
	}

	bz := store.Get([]byte(key))
	var record AdminConfig
	k.cdc.MustUnmarshalBinaryBare(bz, &record)

	// check if we need to add a default value
	if record.Value == "" {
		if record.Key == GSLKey {
			record.Value = "0.3" // default to 30%
		}
		if record.Key == TSLKey {
			record.Value = "0.1" // default to 10%
		}
		if record.Key == StakerAmtIntervalKey {
			record.Value = "100" // default to 100
		}
		if record.Key == MRRAKey {
			record.Value = "1" // default 1 Rune
		}
	}

	return record
}

// GetAdminConfigIterator iterate admin configs
func (k Keeper) GetAdminConfigIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixAdmin))
}
