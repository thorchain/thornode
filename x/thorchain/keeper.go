package thorchain

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

	"gitlab.com/thorchain/bepswap/thornode/common"
)

type dbPrefix string

const (
	prefixTxIn             dbPrefix = "tx_"
	prefixPool             dbPrefix = "pool_"
	prefixTxOut            dbPrefix = "txout_"
	prefixPoolStaker       dbPrefix = "poolstaker_"
	prefixStakerPool       dbPrefix = "stakerpool_"
	prefixAdmin            dbPrefix = "admin_"
	prefixTxInIndex        dbPrefix = "txinIndex_"
	prefixInCompleteEvents dbPrefix = "incomplete_events_"
	prefixCompleteEvent    dbPrefix = "complete_event_"
	prefixLastEventID      dbPrefix = "last_event_id_"
	prefixLastChainHeight  dbPrefix = "last_chain_height_"
	prefixLastSignedHeight dbPrefix = "last_signed_height_"
	prefixNodeAccount      dbPrefix = "node_account_"
	prefixActiveObserver   dbPrefix = "active_observer_"
	prefixPoolAddresses    dbPrefix = "pooladdresses_"
	prefixValidatorMeta    dbPrefix = "validator_meta_"
	prefixSupportedChains  dbPrefix = "supported_chains_"
	prefixYggdrasilPool    dbPrefix = "yggdrasil_"
	prefixVaultData        dbPrefix = "vault_data_"
)

const poolIndexKey = "poolindexkey"

func getKey(prefix dbPrefix, key string, version int64) string {
	return fmt.Sprintf("%s_%d_%s", prefix, version, strings.ToUpper(key))
}

// Keeper maintains the link to data storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	coinKeeper   bank.Keeper
	supplyKeeper supply.Keeper
	storeKey     sdk.StoreKey // Unexposed key to access store from sdk.Context
	cdc          *codec.Codec // The wire codec for binary encoding/decoding.
}

// NewKeeper creates new instances of the thorchain Keeper
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
	key := getKey(prefixLastSignedHeight, "", getVersion(k.GetLowestActiveVersion(ctx), prefixLastSignedHeight))
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(height))
}

func (k Keeper) GetLastSignedHeight(ctx sdk.Context) (height sdk.Uint) {
	key := getKey(prefixLastSignedHeight, "", getVersion(k.GetLowestActiveVersion(ctx), prefixLastSignedHeight))
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return sdk.ZeroUint()
	}
	bz := store.Get([]byte(key))
	k.cdc.MustUnmarshalBinaryBare(bz, &height)
	return
}

func (k Keeper) SetLastChainHeight(ctx sdk.Context, chain common.Chain, height sdk.Uint) error {
	currentHeight := k.GetLastChainHeight(ctx, chain)
	if currentHeight.GT(height) {
		return errors.Errorf("current block height :%s is larger than %s , block height can't go backward ", currentHeight, height)
	}
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixLastChainHeight, chain.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixLastChainHeight))
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(height))
	return nil
}

func (k Keeper) GetLastChainHeight(ctx sdk.Context, chain common.Chain) (height sdk.Uint) {
	key := getKey(prefixLastChainHeight, chain.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixLastChainHeight))
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(key)) {
		return sdk.ZeroUint()
	}
	bz := store.Get([]byte(key))
	k.cdc.MustUnmarshalBinaryBare(bz, &height)
	return
}

// GetPool get the entire Pool metadata struct for a pool ID
func (k Keeper) GetPool(ctx sdk.Context, asset common.Asset) Pool {
	key := getKey(prefixPool, asset.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixPool))
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
	key := getKey(prefixPool, pool.Asset.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixPool))
	if !store.Has([]byte(key)) {
		if err := k.AddToPoolIndex(ctx, pool.Asset); nil != err {
			ctx.Logger().Error("fail to add asset to pool index", "asset", pool.Asset, "error", err)
		}
	}
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(pool))
}

func (k Keeper) GetPoolBalances(ctx sdk.Context, asset, asset2 common.Asset) (sdk.Uint, sdk.Uint) {
	pool := k.GetPool(ctx, asset)
	if asset2.IsRune() {
		return pool.BalanceRune, pool.BalanceAsset
	}
	return pool.BalanceAsset, pool.BalanceRune
}

