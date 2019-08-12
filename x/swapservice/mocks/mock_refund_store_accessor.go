package mocks

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/statechain/x/swapservice/types"
)

const (
	RefundAdminConfigKey = `RefundAdminConfigKey`
	RefundPoolStructKey  = `RefundPoolStructKey`
)

// MockRefundStoreAccessor implements PoolStorage interface, thus we can mock the error cases
type MockRefundStoreAccessor struct {
}

func NewMockRefundStoreAccessor() *MockRefundStoreAccessor {
	return &MockRefundStoreAccessor{}
}

func (mrsa MockRefundStoreAccessor) GetAdminConfigMRRA(ctx sdk.Context) types.Amount {
	v := ctx.Value(RefundAdminConfigKey)
	if ac, ok := v.(types.Amount); ok {
		return ac
	}
	return types.ZeroAmount
}

// GetPoolStruct return an instance of PoolStruct
func (mrsa MockRefundStoreAccessor) GetPoolStruct(ctx sdk.Context, ticker types.Ticker) types.PoolStruct {
	v := ctx.Value(RefundPoolStructKey)
	if ps, ok := v.(types.PoolStruct); ok {
		return ps
	}
	return types.NewPoolStruct()
}
