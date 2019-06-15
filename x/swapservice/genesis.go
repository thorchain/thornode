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
		if record.Owner == nil {
			return fmt.Errorf("Invalid PoolStructRecord: Value: %s. Error: Missing Owner", record.Value)
		}
		if record.Value == "" {
			return fmt.Errorf("Invalid PoolStructRecord: Owner: %s. Error: Missing Value", record.Owner)
		}
		if record.Price == nil {
			return fmt.Errorf("Invalid PoolStructRecord: Value: %s. Error: Missing Price", record.Value)
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
	for _, record := range data.PoolStructRecords {
		keeper.SetPoolStruct(ctx, record.Value, record)
	}
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
