package thorchain

import (
	"errors"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"

	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

var (
	notExistPoolStakerAsset, _ = common.NewAsset("BNB.NotExistPoolStakerAsset")
	notExistStakerPoolAddr     = common.Address("4252BA642F73FA402FEF18E3CB4550E5A4A6831299D5EB7E76808C8923FC1XXX")
)

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
func (p *MockInMemoryPoolStorage) GetPool(ctx sdk.Context, asset common.Asset) Pool {
	if p, ok := p.store[asset.String()]; ok {
		return p.(Pool)
	}
	return types.NewPool()
}
func (p *MockInMemoryPoolStorage) SetPool(ctx sdk.Context, ps Pool) {
	p.store[ps.Asset.String()] = ps
}

func (p *MockInMemoryPoolStorage) AddToLiquidityFees(ctx sdk.Context, ps Pool, fs sdk.Uint) error {
	return nil
}

func (p *MockInMemoryPoolStorage) GetTotalLiquidityFees(ctx sdk.Context, height uint64) (sdk.Uint, error) {
	return sdk.ZeroUint(), nil
}

func (p *MockInMemoryPoolStorage) GetPoolLiquidityFees(ctx sdk.Context, height uint64, pool Pool) (sdk.Uint, error) {
	return sdk.ZeroUint(), nil
}

func (p *MockInMemoryPoolStorage) GetStakerPool(ctx sdk.Context, stakerID common.Address) (StakerPool, error) {
	if stakerID.Equals(notExistStakerPoolAddr) {
		return NewStakerPool(stakerID), errors.New("simulate error for test")
	}
	key := p.GetKey(ctx, prefixStakerPool, stakerID.String())
	if res, ok := p.store[key]; ok {
		return res.(StakerPool), nil
	}
	return NewStakerPool(stakerID), nil
}
func (p *MockInMemoryPoolStorage) SetStakerPool(ctx sdk.Context, stakerID common.Address, sp StakerPool) {
	key := p.GetKey(ctx, prefixStakerPool, stakerID.String())
	p.store[key] = sp
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
func (p *MockInMemoryPoolStorage) SetPoolStaker(ctx sdk.Context, asset common.Asset, ps PoolStaker) {
	key := p.GetKey(ctx, prefixPoolStaker, asset.String())
	p.store[key] = ps
}

func (p *MockInMemoryPoolStorage) GetLowestActiveVersion(ctx sdk.Context) semver.Version {
	return semver.MustParse("0.1.0")
}

func (p *MockInMemoryPoolStorage) GetAdminConfigDefaultPoolStatus(ctx sdk.Context, addr sdk.AccAddress) PoolStatus {
	return PoolBootstrap
}

func (p *MockInMemoryPoolStorage) GetAdminConfigValue(ctx sdk.Context, key AdminConfigKey, addr sdk.AccAddress) (string, error) {
	storekey := p.GetKey(ctx, prefixAdmin, key.String())
	ac, ok := p.store[storekey]
	if ok {
		return ac.(AdminConfig).Value, nil
	}
	return "", nil
}

func (p *MockInMemoryPoolStorage) GetAdminConfigStakerAmtInterval(ctx sdk.Context, addr sdk.AccAddress) common.Amount {
	return common.NewAmountFromFloat(100)
}

func (p *MockInMemoryPoolStorage) AddIncompleteEvents(ctx sdk.Context, event Event) {}
func (p *MockInMemoryPoolStorage) SetCompletedEvent(ctx sdk.Context, event Event)   {}

func (p *MockInMemoryPoolStorage) AddFeeToReserve(ctx sdk.Context, fee sdk.Uint) {}
