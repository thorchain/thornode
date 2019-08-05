package swapservice

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

type GenesisState struct {
	PoolStructRecords []PoolStruct         `json:"poolstruct_records"`
	TrustAccounts     []types.TrustAccount `json:"trust_accounts"`
}

func NewGenesisState(pools []PoolStruct, trustAccounts []types.TrustAccount) GenesisState {
	return GenesisState{
		PoolStructRecords: pools,
		TrustAccounts:     trustAccounts,
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
	for _, ta := range data.TrustAccounts {
		keeper.SetTrustAccount(ctx, ta)
	}
	return []abci.ValidatorUpdate{}
}

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