// SetPoolData - sets the value string that a pool ID resolves to
func (k Keeper) SetPoolData(ctx sdk.Context, asset common.Asset, ps PoolStatus) {
	pool := k.GetPool(ctx, asset)
	pool.Status = ps
	pool.Asset = asset
	k.SetPool(ctx, pool)
}

// GetPoolDataIterator only iterate pool data
func (k Keeper) GetPoolDataIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixPool))
}

// Picks the most "deserving" pool (by most staked rune) to be enabled and
// enables it
func (k Keeper) EnableAPool(ctx sdk.Context) {
	var pools []Pool
	iterator := k.GetPoolDataIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var pool Pool
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &pool)
		if pool.Status == PoolBootstrap {
			pools = append(pools, pool)
		}
	}

	if len(pools) > 0 {
		pool := pools[0]
		for _, p := range pools {
			if pool.BalanceRune.LT(p.BalanceRune) {
				pool = p
			}
		}
		// ensure we don't enable a pool that doesn't have any rune or assets
		if pool.BalanceAsset.IsZero() || pool.BalanceRune.IsZero() {
			return
		}
		pool.Status = PoolEnabled
		k.SetPool(ctx, pool)

		eventPoolStatusWrapper(ctx, k, pool)
	}
}

// PoolExist check whether the given pool exist in the datastore
func (k Keeper) PoolExist(ctx sdk.Context, asset common.Asset) bool {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixPool, asset.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixPool))
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

// AddToPoolIndex will add the given asset into the poolindex
func (k Keeper) AddToPoolIndex(ctx sdk.Context, asset common.Asset) error {
	pi, err := k.GetPoolIndex(ctx)
	if nil != err {
		return err
	}
	for _, item := range pi {
		if item.Equals(asset) {
			// already in the pool index , don't need to add
			return nil
		}
	}
	pi = append(pi, asset)
	k.SetPoolIndex(ctx, pi)
	return nil
}

// RemoveFromPoolIndex remove the given asset from the poolIndex
func (k Keeper) RemoveFromPoolIndex(ctx sdk.Context, asset common.Asset) error {
	pi, err := k.GetPoolIndex(ctx)
	if nil != err {
		return err
	}
	var newPI PoolIndex
	for _, item := range pi {
		if !item.Equals(asset) {
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
func (k Keeper) GetPoolStaker(ctx sdk.Context, asset common.Asset) (PoolStaker, error) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixPoolStaker, asset.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixPoolStaker))
	if !store.Has([]byte(key)) {
		ctx.Logger().Info("NotExist", "poolstakerkey", key)
		return NewPoolStaker(asset, sdk.ZeroUint()), nil
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
func (k Keeper) SetPoolStaker(ctx sdk.Context, asset common.Asset, ps PoolStaker) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixPoolStaker, asset.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixPoolStaker))
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
func (k Keeper) GetStakerPool(ctx sdk.Context, stakerID common.Address) (StakerPool, error) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixStakerPool, stakerID.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixPoolStaker))
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
func (k Keeper) SetStakerPool(ctx sdk.Context, stakerID common.Address, sp StakerPool) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixStakerPool, stakerID.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixPoolStaker))
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

// GetLowestActiveVersion - get version number of lowest active node
func (k Keeper) GetLowestActiveVersion(ctx sdk.Context) int64 {
	nodes, _ := k.ListActiveNodeAccounts(ctx)
	if len(nodes) > 0 {
		version := nodes[0].Version
		for _, na := range nodes {
			if na.Version < version {
				version = na.Version
			}
		}
		return version
	}
	return 0
}

// IsWhitelistedAccount check whether the given account is white listed
func (k Keeper) IsWhitelistedNode(ctx sdk.Context, addr sdk.AccAddress) bool {
	ctx.Logger().Debug("IsWhitelistedAccount", "account address", addr.String())
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixNodeAccount, addr.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixNodeAccount))
	return store.Has([]byte(key))
}

