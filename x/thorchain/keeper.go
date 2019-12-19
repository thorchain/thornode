package thorchain

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/log"
)

type Keeper interface {
	Cdc() *codec.Codec
	Supply() supply.Keeper
	CoinKeeper() bank.Keeper
	Logger(ctx sdk.Context) log.Logger
	GetKey(ctx sdk.Context, prefix dbPrefix, key string) string

	// Keeper Interfaces
	KeeperPool
	KeeperLastHeight
	KeeperPoolStaker
	KeeperStakerPool
	KeeperNodeAccount
	KeeperObserver
	KeeperObservedTx
	KeeperTxOut
	KeeperLiquidityFees
	KeeperEvents
	KeeperPoolAddresses
	KeeperChains
	KeeperVault
	KeeperReserveContributors
	KeeperVaultData
	KeeperAdminConfig
	KeeperTss
	KeeperKeygen
}

// NOTE: Always end a dbPrefix with a slash ("/"). This is to ensure that there
// are no prefixes that contain another prefix. In the scenario where this is
// true, an iterator for a specific type, will get more than intended, and may
// include a different type. The slash is used to protect us from this
// scenario.
// Also, use underscores between words and use lowercase characters only
type dbPrefix string

const (
	prefixObservedTx         dbPrefix = "observed_tx/"
	prefixPool               dbPrefix = "pool/"
	prefixTxOut              dbPrefix = "txout/"
	prefixTotalLiquidityFee  dbPrefix = "total_liquidity_fee/"
	prefixPoolLiquidityFee   dbPrefix = "pool_liquidity_fee/"
	prefixPoolStaker         dbPrefix = "pool_staker/"
	prefixStakerPool         dbPrefix = "staker_pool/"
	prefixAdmin              dbPrefix = "admin/"
	prefixInCompleteEvents   dbPrefix = "incomplete_events/"
	prefixCompleteEvent      dbPrefix = "complete_event/"
	prefixLastEventID        dbPrefix = "last_event_id/"
	prefixLastChainHeight    dbPrefix = "last_chain_height/"
	prefixLastSignedHeight   dbPrefix = "last_signed_height/"
	prefixNodeAccount        dbPrefix = "node_account/"
	prefixActiveObserver     dbPrefix = "active_observer/"
	prefixPoolAddresses      dbPrefix = "pool_addresses/"
	prefixSupportedChains    dbPrefix = "supported_chains/"
	prefixVaultPool          dbPrefix = "vault/"
	prefixVaultData          dbPrefix = "vault_data/"
	prefixObservingAddresses dbPrefix = "observing_addresses/"
	prefixReserves           dbPrefix = "reserves/"
	prefixTss                dbPrefix = "tss/"
	prefixKeygen             dbPrefix = "keygen/"
)

func dbError(ctx sdk.Context, wrapper string, err error) error {
	err = errors.Wrap(err, fmt.Sprintf("KVStore Error: %s", wrapper))
	ctx.Logger().Error(err.Error())
	return err
}

// Keeper maintains the link to data storage and exposes getter/setter methods for the various parts of the state machine
type KVStore struct {
	coinKeeper   bank.Keeper
	supplyKeeper supply.Keeper
	storeKey     sdk.StoreKey // Unexposed key to access store from sdk.Context
	cdc          *codec.Codec // The wire codec for binary encoding/decoding.
}

// NewKVStore creates new instances of the thorchain Keeper
func NewKVStore(coinKeeper bank.Keeper, supplyKeeper supply.Keeper, storeKey sdk.StoreKey, cdc *codec.Codec) KVStore {
	return KVStore{
		coinKeeper:   coinKeeper,
		supplyKeeper: supplyKeeper,
		storeKey:     storeKey,
		cdc:          cdc,
	}
}

func (k KVStore) Cdc() *codec.Codec {
	return k.cdc
}

func (k KVStore) Supply() supply.Keeper {
	return k.supplyKeeper
}

func (k KVStore) CoinKeeper() bank.Keeper {
	return k.coinKeeper
}

func (k KVStore) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", ModuleName))
}

func (k KVStore) GetKey(ctx sdk.Context, prefix dbPrefix, key string) string {
	version := getVersion(k.GetLowestActiveVersion(ctx), prefix)
	return fmt.Sprintf("%s%d/%s", prefix, version.Minor, strings.ToUpper(key))
}
