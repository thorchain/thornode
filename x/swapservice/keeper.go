package swapservice

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/pkg/errors"

	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

const poolIndexKey = `poolindexkey`

// Keeper maintains the link to data storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	coinKeeper   bank.Keeper
	supplyKeeper *supply.Keeper
	storeKey     sdk.StoreKey // Unexposed key to access store from sdk.Context
	cdc          *codec.Codec // The wire codec for binary encoding/decoding.
}

// NewKeeper creates new instances of the swapservice Keeper
func NewKeeper(coinKeeper bank.Keeper, supplyKeeper *supply.Keeper, storeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return Keeper{
		coinKeeper:   coinKeeper,
		storeKey:     storeKey,
		supplyKeeper: supplyKeeper,
		cdc:          cdc,
	}
}

// GetPoolStruct get the entire PoolStruct metadata struct for a pool ID
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
	if !store.Has([]byte(poolID)) {
		if err := k.AddToPoolIndex(ctx, poolID); nil != err {
			ctx.Logger().Error("fail to add poolID to pool index", "poolID", poolID, "error", err)
		}
	}
	store.Set([]byte(poolID), k.cdc.MustMarshalBinaryBare(poolstruct))
}

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

// PoolExist check whether the given pool exist in the datastore
func (k Keeper) PoolExist(ctx sdk.Context, poolID string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has([]byte(poolID))
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
		ctx.Logger().Error(fmt.Sprintf("fail to unmarshal poolindex,err: %s", err))
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
		ctx.Logger().Info("NotExist", "poolstakerkey", poolStakerKey)
		return types.NewPoolStaker(poolID, "0"), nil
	}
	var ps types.PoolStaker
	buf := store.Get([]byte(poolStakerKey))
	if err := k.cdc.UnmarshalBinaryBare(buf, &ps); nil != err {
		ctx.Logger().Error("fail to unmarshal poolstaker", err)
		return types.PoolStaker{}, err
	}
	return ps, nil
}

// SetPoolStaker store the poolstaker to datastore
func (k Keeper) SetPoolStaker(ctx sdk.Context, poolID string, ps types.PoolStaker) {
	store := ctx.KVStore(k.storeKey)
	poolStakerKey := types.PoolStakerKeyPrefix + poolID
	ctx.Logger().Info(fmt.Sprintf("key:%s ,pool staker:%s", poolStakerKey, ps))
	result := k.cdc.MustMarshalBinaryBare(ps)
	store.Set([]byte(poolStakerKey), result)
}

// GetStakerPool get the stakerpool from key value store
func (k Keeper) GetStakerPool(ctx sdk.Context, stakerID string) (types.StakerPool, error) {
	store := ctx.KVStore(k.storeKey)
	stakerPoolKey := types.StakerPoolKeyPrefix + stakerID
	ctx.Logger().Info("get staker pool", "stakerpoolkey", stakerPoolKey)
	if !store.Has([]byte(stakerPoolKey)) {
		return types.NewStakerPool(stakerID), nil
	}
	var ps types.StakerPool
	buf := store.Get([]byte(stakerPoolKey))
	if err := k.cdc.UnmarshalBinaryBare(buf, &ps); nil != err {
		ctx.Logger().Error("fail to unmarshal stakerpool", err)
		return types.StakerPool{}, errors.Wrap(err, "fail to unmarshal stakerpool")
	}
	return ps, nil
}

// SetStakerPool save the given stakerpool object to key value store
func (k Keeper) SetStakerPool(ctx sdk.Context, stakerID string, sp types.StakerPool) {
	store := ctx.KVStore(k.storeKey)
	stakerPoolKey := types.StakerPoolKeyPrefix + stakerID
	ctx.Logger().Info(fmt.Sprintf("key:%s ,stakerpool:%s", stakerPoolKey, sp))
	store.Set([]byte(stakerPoolKey), k.cdc.MustMarshalBinaryBare(sp))
}

// SetSwapRecord save the swap record to store
func (k Keeper) SetSwapRecord(ctx sdk.Context, sr types.SwapRecord) error {
	store := ctx.KVStore(k.storeKey)
	swapRecordKey := swapRecordKeyPrefix + sr.RequestTxHash
	ctx.Logger().Debug("upsert swaprecord", "key", swapRecordKey)
	store.Set([]byte(swapRecordKey), k.cdc.MustMarshalBinaryBare(sr))
	return nil
}

// GetSwapRecord retrieve the swap record from data store.
func (k Keeper) GetSwapRecord(ctx sdk.Context, requestTxHash string) (types.SwapRecord, error) {
	if isEmptyString(requestTxHash) {
		return types.SwapRecord{}, errors.New("request tx hash is empty")
	}
	store := ctx.KVStore(k.storeKey)
	swapRecordKey := swapRecordKeyPrefix + requestTxHash
	ctx.Logger().Debug("get swap record", "key", swapRecordKey)
	if !store.Has([]byte(swapRecordKey)) {
		ctx.Logger().Debug("record not found", "key", swapRecordKey)
		return types.SwapRecord{
			RequestTxHash: requestTxHash,
		}, nil
	}
	var sw types.SwapRecord
	buf := store.Get([]byte(swapRecordKey))
	if err := k.cdc.UnmarshalBinaryBare(buf, &sw); nil != err {
		return types.SwapRecord{}, errors.Wrap(err, "fail to unmarshal SwapRecord")
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

func (k Keeper) AddSomeCoins(ctx sdk.Context, address sdk.AccAddress) {
	coins := sdk.Coins{
		sdk.Coin{
			Denom:  "bnb",
			Amount: sdk.NewInt(10000),
		},
		sdk.Coin{
			Denom:  "btc",
			Amount: sdk.NewInt(10000),
		},
	}
	if nil == k.supplyKeeper {
		ctx.Logger().Error("supply keeper is nil")
	}
	err := k.supplyKeeper.MintCoins(ctx, ModuleName, coins)
	if nil != err {
		ctx.Logger().Error("fail to mint coins", "error", err)
		return
	}
	if err := k.coinKeeper.SendCoins(ctx, k.supplyKeeper.GetModuleAddress(ModuleName), address, coins); nil != err {
		ctx.Logger().Error("fail to give you coins", "error", err)
		return
	}
}