// GetNodeAccount try to get node account with the given address from db
func (k Keeper) GetNodeAccount(ctx sdk.Context, addr sdk.AccAddress) (NodeAccount, error) {
	ctx.Logger().Debug("GetNodeAccount", "node account", addr.String())
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixNodeAccount, addr.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixNodeAccount))
	payload := store.Get([]byte(key))
	var na NodeAccount
	if err := k.cdc.UnmarshalBinaryBare(payload, &na); nil != err {
		return na, errors.Wrap(err, "fail to unmarshal node account")
	}
	return na, nil
}

// GetNodeAccountByPubKey try to get node account with the given pubkey from db
func (k Keeper) GetNodeAccountByPubKey(ctx sdk.Context, pk common.PubKey) (NodeAccount, error) {
	addr, err := pk.GetThorAddress()
	if err != nil {
		return NodeAccount{}, err
	}
	return k.GetNodeAccount(ctx, addr)
}

// GetNodeAccountByBondAddress go through data store to get node account by it's signer bnb address
func (k Keeper) GetNodeAccountByBondAddress(ctx sdk.Context, addr common.Address) (NodeAccount, error) {
	ctx.Logger().Debug("GetNodeAccountByBondAddress", "signer bnb address", addr.String())
	var na NodeAccount
	nodeAccounts, err := k.ListNodeAccounts(ctx)
	if nil != err {
		return na, fmt.Errorf("fail to get all node accounts, %w", err)
	}
	for _, item := range nodeAccounts {
		if item.BondAddress.Equals(addr) {
			return item, nil
		}
	}
	return na, nil
}

// SetNodeAccount save the given node account into datastore
func (k Keeper) SetNodeAccount(ctx sdk.Context, na NodeAccount) {
	ctx.Logger().Debug("SetNodeAccount", "node account", na.String())
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixNodeAccount, na.NodeAddress.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixNodeAccount))
	if na.Status == NodeActive {
		if na.ActiveBlockHeight == 0 {
			// the na is active, and does not have a block height when they
			// became active. This must be the first block they are active, so
			// we will set it now.
			na.ActiveBlockHeight = ctx.BlockHeight()
			na.SlashPoints = 0 // reset slash points
		}
	} else {
		if na.ActiveBlockHeight > 0 {
			// The na seems to have become a non active na. Therefore, lets
			// give them their bond rewards.
			vault := k.GetVaultData(ctx)
			// Find number of blocks they have been active for
			blockCount := ctx.BlockHeight() - na.ActiveBlockHeight
			blocks := calculateNodeAccountBondUints(ctx.BlockHeight(), na.ActiveBlockHeight, na.SlashPoints)
			// calc number of rune they are awarded
			reward := calcNodeRewards(blocks, vault.TotalBondUnits, vault.BondRewardRune)
			if reward.GT(sdk.ZeroUint()) {
				na.Bond = na.Bond.Add(reward)
				if vault.TotalBondUnits.GTE(sdk.NewUint(uint64(blockCount))) {
					vault.TotalBondUnits = vault.TotalBondUnits.Sub(sdk.NewUint(uint64(blockCount)))
				} else {
					vault.TotalBondUnits = sdk.ZeroUint()
				}
				// Minus the number of units na has (do not include slash points)
				// Minus the number of rune we have awarded them
				if vault.BondRewardRune.GTE(reward) {
					vault.BondRewardRune = vault.BondRewardRune.Sub(reward)
				} else {
					vault.BondRewardRune = sdk.ZeroUint()
				}
				k.SetVaultData(ctx, vault)
			}
		}
		na.ActiveBlockHeight = 0
	}
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(na))

	// When a node is in active status, we need to add the observer address to active
	// if it is not , then we could remove them
	if na.Status == NodeActive {
		k.SetActiveObserver(ctx, na.NodeAddress)
	} else {
		k.RemoveActiveObserver(ctx, na.NodeAddress)
	}
}

// Slash the bond of a node account
// NOTE: Should be careful not to slash too much, and have their Yggdrasil
// vault have more in funds than their bond. This could trigger them to have a
// untimely exit, stealing an amount of funds from stakers.
func (k Keeper) SlashNodeAccountBond(ctx sdk.Context, na *NodeAccount, slash sdk.Uint) {
	if slash.GT(na.Bond) {
		na.Bond = sdk.ZeroUint()
	} else {
		na.Bond = na.Bond.Sub(slash)
	}
	k.SetNodeAccount(ctx, *na)
}

