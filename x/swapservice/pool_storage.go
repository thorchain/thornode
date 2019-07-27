package swapservice

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PoolStorage allow us to access the pool struct from key values store
// this is an interface thus we could write unit tests
type poolStorage interface {
	PoolExist(ctx sdk.Context, poolID string) bool
	GetPoolStruct(ctx sdk.Context, poolID string) PoolStruct
	SetPoolStruct(ctx sdk.Context, poolID string, ps PoolStruct)
}
