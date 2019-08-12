package mocks

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

// MockPoolStorage implements PoolStorage interface, thus we can mock the error cases
type MockPoolStorage struct {
}

func (mps MockPoolStorage) PoolExist(ctx sdk.Context, ticker types.Ticker) bool {
	if ticker.Equals(types.Ticker("NOTEXIST")) {
		return false
	}
	return true
}

func (mps MockPoolStorage) GetPoolStruct(ctx sdk.Context, ticker types.Ticker) types.PoolStruct {
	if ticker.Equals(types.Ticker("NOTEXIST")) {
		return types.PoolStruct{}
	} else {
		return types.PoolStruct{
			BalanceRune:  "100",
			BalanceToken: "100",
			Status:       types.Enabled,
			Ticker:       ticker,
		}
	}
}
func (mps MockPoolStorage) SetPoolStruct(ctx sdk.Context, ticker types.Ticker, ps types.PoolStruct) {

}

func (mps MockPoolStorage) SetSwapRecord(ctx sdk.Context, sr types.SwapRecord) error {
	if sr.RequestTxHash.Equals(types.TxID("ASKFORERROR")) {
		return errors.New("you asked for it")
	}
	return nil
}

func (mps MockPoolStorage) GetSwapRecord(ctx sdk.Context, requestTxHash types.TxID) (types.SwapRecord, error) {
	if requestTxHash.Equals(types.TxID("ASKFORERROR")) {
		return types.SwapRecord{}, errors.New("you asked for it")
	}
	return types.SwapRecord{}, nil
}

func (mps MockPoolStorage) UpdateSwapRecordPayTxHash(ctx sdk.Context, requestTxHash, payTxHash types.TxID) error {
	if requestTxHash.Equals(types.TxID("ASKFORERROR")) {
		return errors.New("you ask for it")
	}
	return nil
}

func (mps MockPoolStorage) SetUnStakeRecord(ctx sdk.Context, ur types.UnstakeRecord) {}

func (mps MockPoolStorage) GetUnStakeRecord(ctx sdk.Context, requestTxHash types.TxID) (types.UnstakeRecord, error) {
	if requestTxHash.Equals(types.TxID("ASKFORERROR")) {
		return types.UnstakeRecord{}, errors.New("you asked for it")
	}
	return types.UnstakeRecord{}, nil
}

func (mps MockPoolStorage) UpdateUnStakeRecordCompleteTxHash(ctx sdk.Context, requestTxHash, completeTxHash types.TxID) error {
	if requestTxHash.Equals(types.TxID("ASKFORERROR")) {
		return errors.New("you ask for it")
	}
	return nil
}

func (mps MockPoolStorage) GetAdminConfig(ctx sdk.Context, key types.AdminConfigKey) types.AdminConfig {
	return types.NewAdminConfig(key, "FOOBAR")
}
