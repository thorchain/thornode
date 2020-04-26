package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"gitlab.com/thorchain/thornode/common"
)

// GenesisState strcture that used to store the data THORNode put in genesis
type GenesisState struct {
	Pools            []Pool                `json:"pools"`
	PoolStakers      []PoolStaker          `json:"pool_stakers"`
	ObservedTxVoters ObservedTxVoters      `json:"observed_tx_voters"`
	TxOuts           []TxOut               `json:"txouts"`
	NodeAccounts     NodeAccounts          `json:"node_accounts"`
	CurrentEventID   int64                 `json:"current_event_id"`
	Events           Events                `json:"events"`
	Vaults           Vaults                `json:"vaults"`
	Gas              map[string][]sdk.Uint `json:"gas"`
}

// NewGenesisState create a new instance of GenesisState
func NewGenesisState(pools []Pool, nodeAccounts NodeAccounts) GenesisState {
	return GenesisState{
		Pools:        pools,
		NodeAccounts: nodeAccounts,
	}
}

// ValidateGenesis validate genesis is valid or not
func ValidateGenesis(data GenesisState) error {
	for _, record := range data.Pools {
		if err := record.Valid(); err != nil {
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

	for k, v := range data.Gas {
		if len(v) == 0 {
			return fmt.Errorf("Gas %s cannot have empty units", k)
		}
	}

	return nil
}

// DefaultGenesisState the default values THORNode put in the Genesis
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Pools:            []Pool{},
		NodeAccounts:     NodeAccounts{},
		CurrentEventID:   1,
		TxOuts:           make([]TxOut, 0),
		PoolStakers:      make([]PoolStaker, 0),
		Events:           make(Events, 0),
		Vaults:           make(Vaults, 0),
		ObservedTxVoters: make(ObservedTxVoters, 0),
		Gas:              make(map[string][]sdk.Uint, 0),
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

	validators := make([]abci.ValidatorUpdate, 0, len(data.NodeAccounts))
	for _, nodeAccount := range data.NodeAccounts {
		if nodeAccount.Status == NodeActive {
			// Only Active node will become validator
			pk, err := sdk.GetConsPubKeyBech32(nodeAccount.ValidatorConsPubKey)
			if err != nil {
				ctx.Logger().Error("fail to parse consensus public key", "key", nodeAccount.ValidatorConsPubKey, "error", err)
				panic(err)
			}
			validators = append(validators, abci.ValidatorUpdate{
				PubKey: tmtypes.TM2PB.PubKey(pk),
				Power:  100,
			})
		}

		if err := keeper.SetNodeAccount(ctx, nodeAccount); err != nil {
			// we should panic
			panic(err)
		}
	}

	for _, vault := range data.Vaults {
		if err := keeper.SetVault(ctx, vault); err != nil {
			panic(err)
		}
	}

	for _, voter := range data.ObservedTxVoters {
		keeper.SetObservedTxVoter(ctx, voter)
	}

	for _, out := range data.TxOuts {
		if err := keeper.SetTxOut(ctx, &out); err != nil {
			ctx.Logger().Error("fail to save tx out during genesis", "error", err)
			panic(err)
		}
	}

	for _, e := range data.Events {
		if err := keeper.UpsertEvent(ctx, e); err != nil {
			panic(err)
		}
	}

	for k, v := range data.Gas {
		asset, err := common.NewAsset(k)
		if err != nil {
			panic(err)
		}
		keeper.SetGas(ctx, asset, v)
	}

	keeper.SetCurrentEventID(ctx, data.CurrentEventID)

	return validators
}

// ExportGenesis export the data in Genesis
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	currentEventID, _ := k.GetCurrentEventID(ctx)

	var poolStakers []PoolStaker
	iterator := k.GetPoolStakerIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var ps PoolStaker
		k.Cdc().MustUnmarshalBinaryBare(iterator.Value(), &ps)
		poolStakers = append(poolStakers, ps)
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

	gas := make(map[string][]sdk.Uint, 0)
	iterator = k.GetGasIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var g []sdk.Uint
		k.Cdc().MustUnmarshalBinaryBare(iterator.Value(), &g)
		gas[string(iterator.Key())] = g
	}

	return GenesisState{
		Pools:            pools,
		NodeAccounts:     nodeAccounts,
		PoolStakers:      poolStakers,
		ObservedTxVoters: votes,
		TxOuts:           outs,
		CurrentEventID:   currentEventID,
		Events:           events,
		Gas:              gas,
	}
}