// Slash the rewards of a node account
// NOTE: if we slash their rewards so much, they may do an orderly exit and
// rotate out of the active vault, wait in line to rejoin later.
func (k Keeper) SlashNodeAccountRewards(ctx sdk.Context, na *NodeAccount, pts int64) {
	na.SlashPoints += pts
	k.SetNodeAccount(ctx, *na)
}

func (k Keeper) EnsureTrustAccountUnique(ctx sdk.Context, consensusPubKey string, pubKeys common.PubKeys) error {
	iter := k.GetNodeAccountIterator(ctx)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var na NodeAccount
		if err := k.cdc.UnmarshalBinaryBare(iter.Value(), &na); nil != err {
			return errors.Wrap(err, "fail to unmarshal node account")
		}
		if na.ValidatorConsPubKey == consensusPubKey {
			return errors.Errorf("%s already exist", na.ValidatorConsPubKey)
		}
		if na.NodePubKey.Equals(pubKeys) {
			return errors.Errorf("%s already exist", pubKeys)
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
	key := getKey(prefixActiveObserver, addr.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixActiveObserver))
	ctx.Logger().Info("set_active_observer", "key", key)
	store.Set([]byte(key), addr.Bytes())
}

// RemoveActiveObserver remove the given address from active observer
func (k Keeper) RemoveActiveObserver(ctx sdk.Context, addr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixActiveObserver, addr.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixActiveObserver))
	store.Delete([]byte(key))
}

// IsActiveObserver check the given account address, whether they are active
func (k Keeper) IsActiveObserver(ctx sdk.Context, addr sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixActiveObserver, addr.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixActiveObserver))
	ctx.Logger().Info("is_active_observer", "key", key)
	return store.Has([]byte(key))
}

// SetTxHas - saving a given txhash to the KVStore
func (k Keeper) SetTxInVoter(ctx sdk.Context, tx TxInVoter) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixTxIn, tx.Key().String(), getVersion(k.GetLowestActiveVersion(ctx), prefixTxIn))
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(tx))
}

// GetTxInVoterIterator iterate tx in voters
func (k Keeper) GetTxInVoterIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixTxIn))
}

// GetTxIn - gets information of a tx hash
func (k Keeper) GetTxInVoter(ctx sdk.Context, hash common.TxID) TxInVoter {
	key := getKey(prefixTxIn, hash.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixTxIn))

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
	key := getKey(prefixTxIn, hash.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixTxIn))
	return store.Has([]byte(key))
}

// GetTxInIndexIterator iterate tx in indexes
func (k Keeper) GetTxInIndexIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixTxInIndex))
}

// GetTxInIndex retrieve txIn by height
func (k Keeper) GetTxInIndex(ctx sdk.Context, height uint64) (TxInIndex, error) {
	key := getKey(prefixTxInIndex, strconv.FormatUint(height, 10), getVersion(k.GetLowestActiveVersion(ctx), prefixTxInIndex))
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
	key := getKey(prefixTxInIndex, strconv.FormatUint(height, 10), getVersion(k.GetLowestActiveVersion(ctx), prefixTxInIndex))
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
	key := getKey(prefixTxOut, strconv.FormatUint(blockOut.Height, 10), getVersion(k.GetLowestActiveVersion(ctx), prefixTxOut))
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
	key := getKey(prefixTxOut, strconv.FormatUint(height, 10), getVersion(k.GetLowestActiveVersion(ctx), prefixTxOut))
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

// GetIncompleteEvents retrieve incomplete events
func (k Keeper) GetIncompleteEvents(ctx sdk.Context) (Events, error) {
	key := getKey(prefixInCompleteEvents, "", getVersion(k.GetLowestActiveVersion(ctx), prefixInCompleteEvents))
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
	key := getKey(prefixInCompleteEvents, "", getVersion(k.GetLowestActiveVersion(ctx), prefixInCompleteEvents))
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
	key := getKey(prefixCompleteEvent, fmt.Sprintf("%d", id), getVersion(k.GetLowestActiveVersion(ctx), prefixCompleteEvent))
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
	key := getKey(prefixCompleteEvent, fmt.Sprintf("%d", event.ID), getVersion(k.GetLowestActiveVersion(ctx), prefixCompleteEvent))
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&event))
}

