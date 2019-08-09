package swapservice

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/log"
)

type dbPrefix string

const (
	prefixTxHash       dbPrefix = "tx_"
	prefixPool         dbPrefix = "pool_"
	prefixTxOut        dbPrefix = "txout_"
	prefixTrustAccount dbPrefix = "trustaccount_"
	prefixPoolStaker   dbPrefix = "poolstaker_"
	prefixStakerPool   dbPrefix = "stakerpool_"
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

// GetPoolStruct get the entire PoolStruct metadata struct for a pool ID
func (k Keeper) GetPoolStruct(ctx sdk.Context, ticker string) PoolStruct {
	key := getKey(prefixPool, ticker)
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return NewPoolStruct()
	}
	bz := store.Get([]byte(key))
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
func (k Keeper) SetPoolStruct(ctx sdk.Context, ticker string, poolstruct PoolStruct) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixPool, ticker)
	if !store.Has([]byte(key)) {
		if err := k.AddToPoolIndex(ctx, ticker); nil != err {
			ctx.Logger().Error("fail to add ticker to pool index", "ticker", ticker, "error", err)
		}
	}
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(poolstruct))
}

func (k Keeper) GetPoolBalances(ctx sdk.Context, ticker, ticker2 string) (string, string) {
	poolstruct := k.GetPoolStruct(ctx, ticker)
	if strings.EqualFold(ticker2, RuneTicker) {
		return poolstruct.BalanceRune, poolstruct.BalanceToken
	}
	return poolstruct.BalanceToken, poolstruct.BalanceRune
}

// SetPoolData - sets the value string that a pool ID resolves to
func (k Keeper) SetPoolData(ctx sdk.Context, ticker, balanceRune, balanceToken, poolAddress string, ps PoolStatus) {
	poolstruct := k.GetPoolStruct(ctx, ticker)
	if poolstruct.PoolUnits == "" {
		poolstruct.PoolUnits = "0"
	}
	poolstruct.PoolAddress = poolAddress
	poolstruct.Status = ps.String()
	poolstruct.Ticker = strings.ToUpper(ticker)
	poolstruct.BalanceRune = balanceRune
	poolstruct.BalanceToken = balanceToken
	k.SetPoolStruct(ctx, ticker, poolstruct)
}

// SetBalances - sets the current balances of a pool
func (k Keeper) SetBalances(ctx sdk.Context, ticker, rune, token string) {
	poolstruct := k.GetPoolStruct(ctx, ticker)
	poolstruct.BalanceRune = rune
	poolstruct.BalanceToken = token
	k.SetPoolStruct(ctx, ticker, poolstruct)
}

// GetPoolStructDataIterator only iterate pool data
func (k Keeper) GetPoolStructDataIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixPool))
}

// PoolExist check whether the given pool exist in the datastore
func (k Keeper) PoolExist(ctx sdk.Context, ticker string) bool {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixPool, ticker)
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
func (k Keeper) AddToPoolIndex(ctx sdk.Context, ticker string) error {
	pi, err := k.GetPoolIndex(ctx)
	if nil != err {
		return err
	}
	for _, item := range pi {
		if strings.EqualFold(item, ticker) {
			// already in the pool index , don't need to add
			return nil
		}
	}
	pi = append(pi, strings.ToUpper(ticker))
	k.SetPoolIndex(ctx, pi)
	return nil
}

// RemoveFromPoolIndex remove the given ticker from the poolIndex
func (k Keeper) RemoveFromPoolIndex(ctx sdk.Context, ticker string) error {
	pi, err := k.GetPoolIndex(ctx)
	if nil != err {
		return err
	}
	var newPI PoolIndex
	for _, item := range pi {
		if !strings.EqualFold(item, ticker) {
			newPI = append(newPI, item)
		}
	}
	k.SetPoolIndex(ctx, pi)
	return nil
}

// GetPoolStaker retrieve poolStaker from the data store
func (k Keeper) GetPoolStaker(ctx sdk.Context, ticker string) (PoolStaker, error) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixPoolStaker, ticker)
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
func (k Keeper) SetPoolStaker(ctx sdk.Context, ticker string, ps PoolStaker) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixPoolStaker, ticker)
	ctx.Logger().Info(fmt.Sprintf("key:%s ,pool staker:%s", key, ps))
	result := k.cdc.MustMarshalBinaryBare(ps)
	store.Set([]byte(key), result)
}

// GetStakerPool get the stakerpool from key value store
func (k Keeper) GetStakerPool(ctx sdk.Context, stakerID string) (StakerPool, error) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixStakerPool, stakerID)
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
func (k Keeper) SetStakerPool(ctx sdk.Context, stakerID string, sp StakerPool) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixStakerPool, stakerID)
	ctx.Logger().Info(fmt.Sprintf("key:%s ,stakerpool:%s", key, sp))
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(sp))
}

// SetSwapRecord save the swap record to store
func (k Keeper) SetSwapRecord(ctx sdk.Context, sr SwapRecord) error {
	store := ctx.KVStore(k.storeKey)
	swapRecordKey := swapRecordKeyPrefix + sr.RequestTxHash
	ctx.Logger().Debug("upsert swaprecord", "key", swapRecordKey)
	store.Set([]byte(swapRecordKey), k.cdc.MustMarshalBinaryBare(sr))
	return nil
}

