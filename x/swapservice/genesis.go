package swapservice

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

// GenesisState strcture that used to store the data we put in genesis
type GenesisState struct {
	PoolStructRecords []PoolStruct         `json:"poolstruct_records"`
	TrustAccounts     []types.TrustAccount `json:"trust_accounts"`
}

// NewGenesisState create a new instance of GenesisState
func NewGenesisState(pools []PoolStruct, trustAccounts []types.TrustAccount) GenesisState {
	return GenesisState{
		PoolStructRecords: pools,
		TrustAccounts:     trustAccounts,
	}
}

// ValidateGenesis validate genesis is valid or not
func ValidateGenesis(data GenesisState) error {
	for _, record := range data.PoolStructRecords {
		if len(record.TokenName) == 0 {
			return fmt.Errorf("invalid PoolStruct, error: missing token name")
		}
		if len(record.Ticker) == 0 {
			return fmt.Errorf("invalid PoolStruct, error: missing ticker")
		}
		if len(record.PoolAddress) == 0 {
			return fmt.Errorf("invalid PoolStruct, error: missing pool address")
		}
	}
	for _, ta := range data.TrustAccounts {
		if len(ta.Name) == 0 {
			return fmt.Errorf("invalid trust account record, error: missing account name")
		}
		if ta.Address.Empty() {
			return fmt.Errorf("invalid trust account record, error: missing account address: %s", ta.Address)
		}
	}
	return nil
}

// DefaultGenesisState the default values we put in the Genesis
func DefaultGenesisState() GenesisState {
	return GenesisState{
		PoolStructRecords: []PoolStruct{
			{
				BalanceRune:  "0",
				BalanceToken: "0",
				TokenName:    "Binance Coin",
				Ticker:       "BNB",
				PoolUnits:    "0",
				PoolAddress:  "bnbxxdfdfdfdfdf",
				Status:       types.Active.String(),
			},
		},
		TrustAccounts: []types.TrustAccount{},
	}
}

// InitGenesis read the data in GenesisState and apply it to data store
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) []abci.ValidatorUpdate {
	for _, record := range data.PoolStructRecords {
		keeper.SetPoolStruct(ctx, record.Ticker, record)
	}
	for _, ta := range data.TrustAccounts {
		keeper.SetTrustAccount(ctx, ta)
	}
	return []abci.ValidatorUpdate{}
}

// ExportGenesis export the data in Genesis
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	var poolRecords []PoolStruct
	iterator := k.GetPoolStructDataIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var poolstruct PoolStruct
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &poolstruct)
		poolRecords = append(poolRecords, poolstruct)
	}
	var trustAccounts []types.TrustAccount
	taIterator := k.GetTrustAccountIterator(ctx)
	defer taIterator.Close()
	for ; taIterator.Valid(); taIterator.Next() {
		var ta types.TrustAccount
		k.cdc.MustUnmarshalBinaryBare(taIterator.Value(), &ta)
		trustAccounts = append(trustAccounts, ta)
	}
	return GenesisState{
		PoolStructRecords: poolRecords,
		TrustAccounts:     trustAccounts,
	}
}