// CompleteEvent
func (k Keeper) CompleteEvents(ctx sdk.Context, in []common.TxID, out common.Tx) {
	lastEventID := k.GetLastEventID(ctx)

	incomplete, _ := k.GetIncompleteEvents(ctx)

	for _, txID := range in {
		var evts Events
		evts, incomplete = incomplete.PopByInHash(txID)
		for _, evt := range evts {
			if !evt.Empty() {
				voter := k.GetTxInVoter(ctx, txID)
				evt.OutTx = append(evt.OutTx, out)
				// Check if we've seen enough OutTx to the number expected to
				// have seen by the voter.
				// Sometimes we can have voter.NumOuts be zero, for example,
				// when someone is staking there are no out txs.
				if int64(len(evt.OutTx)) >= voter.NumOuts {
					lastEventID += 1
					evt.ID = lastEventID
					k.SetCompletedEvent(ctx, evt)
				} else {
					// since we have more out event, add event back to
					// incomplete evts
					incomplete = append(incomplete, evt)
				}
			}
		}
	}

	// save new list of incomplete events
	k.SetIncompleteEvents(ctx, incomplete)

	k.SetLastEventID(ctx, lastEventID)
}

// GetLastEventID get last event id
func (k Keeper) GetLastEventID(ctx sdk.Context) int64 {
	var lastEventID int64
	key := getKey(prefixLastEventID, "", getVersion(k.GetLowestActiveVersion(ctx), prefixLastEventID))
	store := ctx.KVStore(k.storeKey)
	if store.Has([]byte(key)) {
		buf := store.Get([]byte(key))
		_ = k.cdc.UnmarshalBinaryBare(buf, &lastEventID)
	}
	return lastEventID
}

// SetLastEventID write a last event id
func (k Keeper) SetLastEventID(ctx sdk.Context, id int64) {
	key := getKey(prefixLastEventID, "", getVersion(k.GetLowestActiveVersion(ctx), prefixLastEventID))
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(&id))
}

// SetPoolAddresses save the pool address to key value store
func (k Keeper) SetPoolAddresses(ctx sdk.Context, addresses *PoolAddresses) {
	key := getKey(prefixPoolAddresses, "", getVersion(k.GetLowestActiveVersion(ctx), prefixPoolAddresses))
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(*addresses))
}

// GetPoolAddresses get current pool addresses
func (k Keeper) GetPoolAddresses(ctx sdk.Context) PoolAddresses {
	var addr PoolAddresses
	key := getKey(prefixPoolAddresses, "", getVersion(k.GetLowestActiveVersion(ctx), prefixPoolAddresses))
	store := ctx.KVStore(k.storeKey)
	if store.Has([]byte(key)) {
		buf := store.Get([]byte(key))
		_ = k.cdc.UnmarshalBinaryBare(buf, &addr)
	}
	return addr
}

func (k Keeper) SetValidatorMeta(ctx sdk.Context, meta ValidatorMeta) {
	key := getKey(prefixValidatorMeta, "", getVersion(k.GetLowestActiveVersion(ctx), prefixValidatorMeta))
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(meta))
}

func (k Keeper) GetValidatorMeta(ctx sdk.Context) ValidatorMeta {
	var meta ValidatorMeta
	key := getKey(prefixValidatorMeta, "", getVersion(k.GetLowestActiveVersion(ctx), prefixValidatorMeta))
	store := ctx.KVStore(k.storeKey)
	if store.Has([]byte(key)) {
		buf := store.Get([]byte(key))
		_ = k.cdc.UnmarshalBinaryBare(buf, &meta)
	}
	return meta
}

func (k Keeper) GetChains(ctx sdk.Context) common.Chains {
	chains := make(common.Chains, 0)
	key := getKey(prefixSupportedChains, "", getVersion(k.GetLowestActiveVersion(ctx), prefixSupportedChains))
	store := ctx.KVStore(k.storeKey)
	if store.Has([]byte(key)) {
		buf := store.Get([]byte(key))
		_ = k.cdc.UnmarshalBinaryBare(buf, &chains)
	}
	return chains
}

