package thorchain

import (
	"errors"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"

	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

var notExistPoolStakerAsset, _ = common.NewAsset("BNB.NotExistPoolStakerAsset")

type MockInMemoryPoolStorage struct {
	KVStoreDummy
	store map[string]interface{}
}

// NewMockInMemoryPoolStorage
func NewMockInMemoryPoolStorage() *MockInMemoryPoolStorage {
	return &MockInMemoryPoolStorage{store: make(map[string]interface{})}
}

func (p *MockInMemoryPoolStorage) PoolExist(ctx sdk.Context, asset common.Asset) bool {
	_, ok := p.store[asset.String()]
	return ok
}

func (p *MockInMemoryPoolStorage) GetPool(ctx sdk.Context, asset common.Asset) (Pool, error) {
	if p, ok := p.store[asset.String()]; ok {
		return p.(Pool), nil
	}
	return types.NewPool(), nil
}

func (p *MockInMemoryPoolStorage) SetPool(ctx sdk.Context, ps Pool) error {
	p.store[ps.Asset.String()] = ps
	return nil
}

func (p *MockInMemoryPoolStorage) AddToLiquidityFees(ctx sdk.Context, asset common.Asset, fs sdk.Uint) error {
	return nil
}

func (p *MockInMemoryPoolStorage) GetTotalLiquidityFees(ctx sdk.Context, height uint64) (sdk.Uint, error) {
	return sdk.ZeroUint(), nil
}

func (p *MockInMemoryPoolStorage) GetPoolLiquidityFees(ctx sdk.Context, height uint64, asset common.Asset) (sdk.Uint, error) {
	return sdk.ZeroUint(), nil
}

func (p *MockInMemoryPoolStorage) GetPoolStaker(ctx sdk.Context, asset common.Asset) (PoolStaker, error) {
	if notExistPoolStakerAsset.Equals(asset) {
		return NewPoolStaker(asset, sdk.ZeroUint()), errors.New("simulate error for test")
	}
	key := p.GetKey(ctx, prefixPoolStaker, asset.String())
	if res, ok := p.store[key]; ok {
		return res.(PoolStaker), nil
	}
	return NewPoolStaker(asset, sdk.ZeroUint()), nil
}

func (p *MockInMemoryPoolStorage) SetPoolStaker(ctx sdk.Context, ps PoolStaker) {
	key := p.GetKey(ctx, prefixPoolStaker, ps.Asset.String())
	p.store[key] = ps
}

func (p *MockInMemoryPoolStorage) GetLowestActiveVersion(ctx sdk.Context) semver.Version {
	return constants.SWVersion
}

func (p *MockInMemoryPoolStorage) AddIncompleteEvents(ctx sdk.Context, event Event) error { return nil }
func (p *MockInMemoryPoolStorage) SetCompletedEvent(ctx sdk.Context, event Event)         {}

func (p *MockInMemoryPoolStorage) AddFeeToReserve(ctx sdk.Context, fee sdk.Uint) error { return nil }

func (p *MockInMemoryPoolStorage) GetGas(ctx sdk.Context, asset common.Asset) ([]sdk.Uint, error) {
	return []sdk.Uint{sdk.NewUint(37500), sdk.NewUint(30000)}, nil
}
