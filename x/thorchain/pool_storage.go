package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
)

// PoolStorage allow us to access the pool struct from key values store
// this is an interface thus we could write unit tests
type poolStorage interface {
	PoolExist(ctx sdk.Context, asset common.Asset) bool

	GetPool(ctx sdk.Context, asset common.Asset) Pool
	SetPool(ctx sdk.Context, ps Pool)

	GetStakerPool(ctx sdk.Context, stakerID common.Address) (StakerPool, error)
	SetStakerPool(ctx sdk.Context, stakerID common.Address, sp StakerPool)

	AddToLiquidityFees(ctx sdk.Context, pool Pool, fee sdk.Uint) error

	GetPoolStaker(ctx sdk.Context, asset common.Asset) (PoolStaker, error)
	SetPoolStaker(ctx sdk.Context, asset common.Asset, ps PoolStaker)

	GetAdminConfigDefaultPoolStatus(ctx sdk.Context, add sdk.AccAddress) PoolStatus
	GetAdminConfigValue(ctx sdk.Context, key AdminConfigKey, addr sdk.AccAddress) (string, error)

	GetAdminConfigStakerAmtInterval(ctx sdk.Context, addr sdk.AccAddress) common.Amount
	GetLowestActiveVersion(ctx sdk.Context) int64

	AddIncompleteEvents(ctx sdk.Context, event Event)
	SetCompletedEvent(ctx sdk.Context, event Event)
}
