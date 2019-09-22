package swapservice

import (
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	"gitlab.com/thorchain/bepswap/common"
)

type txIndex struct {
	Height uint64    `json:"height"`
	Index  TxInIndex `json:"index"`
}

// GenesisState strcture that used to store the data we put in genesis
type GenesisState struct {
	Pools            []Pool         `json:"pools"`
	PoolIndex        PoolIndex      `json:"pool_index"`
	PoolStakers      []PoolStaker   `json:"pool_stakers"`
	StakerPools      []StakerPool   `json:"staker_pools"`
	TxInVoters       []TxInVoter    `json:"txin_voters"`
	TxInIndexes      []txIndex      `json:"txin_indexes"`
	TxOuts           []TxOut        `json:"txouts"`
	CompleteEvents   Events         `json:"complete_events"`
	IncompleteEvents Events         `json:"incomplete_events"`
	TrustAccounts    []TrustAccount `json:"trust_accounts"`
	AdminConfigs     []AdminConfig  `json:"admin_configs"`
	LastEventID      common.Amount  `json:"last_event_id"`
}

// NewGenesisState create a new instance of GenesisState
func NewGenesisState(pools []Pool, trustAccounts []TrustAccount, configs []AdminConfig) GenesisState {
	return GenesisState{
		Pools:         pools,
		TrustAccounts: trustAccounts,
		AdminConfigs:  configs,
	}
}

// ValidateGenesis validate genesis is valid or not
func ValidateGenesis(data GenesisState) error {
	for _, record := range data.Pools {
		if err := record.Valid(); err != nil {
			return err
		}
	}

	for _, stake := range data.StakerPools {
		if err := stake.Valid(); err != nil {
			return err
		}
	}

	for _, voter := range data.TxInVoters {
		if err := voter.Valid(); err != nil {
			return err
		}
	}

	for _, index := range data.TxInIndexes {
		if index.Height == 0 {
			return errors.New("Tx In Index cannot have a height of zero")
		}
	}

	for _, out := range data.TxOuts {
		if err := out.Valid(); err != nil {
			return err
		}
	}

	for _, config := range data.AdminConfigs {
		if err := config.Valid(); err != nil {
			return err
		}
	}

	for _, ta := range data.TrustAccounts {
		if err := ta.IsValid(); err != nil {
			return err
		}
	}

	if data.LastEventID.IsEmpty() {
		return errors.New("Missing last event ID")
	}

	return nil
}

// DefaultGenesisState the default values we put in the Genesis
func DefaultGenesisState() GenesisState {
	return GenesisState{
		AdminConfigs:  []AdminConfig{},
		Pools:         []Pool{},
		TrustAccounts: []TrustAccount{},
		LastEventID:   common.ZeroAmount,
	}
}

// InitGenesis read the data in GenesisState and apply it to data store
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) []abci.ValidatorUpdate {
	for _, record := range data.Pools {
		keeper.SetPool(ctx, record)
	}

	if data.PoolIndex != nil {
		keeper.SetPoolIndex(ctx, data.PoolIndex)
	}

	for _, stake := range data.PoolStakers {
		keeper.SetPoolStaker(ctx, stake.Ticker, stake)
	}

	for _, config := range data.AdminConfigs {
		keeper.SetAdminConfig(ctx, config)
	}

	for _, ta := range data.TrustAccounts {
		keeper.SetTrustAccount(ctx, ta)
	}

	for _, stake := range data.StakerPools {
		keeper.SetStakerPool(ctx, stake.StakerID, stake)
	}

	for _, voter := range data.TxInVoters {
		keeper.SetTxInVoter(ctx, voter)
	}

	for _, out := range data.TxOuts {
		keeper.SetTxOut(ctx, &out)
	}

	for _, index := range data.TxInIndexes {
		keeper.SetTxInIndex(ctx, index.Height, index.Index)
	}

	keeper.SetIncompleteEvents(ctx, data.IncompleteEvents)

	for _, event := range data.CompleteEvents {
		keeper.SetCompletedEvent(ctx, event)
	}

	keeper.SetLastEventID(ctx, data.LastEventID)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis export the data in Genesis
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	lastEventID := k.GetLastEventID(ctx)

	var adminConfigs []AdminConfig
	iterator := k.GetAdminConfigIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var config AdminConfig
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &config)
		adminConfigs = append(adminConfigs, config)
	}

	poolIndex, _ := k.GetPoolIndex(ctx)

	var poolStakers []PoolStaker
	iterator = k.GetPoolStakerIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var ps PoolStaker
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &ps)
		poolStakers = append(poolStakers, ps)
	}

	var stakerPools []StakerPool
	iterator = k.GetStakerPoolIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var sp StakerPool
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &sp)
		stakerPools = append(stakerPools, sp)
	}

	var trustAccounts []TrustAccount
	iterator = k.GetTrustAccountIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var ta TrustAccount
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &ta)
		trustAccounts = append(trustAccounts, ta)
	}

	var poolRecords []Pool
	iterator = k.GetPoolDataIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var pool Pool
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &pool)
		pool.PoolAddress = k.GetAdminConfigPoolAddress(ctx, common.NoBnbAddress)
		pool.ExpiryUtc = k.GetAdminConfigPoolExpiry(ctx, common.NoBnbAddress)
		poolRecords = append(poolRecords, pool)
	}

	var votes []TxInVoter
	iterator = k.GetTxInVoterIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var vote TxInVoter
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &vote)
		votes = append(votes, vote)
	}

	var indexes []txIndex
	iterator = k.GetTxInIndexIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var index txIndex
		index.Height, _ = strconv.ParseUint(
			strings.TrimLeft(string(iterator.Key()), string(prefixTxInIndex)),
			10,
			64,
		)
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &index.Index)
		indexes = append(indexes, index)
	}

	var outs []TxOut
	iterator = k.GetTxOutIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var out TxOut
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &out)
		outs = append(outs, out)
	}

	var completed []Event
	iterator = k.GetCompleteEventIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var e Event
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &e)
		completed = append(completed, e)
	}

	incomplete, _ := k.GetIncompleteEvents(ctx)

	return GenesisState{
		Pools:            poolRecords,
		PoolIndex:        poolIndex,
		TrustAccounts:    trustAccounts,
		AdminConfigs:     adminConfigs,
		LastEventID:      lastEventID,
		PoolStakers:      poolStakers,
		StakerPools:      stakerPools,
		TxInVoters:       votes,
		TxInIndexes:      indexes,
		TxOuts:           outs,
		CompleteEvents:   completed,
		IncompleteEvents: incomplete,
	}
}
