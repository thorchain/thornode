package swapservice

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"gitlab.com/thorchain/bepswap/common"
)

// GenesisState strcture that used to store the data we put in genesis
type GenesisState struct {
	PoolRecords   []Pool         `json:"pool_records"`
	TrustAccounts []TrustAccount `json:"trust_accounts"`
	AdminConfigs  []AdminConfig  `json:"admin_configs"`
}

// NewGenesisState create a new instance of GenesisState
func NewGenesisState(pools []Pool, trustAccounts []TrustAccount, configs []AdminConfig) GenesisState {
	return GenesisState{
		PoolRecords:   pools,
		TrustAccounts: trustAccounts,
		AdminConfigs:  configs,
	}
}

// ValidateGenesis validate genesis is valid or not
func ValidateGenesis(data GenesisState) error {
	for _, record := range data.PoolRecords {
		if len(record.Ticker) == 0 {
			return fmt.Errorf("invalid Pool, error: missing ticker")
		}
	}

	for _, config := range data.AdminConfigs {
		if err := config.Valid(); err != nil {
			return err
		}
	}

	for _, ta := range data.TrustAccounts {
		if ta.AdminAddress.IsEmpty() {
			return fmt.Errorf("invalid trust account record, error: missing admin address")
		}
		if ta.SignerAddress.IsEmpty() {
			return fmt.Errorf("invalid trust account record, error: missing signer address")
		}
		if ta.ObserverAddress.Empty() {
			return fmt.Errorf("invalid trust account record, error: missing observer address")
		}
	}
	return nil
}

// DefaultGenesisState the default values we put in the Genesis
func DefaultGenesisState() GenesisState {
	return GenesisState{
		AdminConfigs:  []AdminConfig{},
		PoolRecords:   []Pool{},
		TrustAccounts: []TrustAccount{},
	}
}

// InitGenesis read the data in GenesisState and apply it to data store
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) []abci.ValidatorUpdate {
	for _, record := range data.PoolRecords {
		keeper.SetPool(ctx, record)
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
	var poolRecords []Pool
	iterator := k.GetPoolDataIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var pool Pool
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &pool)
		pool.PoolAddress = k.GetAdminConfigPoolAddress(ctx, common.NoBnbAddress)
		pool.ExpiryUtc = k.GetAdminConfigPoolExpiry(ctx, common.NoBnbAddress)
		poolRecords = append(poolRecords, pool)
	}

	return GenesisState{
		PoolRecords:   poolRecords,
		TrustAccounts: trustAccounts,
	}
}