// GetSwapRecord retrieve the swap record from data store.
func (k Keeper) GetSwapRecord(ctx sdk.Context, requestTxHash string) (SwapRecord, error) {
	if isEmptyString(requestTxHash) {
		return SwapRecord{}, errors.New("request tx hash is empty")
	}
	store := ctx.KVStore(k.storeKey)
	swapRecordKey := swapRecordKeyPrefix + requestTxHash
	ctx.Logger().Debug("get swap record", "key", swapRecordKey)
	if !store.Has([]byte(swapRecordKey)) {
		ctx.Logger().Debug("record not found", "key", swapRecordKey)
		return SwapRecord{
			RequestTxHash: requestTxHash,
		}, nil
	}
	var sw SwapRecord
	buf := store.Get([]byte(swapRecordKey))
	if err := k.cdc.UnmarshalBinaryBare(buf, &sw); nil != err {
		return SwapRecord{}, errors.Wrap(err, "fail to unmarshal SwapRecord")
	}
	return sw, nil
}

// UpdateSwapRecordPayTxHash update the swap record with the given paytxhash
func (k Keeper) UpdateSwapRecordPayTxHash(ctx sdk.Context, requestTxHash, payTxHash string) error {
	if isEmptyString(requestTxHash) {
		return errors.New("request tx hash is empty")
	}
	if isEmptyString(payTxHash) {
		return errors.New("pay tx hash is empty")
	}
	sr, err := k.GetSwapRecord(ctx, requestTxHash)
	if nil != err {
		return errors.Wrapf(err, "fail to get swap record with request hash:%s", requestTxHash)
	}
	sr.PayTxHash = payTxHash
	return k.SetSwapRecord(ctx, sr)
}

// GetSwapRecordIterator only iterate swap record
func (k Keeper) GetSwapRecordIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(swapRecordKeyPrefix))
}

// SetUnStakeRecord write an UnStake record to key value store
func (k Keeper) SetUnStakeRecord(ctx sdk.Context, ur UnstakeRecord) {
	store := ctx.KVStore(k.storeKey)
	unStakeRecordKey := unstakeRecordPrefix + ur.RequestTxHash
	ctx.Logger().Debug("upsert UnStake", "key", unStakeRecordKey)
	store.Set([]byte(unStakeRecordKey), k.cdc.MustMarshalBinaryBare(ur))
}

// GetUnStakeRecord query unstake record from Key Value store
func (k Keeper) GetUnStakeRecord(ctx sdk.Context, requestTxHash string) (UnstakeRecord, error) {
	if isEmptyString(requestTxHash) {
		return UnstakeRecord{}, errors.New("request tx hash is empty")
	}
	store := ctx.KVStore(k.storeKey)
	unStakeRecordKey := unstakeRecordPrefix + requestTxHash
	ctx.Logger().Debug("get UnStake record", "key", unStakeRecordKey)
	if !store.Has([]byte(unStakeRecordKey)) {
		ctx.Logger().Debug("record not found", "key", unStakeRecordKey)
		return UnstakeRecord{
			RequestTxHash: requestTxHash,
		}, nil
	}
	var ur UnstakeRecord
	buf := store.Get([]byte(unStakeRecordKey))
	if err := k.cdc.UnmarshalBinaryBare(buf, &ur); nil != err {
		return UnstakeRecord{}, errors.Wrap(err, "fail to unmarshal UnstakeRecord")
	}
	return ur, nil
}

// UpdateUnStakeRecordCompleteTxHash update the complete txHash
func (k Keeper) UpdateUnStakeRecordCompleteTxHash(ctx sdk.Context, requestTxHash, completeTxHash string) error {
	if isEmptyString(requestTxHash) {
		return errors.New("request tx hash is empty")
	}
	if isEmptyString(completeTxHash) {
		return errors.New("complete tx hash is empty")
	}
	ur, err := k.GetUnStakeRecord(ctx, requestTxHash)
	if nil != err {
		return errors.Wrapf(err, "fail to get UnStake record with request hash:%s", requestTxHash)
	}
	ur.CompleteTxHash = completeTxHash
	k.SetUnStakeRecord(ctx, ur)
	return nil
}

// GetUnstakeRecordIterator only iterate unstake record
func (k Keeper) GetUnstakeRecordIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(unstakeRecordPrefix))
}

// IsTrustAccount check whether the account is trust , and can send tx
func (k Keeper) IsTrustAccount(ctx sdk.Context, addr sdk.AccAddress) bool {
	ctx.Logger().Debug("IsTrustAccount", "account address", addr.String())
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixTrustAccount, addr.String())
	return store.Has([]byte(key))
}

// SetTrustAccount save the given trust account into data store
func (k Keeper) SetTrustAccount(ctx sdk.Context, ta TrustAccount) {
	ctx.Logger().Debug("SetTrustAccount", "trust account", ta.String())
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixTrustAccount, ta.Address.String())
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(ta))
}

// GetTrustAccountIterator iterate trust accounts
func (k Keeper) GetTrustAccountIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixTrustAccount))
}

// SetTxHas - saving a given txhash to the KVStore
func (k Keeper) SetTxHash(ctx sdk.Context, tx TxHash) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixTxHash, tx.Key())
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(tx))
}

// GetTxHash - gets information of a tx hash
func (k Keeper) GetTxHash(ctx sdk.Context, hash string) TxHash {
	key := getKey(prefixTxHash, hash)

	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return TxHash{}
	}

	bz := store.Get([]byte(key))
	var record TxHash
	k.cdc.MustUnmarshalBinaryBare(bz, &record)
	return record
}

// CheckTxHash - check to see if we have already processed a specific tx
func (k Keeper) CheckTxHash(ctx sdk.Context, hash string) bool {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixTxHash, hash)
	return store.Has([]byte(key))
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
	buf := store.Get([]byte(key))
	var txOut TxOut
	if err := k.cdc.UnmarshalBinaryBare(buf, &txOut); nil != err {
		return nil, errors.Wrap(err, "fail to unmarshal tx out")
	}
	return &txOut, nil
}
