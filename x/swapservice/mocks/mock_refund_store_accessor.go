package mocks

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/bepswap/common"
	"gitlab.com/thorchain/bepswap/statechain/x/swapservice/types"
)

const (
	RefundAdminConfigKeyMRRA = `RefundAdminConfigKeyMRRA`
	RefundPoolKey            = `RefundPoolKey`
)

// MockRefundStoreAccessor implements PoolStorage interface, thus we can mock the error cases
type MockRefundStoreAccessor struct {
}

func NewMockRefundStoreAccessor() *MockRefundStoreAccessor {
	return &MockRefundStoreAccessor{}
}

func (mrsa MockRefundStoreAccessor) GetAdminConfigMRRA(ctx sdk.Context, bnb common.BnbAddress) common.Amount {
	v := ctx.Value(RefundAdminConfigKeyMRRA)
	if ac, ok := v.(common.Amount); ok {
		return ac
	}
	return common.ZeroAmount
}

// GetPool return an instance of Pool
func (mrsa MockRefundStoreAccessor) GetPool(ctx sdk.Context, ticker common.Ticker) types.Pool {
	v := ctx.Value(RefundPoolKey)
	if ps, ok := v.(types.Pool); ok {
		return ps
	}
	return types.NewPool()
}
