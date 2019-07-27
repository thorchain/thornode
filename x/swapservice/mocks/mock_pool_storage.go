package mocks

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

// MockPoolStorage implements PoolStorage interface, thus we can mock the error cases
type MockPoolStorage struct {
}

func (mps MockPoolStorage) PoolExist(ctx sdk.Context, poolID string) bool {
	if strings.EqualFold(poolID, types.PoolDataKeyPrefix+"NOTEXIST") {
		return false
	}
	return true
}
func (mps MockPoolStorage) GetPoolStruct(ctx sdk.Context, poolID string) types.PoolStruct {
	if strings.EqualFold(poolID, types.PoolDataKeyPrefix+"NOTEXIST") {
		return types.PoolStruct{}
	} else {
		return types.PoolStruct{
			BalanceRune:  "100",
			BalanceToken: "100",
			PoolID:       poolID,
			PoolUnits:    "100",
			Status:       types.Active.String(),
			PoolAddress:  "hello",
			Ticker:       strings.TrimPrefix(types.PoolDataKeyPrefix, poolID),
			TokenName:    "BNB",
		}
	}
}
func (mps MockPoolStorage) SetPoolStruct(ctx sdk.Context, poolID string, ps types.PoolStruct) {

}
