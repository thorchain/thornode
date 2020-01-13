package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// GenesisState strcture that used to store the data THORNode put in genesis
type GenesisState struct {
	Pools            []Pool           `json:"pools"`
	PoolStakers      []PoolStaker     `json:"pool_stakers"`
	StakerPools      []StakerPool     `json:"staker_pools"`
	ObservedTxVoters ObservedTxVoters `json:"observed_tx_voters"`
	TxOuts           []TxOut          `json:"txouts"`
	NodeAccounts     NodeAccounts     `json:"node_accounts"`
	AdminConfigs     []AdminConfig    `json:"admin_configs"`
	CurrentEventID   int64            `json:"current_event_id"`
	Events           Events           `json:"events"`
	Vaults           Vaults           `json:"vaults"`
}

// NewGenesisState create a new instance of GenesisState
func NewGenesisState(pools []Pool, nodeAccounts NodeAccounts, configs []AdminConfig) GenesisState {
	return GenesisState{
		Pools:        pools,
		NodeAccounts: nodeAccounts,
		AdminConfigs: configs,
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

	for _, voter := range data.ObservedTxVoters {
		if err := voter.Valid(); err != nil {
			return err
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

	for _, ta := range data.NodeAccounts {
		if err := ta.IsValid(); err != nil {
			return err
		}
	}

	for _, vault := range data.Vaults {
		if err := vault.IsValid(); err != nil {
			return err
		}
	}

	return nil
}

// DefaultGenesisState the default values THORNode put in the Genesis
func DefaultGenesisState() GenesisState {
	return GenesisState{
		AdminConfigs:     []AdminConfig{},
		Pools:            []Pool{},
		NodeAccounts:     NodeAccounts{},
		CurrentEventID:   1,
		TxOuts:           make([]TxOut, 0),
		PoolStakers:      make([]PoolStaker, 0),
		StakerPools:      make([]StakerPool, 0),
		Events:           make(Events, 0),
		Vaults:           make(Vaults, 0),
		ObservedTxVoters: make(ObservedTxVoters, 0),
	}
}

// InitGenesis read the data in GenesisState and apply it to data store
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) []abci.ValidatorUpdate {
	for _, record := range data.Pools {
		if err := keeper.SetPool(ctx, record); err != nil {
			panic(err)
		}
	}

	for _, stake := range data.PoolStakers {
		keeper.SetPoolStaker(ctx, stake)
	}

	for _, config := range data.AdminConfigs {
		keeper.SetAdminConfig(ctx, config)
	}

	validators := make([]abci.ValidatorUpdate, 0, len(data.NodeAccounts))
	for _, nodeAccount := range data.NodeAccounts {
		if nodeAccount.Status == NodeActive {
			// Only Active node will become validator
			pk, err := sdk.GetConsPubKeyBech32(nodeAccount.ValidatorConsPubKey)
			if nil != err {
				ctx.Logger().Error("fail to parse consensus public key", "key", nodeAccount.ValidatorConsPubKey, "error", err)
				panic(err)
			}
			validators = append(validators, abci.ValidatorUpdate{
				PubKey: tmtypes.TM2PB.PubKey(pk),
				Power:  100,
			})
		}

		if err := keeper.SetNodeAccount(ctx, nodeAccount); nil != err {
			// we should panic
			panic(err)
		}
	}

	for _, vault := range data.Vaults {
		if err := keeper.SetVault(ctx, vault); err != nil {
			panic(err)
		}
	}

	for _, stake := range data.StakerPools {
		keeper.SetStakerPool(ctx, stake)
	}

	for _, voter := range data.ObservedTxVoters {
		keeper.SetObservedTxVoter(ctx, voter)
	}

	for _, out := range data.TxOuts {
		if err := keeper.SetTxOut(ctx, &out); nil != err {
			ctx.Logger().Error("fail to save tx out during genesis", "error", err)
			panic(err)
		}
	}

	for _, e := range data.Events {
		if err := keeper.UpsertEvent(ctx, e); nil != err {
			panic(err)
		}
	}

	keeper.SetCurrentEventID(ctx, data.CurrentEventID)

	return validators

}

// ExportGenesis export the data in Genesis
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	currentEventID, _ := k.GetCurrentEventID(ctx)

	var adminConfigs []AdminConfig
	iterator := k.GetAdminConfigIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var config AdminConfig
		k.Cdc().MustUnmarshalBinaryBare(iterator.Value(), &config)
		adminConfigs = append(adminConfigs, config)
	}

	var poolStakers []PoolStaker
	iterator = k.GetPoolStakerIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var ps PoolStaker
		k.Cdc().MustUnmarshalBinaryBare(iterator.Value(), &ps)
		poolStakers = append(poolStakers, ps)
	}

	var stakerPools []StakerPool
	iterator = k.GetStakerPoolIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var sp StakerPool
		k.Cdc().MustUnmarshalBinaryBare(iterator.Value(), &sp)
		stakerPools = append(stakerPools, sp)
	}

	var nodeAccounts NodeAccounts
	iterator = k.GetNodeAccountIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var na NodeAccount
		k.Cdc().MustUnmarshalBinaryBare(iterator.Value(), &na)
		nodeAccounts = append(nodeAccounts, na)
	}

	pools, err := k.GetPools(ctx)
	if err != nil {
		panic(err)
	}

	var votes ObservedTxVoters
	iterator = k.GetObservedTxVoterIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var vote ObservedTxVoter
		k.Cdc().MustUnmarshalBinaryBare(iterator.Value(), &vote)
		votes = append(votes, vote)
	}

	var outs []TxOut
	iterator = k.GetTxOutIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var out TxOut
		k.Cdc().MustUnmarshalBinaryBare(iterator.Value(), &out)
		outs = append(outs, out)
	}

	var events []Event
	iterator = k.GetEventsIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var e Event
		k.Cdc().MustUnmarshalBinaryBare(iterator.Value(), &e)
		events = append(events, e)
	}

	return GenesisState{
		Pools:            pools,
		NodeAccounts:     nodeAccounts,
		AdminConfigs:     adminConfigs,
		PoolStakers:      poolStakers,
		StakerPools:      stakerPools,
		ObservedTxVoters: votes,
		TxOuts:           outs,
		CurrentEventID:   currentEventID,
		Events:           events,
	}
}
