package swapservice

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

type GenesisState struct {
	PoolStructRecords []PoolStruct `json:"poolstruct_records"`
	AccStructRecords  []AccStruct  `json:"accstruct_records"`
}

func NewGenesisState(pools []PoolStruct, accs []AccStruct) GenesisState {
	return GenesisState{
		PoolStructRecords: pools,
		AccStructRecords:  accs,
	}
}

func ValidateGenesis(data GenesisState) error {
	for _, record := range data.PoolStructRecords {
		if record.TokenName == "" {
			return fmt.Errorf("Invalid PoolStructRecord: Value: %s. Error: Missing Token Name", record.TokenName)
		}
		if record.Ticker == "" {
			return fmt.Errorf("Invalid PoolStructRecord: Owner: %s. Error: Missing Ticker", record.Ticker)
		}
	}
	for _, record := range data.AccStructRecords {
		if record.Name == "" {
			return fmt.Errorf("Invalid AccStructRecord: Name: %s. Error: Missing Name", record.Name)
		}
	}
	return nil
}

func DefaultGenesisState() GenesisState {
	return GenesisState{
		PoolStructRecords: []PoolStruct{},
		AccStructRecords:  []AccStruct{},
	}
}

func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) []abci.ValidatorUpdate {
	for _, record := range data.PoolStructRecords {
		keeper.SetPoolStruct(ctx, fmt.Sprintf("pool-%s", strings.ToUpper(record.Ticker)), record)
	}
	for _, record := range data.AccStructRecords {
		keeper.SetAccStruct(ctx, fmt.Sprintf("acct-%s", strings.ToLower(record.Name)), record)
	}
	return []abci.ValidatorUpdate{}
}

func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	var accRecords []AccStruct
	var poolRecords []PoolStruct
	iterator := k.GetPoolDatasIterator(ctx)
	for ; iterator.Valid(); iterator.Next() {
		key := string(iterator.Key())
		if strings.HasPrefix("pool-", key) {
			var poolstruct PoolStruct
			poolstruct = k.GetPoolStruct(ctx, key)
			poolRecords = append(poolRecords, poolstruct)
		} else if strings.HasPrefix("acc-", key) {
			var accstruct AccStruct
			accstruct = k.GetAccStruct(ctx, key)
			accRecords = append(accRecords, accstruct)
		}
	}
	return GenesisState{
		PoolStructRecords: poolRecords,
		AccStructRecords:  accRecords,
	}
}
