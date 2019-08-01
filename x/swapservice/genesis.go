package swapservice

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

type GenesisState struct {
	PoolStructRecords []PoolStruct `json:"poolstruct_records"`
}

func NewGenesisState(pools []PoolStruct) GenesisState {
	return GenesisState{
		PoolStructRecords: pools,
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
	return nil
}

func DefaultGenesisState() GenesisState {
	return GenesisState{
		PoolStructRecords: []PoolStruct{
			{
				PoolID:       types.GetPoolNameFromTicker("BNB"),
				BalanceRune:  "0",
				BalanceToken: "0",
				TokenName:    "Binance Coin",
				Ticker:       "BNB",
				PoolUnits:    "0",
				PoolAddress:  "bnbxxdfdfdfdfdf",
				Status:       types.Active.String(),
			},
		},
	}
}

func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) []abci.ValidatorUpdate {
	for _, record := range data.PoolStructRecords {
		keeper.SetPoolStruct(ctx, types.GetPoolNameFromTicker(record.Ticker), record)
	}
	return []abci.ValidatorUpdate{}
}

func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	var poolRecords []PoolStruct
	iterator := k.GetDatasIterator(ctx)
	for ; iterator.Valid(); iterator.Next() {
		key := string(iterator.Key())
		if strings.HasPrefix(types.PoolDataKeyPrefix, key) {
			var poolstruct PoolStruct
			poolstruct = k.GetPoolStruct(ctx, key)
			poolRecords = append(poolRecords, poolstruct)
		}
	}
	return GenesisState{
		PoolStructRecords: poolRecords,
	}
}
