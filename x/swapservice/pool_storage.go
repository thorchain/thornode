package swapservice

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

// PoolStorage allow us to access the pool struct from key values store
// this is an interface thus we could write unit tests
type poolStorage interface {
	PoolExist(ctx sdk.Context, ticker string) bool

	GetPoolStruct(ctx sdk.Context, ticker string) PoolStruct
	SetPoolStruct(ctx sdk.Context, ticker string, ps PoolStruct)

	GetStakerPool(ctx sdk.Context, stakerID string) (types.StakerPool, error)
	SetStakerPool(ctx sdk.Context, stakerID string, sp types.StakerPool)

	GetPoolStaker(ctx sdk.Context, ticker string) (types.PoolStaker, error)
	SetPoolStaker(ctx sdk.Context, ticker string, ps types.PoolStaker)

	SetSwapRecord(ctx sdk.Context, sr types.SwapRecord) error
	GetSwapRecord(ctx sdk.Context, requestTxHash string) (types.SwapRecord, error)
	UpdateSwapRecordPayTxHash(ctx sdk.Context, requestTxHash, payTxHash string) error

	SetUnStakeRecord(ctx sdk.Context, ur types.UnstakeRecord)
	GetUnStakeRecord(ctx sdk.Context, requestTxHash string) (types.UnstakeRecord, error)
	UpdateUnStakeRecordCompleteTxHash(ctx sdk.Context, requestTxHash, completeTxHash string) error
}
