package mocks

import (
	"errors"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

// MockPoolStorage implements PoolStorage interface, thus we can mock the error cases
type MockPoolStorage struct {
}

func (mps MockPoolStorage) PoolExist(ctx sdk.Context, ticker string) bool {
	if strings.EqualFold(ticker, "NOTEXIST") {
		return false
	}
	return true
}
func (mps MockPoolStorage) GetPoolStruct(ctx sdk.Context, ticker string) types.PoolStruct {
	if strings.EqualFold(ticker, "NOTEXIST") {
		return types.PoolStruct{}
	} else {
		return types.PoolStruct{
			BalanceRune:  "100",
			BalanceToken: "100",
			PoolUnits:    "100",
			Status:       types.Active.String(),
			PoolAddress:  "hello",
			Ticker:       ticker,
			TokenName:    "BNB",
		}
	}
}
func (mps MockPoolStorage) SetPoolStruct(ctx sdk.Context, ticker string, ps types.PoolStruct) {

}
func (mps MockPoolStorage) GetStakerPool(ctx sdk.Context, stakerID string) (types.StakerPool, error) {
	if strings.EqualFold(stakerID, "NOTEXISTSTAKER") {
		return types.StakerPool{}, errors.New("you asked for it")
	}
	return types.NewStakerPool(stakerID), nil
}
func (mps MockPoolStorage) SetStakerPool(ctx sdk.Context, stakerID string, sp types.StakerPool) {

}

func (mps MockPoolStorage) GetPoolStaker(ctx sdk.Context, ticker string) (types.PoolStaker, error) {
	if strings.EqualFold(ticker, "NOTEXISTSTICKER") {
		return types.PoolStaker{}, errors.New("you asked for it")
	}
	return types.NewPoolStaker(ticker, "100"), nil
}
func (mps MockPoolStorage) SetPoolStaker(ctx sdk.Context, ticker string, ps types.PoolStaker) {

}
func (mps MockPoolStorage) SetSwapRecord(ctx sdk.Context, sr types.SwapRecord) error {
	if strings.EqualFold(sr.RequestTxHash, "ASKFORERROR") {
		return errors.New("you asked for it")
	}
	return nil
}
func (mps MockPoolStorage) GetSwapRecord(ctx sdk.Context, requestTxHash string) (types.SwapRecord, error) {
	if strings.EqualFold(requestTxHash, "ASKFORERROR") {
		return types.SwapRecord{}, errors.New("you asked for it")
	}
	return types.SwapRecord{}, nil
}
func (mps MockPoolStorage) UpdateSwapRecordPayTxHash(ctx sdk.Context, requestTxHash, payTxHash string) error {
	if strings.EqualFold(requestTxHash, "ASKFORERROR") {
		return errors.New("you ask for it")
	}
	return nil
}
func (mps MockPoolStorage) SetUnStakeRecord(ctx sdk.Context, ur types.UnstakeRecord) {
}
func (mps MockPoolStorage) GetUnStakeRecord(ctx sdk.Context, requestTxHash string) (types.UnstakeRecord, error) {
	if strings.EqualFold(requestTxHash, "ASKFORERROR") {
		return types.UnstakeRecord{}, errors.New("you asked for it")
	}
	return types.UnstakeRecord{}, nil
}
func (mps MockPoolStorage) UpdateUnStakeRecordCompleteTxHash(ctx sdk.Context, requestTxHash, completeTxHash string) error {
	if strings.EqualFold(requestTxHash, "ASKFORERROR") {
		return errors.New("you ask for it")
	}
	return nil
}
