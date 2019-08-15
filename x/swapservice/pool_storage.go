package swapservice

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PoolStorage allow us to access the pool struct from key values store
// this is an interface thus we could write unit tests
type poolStorage interface {
	PoolExist(ctx sdk.Context, ticker Ticker) bool

	GetPool(ctx sdk.Context, ticker Ticker) Pool
	SetPool(ctx sdk.Context, ticker Ticker, ps Pool)

	GetStakerPool(ctx sdk.Context, stakerID BnbAddress) (StakerPool, error)
	SetStakerPool(ctx sdk.Context, stakerID BnbAddress, sp StakerPool)

	GetPoolStaker(ctx sdk.Context, ticker Ticker) (PoolStaker, error)
	SetPoolStaker(ctx sdk.Context, ticker Ticker, ps PoolStaker)

	GetAdminConfig(ctx sdk.Context, key AdminConfigKey) AdminConfig
}
