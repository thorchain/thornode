package swapservice

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

// PoolStorage allow us to access the pool struct from key values store
// this is an interface thus we could write unit tests
type poolStorage interface {
	PoolExist(ctx sdk.Context, poolID string) bool

	GetPoolStruct(ctx sdk.Context, poolID string) PoolStruct
	SetPoolStruct(ctx sdk.Context, poolID string, ps PoolStruct)

	GetStakerPool(ctx sdk.Context, stakerID string) (types.StakerPool, error)
	SetStakerPool(ctx sdk.Context, stakerID string, sp types.StakerPool)

	GetPoolStaker(ctx sdk.Context, poolID string) (types.PoolStaker, error)
	SetPoolStaker(ctx sdk.Context, poolID string, ps types.PoolStaker)

	SetSwapRecord(ctx sdk.Context, sr SwapRecord) error
	GetSwapRecord(ctx sdk.Context, requestTxHash string) (SwapRecord, error)
	UpdateSwapRecordPayTxHash(ctx sdk.Context, requestTxHash, payTxHash string) error
}
