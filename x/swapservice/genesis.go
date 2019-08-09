package swapservice

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/jpthor/cosmos-swap/x/swapservice/types"
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
	}

	for _, config := range data.AdminConfigs {
		if err := config.Valid(); err != nil {
			return err
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
	// TODO: make hard coded address dynamic for integration testing
	// To get your address for jack `sscli key show jack -a`
	addr, _ := sdk.AccAddressFromBech32("rune1gnaghgzcpd73hcxeturml96maa0fajg9t8m0yj")
	return GenesisState{
		AdminConfigs: []AdminConfig{
			{
				Key:   PoolAddressKey,
				Value: "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6",
			},
		},
		PoolStructRecords: []PoolStruct{
			{
				BalanceRune:  "0",
				BalanceToken: "0",
				Ticker:       "BNB",
				PoolUnits:    "0",
				Status:       PoolBootstrap,
			},
		},
		TrustAccounts: []TrustAccount{
			{Name: "Jack", Address: addr},
		},
	}
}

// InitGenesis read the data in GenesisState and apply it to data store
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) []abci.ValidatorUpdate {
	for _, record := range data.PoolStructRecords {
		keeper.SetPoolStruct(ctx, record.Ticker, record)
	}

	for _, config := range data.AdminConfigs {
		keeper.SetAdminConfig(ctx, config)
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

	var adminConfigs []AdminConfig
	configIterator := k.GetAdminConfigIterator(ctx)
	defer configIterator.Close()
	for ; configIterator.Valid(); configIterator.Next() {
		var config AdminConfig
		k.cdc.MustUnmarshalBinaryBare(configIterator.Value(), &config)
		adminConfigs = append(adminConfigs, config)
	}

	var trustAccounts []TrustAccount
	taIterator := k.GetTrustAccountIterator(ctx)
	defer taIterator.Close()
	for ; taIterator.Valid(); taIterator.Next() {
		var ta TrustAccount
		k.cdc.MustUnmarshalBinaryBare(taIterator.Value(), &ta)
		trustAccounts = append(trustAccounts, ta)
	}

	return GenesisState{
		PoolStructRecords: poolRecords,
		TrustAccounts:     trustAccounts,
	}
}
