package swapservice

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// GenesisState strcture that used to store the data we put in genesis
type GenesisState struct {
	PoolStructRecords []PoolStruct   `json:"poolstruct_records"`
	TrustAccounts     []TrustAccount `json:"trust_accounts"`
	AdminConfigs      []AdminConfig  `json:"admin_configs"`
}

// NewGenesisState create a new instance of GenesisState
func NewGenesisState(pools []PoolStruct, trustAccounts []TrustAccount, configs []AdminConfig) GenesisState {
	return GenesisState{
		PoolStructRecords: pools,
		TrustAccounts:     trustAccounts,
		AdminConfigs:      configs,
	}
}

// ValidateGenesis validate genesis is valid or not
func ValidateGenesis(data GenesisState) error {
	for _, record := range data.PoolStructRecords {
		if len(record.Ticker) == 0 {
			return fmt.Errorf("invalid PoolStruct, error: missing ticker")
		}
		if len(record.PoolAddress) == 0 {
			return fmt.Errorf("invalid PoolStruct, error: missing pool address")
		}
	}

	for _, record := range data.AdminConfigs {
		if len(record.Key) == 0 {
			return fmt.Errorf("invalid admin config, error: missing key")
		}
		if len(record.Value) == 0 {
			return fmt.Errorf("invalid admin config, error: missing value")
		}
	}

	for _, ta := range data.TrustAccounts {
		if len(ta.Name) == 0 {
			return fmt.Errorf("invalid trust account record, error: missing account name")
		}
		if ta.Address.Empty() {
			return fmt.Errorf("invalid trust account record, error: missing account address")
		}
	}
	for _, ta := range data.TrustAccounts {
		if len(ta.Name) == 0 {
			return fmt.Errorf("invalid trust account record, error: missing account name")
		}
		if ta.Address.Empty() {
			return fmt.Errorf("invalid trust account record, error: missing account address")
		}
	}
	return nil
}

// DefaultGenesisState the default values we put in the Genesis
func DefaultGenesisState() GenesisState {
	return GenesisState{
		AdminConfigs: []AdminConfig{
			{Key: "TSL", Value: "10"},
			{Key: "GSL", Value: "30"},
		},
		PoolStructRecords: []PoolStruct{
			{
				BalanceRune:  "0",
				BalanceToken: "0",
				Ticker:       "BNB",
				PoolUnits:    "0",
				PoolAddress:  "bnbxxdfdfdfdfdf",
				Status:       PoolBootstrap,
			},
		},
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
	for _, ta := range data.AdminConfigs {
		keeper.SetAdminConfig(ctx, ta)
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

	var trustAccounts []TrustAccount
	taIterator := k.GetTrustAccountIterator(ctx)
	defer taIterator.Close()
	for ; taIterator.Valid(); taIterator.Next() {
		var ta TrustAccount
		k.cdc.MustUnmarshalBinaryBare(taIterator.Value(), &ta)
		trustAccounts = append(trustAccounts, ta)
	}

	var configs []AdminConfig
	configIterator := k.GetAdminConfigIterator(ctx)
	defer configIterator.Close()
	for ; configIterator.Valid(); configIterator.Next() {
		var config AdminConfig
		k.cdc.MustUnmarshalBinaryBare(configIterator.Value(), &config)
		configs = append(configs, config)
	}

	return GenesisState{
		PoolStructRecords: poolRecords,
		TrustAccounts:     trustAccounts,
		AdminConfigs:      configs,
	}
}
