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
	Stakers          []Staker              `json:"stakers"`
	ObservedTxVoters ObservedTxVoters      `json:"observed_tx_voters"`
	TxOuts           []TxOut               `json:"txouts"`
	NodeAccounts     NodeAccounts          `json:"node_accounts"`
	CurrentEventID   int64                 `json:"current_event_id"`
	Events           Events                `json:"events"`
	Vaults           Vaults                `json:"vaults"`
	Gas              map[string][]sdk.Uint `json:"gas"`
	Reserve          uint64                `json:"reserve"`
}

// NewGenesisState create a new instance of GenesisState
func NewGenesisState() GenesisState {
	return GenesisState{}
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
		Stakers:          make([]Staker, 0),
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

	for _, stake := range data.Stakers {
		keeper.SetStaker(ctx, stake)
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

	if common.RuneAsset().Chain.Equals(common.THORChain) {
		// Mint coins into the reserve
		coin, err := common.NewCoin(common.RuneNative, sdk.NewUint(data.Reserve)).Native()
		if err != nil {
			panic(err)
		}
		coins := sdk.NewCoins(coin)
		if err := keeper.Supply().MintCoins(ctx, ModuleName, coins); err != nil {
			panic(err)
		}
		if err := keeper.Supply().SendCoinsFromModuleToModule(ctx, ModuleName, ReserveName, coins); err != nil {
			panic(err)
		}
	}

	// give mimir gas
	coinsToMint, err := sdk.ParseCoins("1000thor")
	if err != nil {
		panic(err)
	}
	// mint some gas asset
	err = keeper.Supply().MintCoins(ctx, ModuleName, coinsToMint)
	if err != nil {
		panic(err)
	}
	if err := keeper.Supply().SendCoinsFromModuleToAccount(ctx, ModuleName, ADMIN, coinsToMint); err != nil {
		panic(err)
	}

	return validators
}

// ExportGenesis export the data in Genesis
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	var iterator sdk.Iterator
	currentEventID, _ := k.GetCurrentEventID(ctx)

	pools, err := k.GetPools(ctx)
	if err != nil {
		panic(err)
	}

	var stakers []Staker
	for _, pool := range pools {
		iterator := k.GetStakerIterator(ctx, pool.Asset)
		defer iterator.Close()
		for ; iterator.Valid(); iterator.Next() {
			var ps Staker
			k.Cdc().MustUnmarshalBinaryBare(iterator.Value(), &ps)
			stakers = append(stakers, ps)
		}
	}

	var nodeAccounts NodeAccounts
	iterator = k.GetNodeAccountIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var na NodeAccount
		k.Cdc().MustUnmarshalBinaryBare(iterator.Value(), &na)
		nodeAccounts = append(nodeAccounts, na)
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
		Stakers:          stakers,
		ObservedTxVoters: votes,
		TxOuts:           outs,
		CurrentEventID:   currentEventID,
		Events:           events,
		Gas:              gas,
	}
}
