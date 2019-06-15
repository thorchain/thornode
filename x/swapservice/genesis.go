package swapservice

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

type GenesisState struct {
	PoolStructRecords []PoolStruct `json:"poolstruct_records"`
}

func NewGenesisState(whoIsRecords []PoolStruct) GenesisState {
	return GenesisState{PoolStructRecords: nil}
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
	return nil
}

func DefaultGenesisState() GenesisState {
	return GenesisState{
		PoolStructRecords: []PoolStruct{},
	}
}

func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) []abci.ValidatorUpdate {
	// TODO: INIT GENESIS
	return []abci.ValidatorUpdate{}
}

func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	var records []PoolStruct
	iterator := k.GetPoolDatasIterator(ctx)
	for ; iterator.Valid(); iterator.Next() {
		pooldata := string(iterator.Key())
		var poolstruct PoolStruct
		poolstruct = k.GetPoolStruct(ctx, pooldata)
		records = append(records, poolstruct)
	}
	return GenesisState{PoolStructRecords: records}
}