func (k Keeper) SupportedChain(ctx sdk.Context, chain common.Chain) bool {
	for _, ch := range k.GetChains(ctx) {
		if ch.Equals(chain) {
			return true
		}
	}
	return false
}

func (k Keeper) AddChain(ctx sdk.Context, chain common.Chain) {
	key := getKey(prefixSupportedChains, "", getVersion(k.GetLowestActiveVersion(ctx), prefixSupportedChains))
	if k.SupportedChain(ctx, chain) {
		// already added
		return
	}
	chains := k.GetChains(ctx)
	chains = append(chains, chain)
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(chains))
}

// GetYggdrasilIterator only iterate yggdrasil pools
func (k Keeper) GetYggdrasilIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixYggdrasilPool))
}

func (k Keeper) FindPubKeyOfAddress(ctx sdk.Context, addr common.Address, chain common.Chain) (common.PubKey, error) {
	iterator := k.GetYggdrasilIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var ygg Yggdrasil
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &ygg)
		address, err := ygg.PubKey.GetAddress(chain)
		if err != nil {
			return common.EmptyPubKey, err
		}
		if !address.IsEmpty() && address.Equals(addr) {
			return ygg.PubKey, nil
		}
	}
	return common.EmptyPubKey, nil
}

func (k Keeper) SetYggdrasil(ctx sdk.Context, ygg Yggdrasil) {
	key := getKey(prefixYggdrasilPool, ygg.PubKey.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixYggdrasilPool))
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(ygg))
}

func (k Keeper) GetYggdrasil(ctx sdk.Context, pk common.PubKey) Yggdrasil {
	var ygg Yggdrasil
	key := getKey(prefixYggdrasilPool, pk.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixYggdrasilPool))
	store := ctx.KVStore(k.storeKey)
	if store.Has([]byte(key)) {
		buf := store.Get([]byte(key))
		_ = k.cdc.UnmarshalBinaryBare(buf, &ygg)
	}
	if ygg.PubKey.IsEmpty() {
		ygg.PubKey = pk
	}
	return ygg
}

// ////////////////////// Vault Data //////////////////////////
func (k Keeper) GetVaultData(ctx sdk.Context) VaultData {
	data := NewVaultData()
	key := getKey(prefixVaultData, "", getVersion(k.GetLowestActiveVersion(ctx), prefixVaultData))
	store := ctx.KVStore(k.storeKey)
	if store.Has([]byte(key)) {
		buf := store.Get([]byte(key))
		_ = k.cdc.UnmarshalBinaryBare(buf, &data)
	}
	return data
}

func (k Keeper) SetVaultData(ctx sdk.Context, data VaultData) {
	key := getKey(prefixVaultData, "", getVersion(k.GetLowestActiveVersion(ctx), prefixVaultData))
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(data))
}

// Update the vault data to reflect changing in this block
func (k Keeper) UpdateVaultData(ctx sdk.Context) {
	vault := k.GetVaultData(ctx)

	bondReward, totalPoolRewards := calcBlockRewards(vault.TotalReserve)
	vault.TotalReserve = vault.TotalReserve.Sub(bondReward).Sub(totalPoolRewards)
	vault.BondRewardRune = vault.BondRewardRune.Add(bondReward)

	// Pass out block rewards to stakers via placing rune into pools relative
	// to the pool's depth (amount of rune).
	totalRune := sdk.ZeroUint()
	assets, _ := k.GetPoolIndex(ctx)
	var pools []Pool
	for _, asset := range assets {
		pool := k.GetPool(ctx, asset)
		if pool.IsEnabled() && !pool.BalanceRune.IsZero() {
			totalRune = totalRune.Add(pool.BalanceRune)
			pools = append(pools, pool)
		}
	}
	poolRewards := calcPoolRewards(totalPoolRewards, totalRune, pools)
	for i, reward := range poolRewards {
		pool := pools[i]
		pool.BalanceRune = pool.BalanceRune.Add(reward)
		k.SetPool(ctx, pool)
	}

	i, _ := k.TotalActiveNodeAccount(ctx)
	vault.TotalBondUnits = vault.TotalBondUnits.Add(sdk.NewUint(uint64(i)))

	k.SetVaultData(ctx, vault)
}

// /////////////////////////////////////////////////////
