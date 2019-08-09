package swapservice

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PoolStorage allow us to access the pool struct from key values store
// this is an interface thus we could write unit tests
type poolStorage interface {
	PoolExist(ctx sdk.Context, ticker Ticker) bool

	GetPoolStruct(ctx sdk.Context, ticker Ticker) PoolStruct
	SetPoolStruct(ctx sdk.Context, ticker Ticker, ps PoolStruct)

	GetStakerPool(ctx sdk.Context, stakerID string) (StakerPool, error)
	SetStakerPool(ctx sdk.Context, stakerID string, sp StakerPool)

	GetPoolStaker(ctx sdk.Context, ticker Ticker) (PoolStaker, error)
	SetPoolStaker(ctx sdk.Context, ticker Ticker, ps PoolStaker)

	SetSwapRecord(ctx sdk.Context, sr SwapRecord) error
	GetSwapRecord(ctx sdk.Context, requestTxHash TxID) (SwapRecord, error)
	UpdateSwapRecordPayTxHash(ctx sdk.Context, requestTxHash, payTxHash TxID) error

	GetAdminConfig(ctx sdk.Context, key string) AdminConfig

	SetUnStakeRecord(ctx sdk.Context, ur UnstakeRecord)
	GetUnStakeRecord(ctx sdk.Context, requestTxHash TxID) (UnstakeRecord, error)
	UpdateUnStakeRecordCompleteTxHash(ctx sdk.Context, requestTxHash, completeTxHash TxID) error
}
