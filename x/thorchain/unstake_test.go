package thorchain

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

type UnstakeSuite struct{}

var _ = Suite(&UnstakeSuite{})

func (s *UnstakeSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

type UnstakeTestKeeper struct {
	KVStoreDummy
	store map[string]interface{}
}

func NewUnstakeTestKeeper() *UnstakeTestKeeper {
	return &UnstakeTestKeeper{
		store: make(map[string]interface{}),
	}
}

func (k *UnstakeTestKeeper) PoolExist(ctx sdk.Context, asset common.Asset) bool {
	if asset.Equals(common.Asset{Chain: common.BNBChain, Symbol: "NOTEXIST", Ticker: "NOTEXIST"}) {
		return false
	}
	return true
}

func (k *UnstakeTestKeeper) GetPool(ctx sdk.Context, asset common.Asset) (types.Pool, error) {
	if asset.Equals(common.Asset{Chain: common.BNBChain, Symbol: "NOTEXIST", Ticker: "NOTEXIST"}) {
		return types.Pool{}, nil
	} else {
		return types.Pool{
			BalanceRune:  sdk.NewUint(100).MulUint64(common.One),
			BalanceAsset: sdk.NewUint(100).MulUint64(common.One),
			PoolUnits:    sdk.NewUint(100).MulUint64(common.One),
			Status:       types.Enabled,
			Asset:        asset,
		}, nil
	}
}

func (k *UnstakeTestKeeper) SetPool(ctx sdk.Context, ps Pool) error {
	k.store[ps.Asset.String()] = ps
	return nil
}

func (k *UnstakeTestKeeper) GetGas(ctx sdk.Context, asset common.Asset) ([]sdk.Uint, error) {
	return []sdk.Uint{sdk.NewUint(37500), sdk.NewUint(30000)}, nil
}

func (p *UnstakeTestKeeper) GetStaker(ctx sdk.Context, asset common.Asset, addr common.Address) (Staker, error) {
	if asset.Equals(common.Asset{Chain: common.BNBChain, Symbol: "NOTEXISTSTICKER", Ticker: "NOTEXISTSTICKER"}) {
		return types.Staker{}, errors.New("you asked for it")
	}
	if notExistStakerAsset.Equals(asset) {
		return Staker{}, errors.New("simulate error for test")
	}
	staker := Staker{
		Asset:       asset,
		RuneAddress: addr,
		Units:       sdk.ZeroUint(),
		PendingRune: sdk.ZeroUint(),
	}
	key := p.GetKey(ctx, prefixStaker, staker.Key())
	if res, ok := p.store[key]; ok {
		return res.(Staker), nil
	}
	return staker, nil
}

func (p *UnstakeTestKeeper) SetStaker(ctx sdk.Context, staker Staker) {
	key := p.GetKey(ctx, prefixStaker, staker.Key())
	p.store[key] = staker
}

func (s UnstakeSuite) TestCalculateUnsake(c *C) {
	inputs := []struct {
		name                 string
		poolUnit             sdk.Uint
		poolRune             sdk.Uint
		poolAsset            sdk.Uint
		stakerUnit           sdk.Uint
		percentage           sdk.Uint
		expectedUnstakeRune  sdk.Uint
		expectedUnstakeAsset sdk.Uint
		expectedUnitLeft     sdk.Uint
		expectedErr          error
	}{
		{
			name:                 "zero-poolunit",
			poolUnit:             sdk.ZeroUint(),
			poolRune:             sdk.ZeroUint(),
			poolAsset:            sdk.ZeroUint(),
			stakerUnit:           sdk.ZeroUint(),
			percentage:           sdk.ZeroUint(),
			expectedUnstakeRune:  sdk.ZeroUint(),
			expectedUnstakeAsset: sdk.ZeroUint(),
			expectedUnitLeft:     sdk.ZeroUint(),
			expectedErr:          errors.New("poolUnits can't be zero"),
		},

		{
			name:                 "zero-poolrune",
			poolUnit:             sdk.NewUint(500 * common.One),
			poolRune:             sdk.ZeroUint(),
			poolAsset:            sdk.ZeroUint(),
			stakerUnit:           sdk.ZeroUint(),
			percentage:           sdk.ZeroUint(),
			expectedUnstakeRune:  sdk.ZeroUint(),
			expectedUnstakeAsset: sdk.ZeroUint(),
			expectedUnitLeft:     sdk.ZeroUint(),
			expectedErr:          errors.New("pool rune balance can't be zero"),
		},

		{
			name:                 "zero-poolasset",
			poolUnit:             sdk.NewUint(500 * common.One),
			poolRune:             sdk.NewUint(500 * common.One),
			poolAsset:            sdk.ZeroUint(),
			stakerUnit:           sdk.ZeroUint(),
			percentage:           sdk.ZeroUint(),
			expectedUnstakeRune:  sdk.ZeroUint(),
			expectedUnstakeAsset: sdk.ZeroUint(),
			expectedUnitLeft:     sdk.ZeroUint(),
			expectedErr:          errors.New("pool asset balance can't be zero"),
		},
		{
			name:                 "negative-stakerUnit",
			poolUnit:             sdk.NewUint(500 * common.One),
			poolRune:             sdk.NewUint(500 * common.One),
			poolAsset:            sdk.NewUint(5100 * common.One),
			stakerUnit:           sdk.ZeroUint(),
			percentage:           sdk.ZeroUint(),
			expectedUnstakeRune:  sdk.ZeroUint(),
			expectedUnstakeAsset: sdk.ZeroUint(),
			expectedUnitLeft:     sdk.ZeroUint(),
			expectedErr:          errors.New("staker unit can't be zero"),
		},

		{
			name:                 "percentage-larger-than-100",
			poolUnit:             sdk.NewUint(500 * common.One),
			poolRune:             sdk.NewUint(500 * common.One),
			poolAsset:            sdk.NewUint(500 * common.One),
			stakerUnit:           sdk.NewUint(100 * common.One),
			percentage:           sdk.NewUint(12000),
			expectedUnstakeRune:  sdk.ZeroUint(),
			expectedUnstakeAsset: sdk.ZeroUint(),
			expectedUnitLeft:     sdk.ZeroUint(),
			expectedErr:          fmt.Errorf("withdraw basis point %s is not valid", sdk.NewUint(12000)),
		},
		{
			name:                 "unstake-1",
			poolUnit:             sdk.NewUint(700 * common.One),
			poolRune:             sdk.NewUint(700 * common.One),
			poolAsset:            sdk.NewUint(700 * common.One),
			stakerUnit:           sdk.NewUint(200 * common.One),
			percentage:           sdk.NewUint(10000),
			expectedUnitLeft:     sdk.ZeroUint(),
			expectedUnstakeAsset: sdk.NewUint(200 * common.One),
			expectedUnstakeRune:  sdk.NewUint(200 * common.One),
			expectedErr:          nil,
		},
		{
			name:                 "unstake-2",
			poolUnit:             sdk.NewUint(100),
			poolRune:             sdk.NewUint(15 * common.One),
			poolAsset:            sdk.NewUint(155 * common.One),
			stakerUnit:           sdk.NewUint(100),
			percentage:           sdk.NewUint(1000),
			expectedUnitLeft:     sdk.NewUint(90),
			expectedUnstakeAsset: sdk.NewUint(1550000000),
			expectedUnstakeRune:  sdk.NewUint(150000000),
			expectedErr:          nil,
		},
	}

	for _, item := range inputs {
		c.Logf("name:%s", item.name)
		withDrawRune, withDrawAsset, unitAfter, err := calculateUnstake(item.poolUnit, item.poolRune, item.poolAsset, item.stakerUnit, item.percentage)
		if item.expectedErr == nil {
			c.Assert(err, IsNil)
		} else {
			c.Assert(err.Error(), Equals, item.expectedErr.Error())
		}
		c.Logf("expected rune:%s,rune:%s", item.expectedUnstakeRune, withDrawRune)
		c.Check(item.expectedUnstakeRune.Uint64(), Equals, withDrawRune.Uint64(), Commentf("Expected %d, got %d", item.expectedUnstakeRune.Uint64(), withDrawRune.Uint64()))
		c.Check(item.expectedUnstakeAsset.Uint64(), Equals, withDrawAsset.Uint64(), Commentf("Expected %d, got %d", item.expectedUnstakeAsset.Uint64(), withDrawAsset.Uint64()))
		c.Check(item.expectedUnitLeft.Uint64(), Equals, unitAfter.Uint64())
	}
}

// TestValidateUnstake is to test validateUnstake function
func (s UnstakeSuite) TestValidateUnstake(c *C) {
	accountAddr := GetRandomNodeAccount(NodeWhiteListed).NodeAddress
	runeAddress, err := common.NewAddress("bnb1g0xakzh03tpa54khxyvheeu92hwzypkdce77rm")
	if err != nil {
		c.Error("fail to create new BNB Address")
	}
	inputs := []struct {
		name          string
		msg           MsgSetUnStake
		expectedError error
	}{
		{
			name: "empty-rune-address",
			msg: MsgSetUnStake{
				RuneAddress:        "",
				UnstakeBasisPoints: sdk.NewUint(10000),
				Asset:              common.BNBAsset,
				Tx:                 common.Tx{ID: "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE"},
				Signer:             accountAddr,
			},
			expectedError: errors.New("empty rune address"),
		},
		{
			name: "empty-withdraw-basis-points",
			msg: MsgSetUnStake{
				RuneAddress:        runeAddress,
				UnstakeBasisPoints: sdk.ZeroUint(),
				Asset:              common.BNBAsset,
				Tx:                 common.Tx{ID: "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE"},
				Signer:             accountAddr,
			},
			expectedError: nil,
		},
		{
			name: "empty-request-txhash",
			msg: MsgSetUnStake{
				RuneAddress:        runeAddress,
				UnstakeBasisPoints: sdk.NewUint(10000),
				Asset:              common.BNBAsset,
				Tx:                 common.Tx{},
				Signer:             accountAddr,
			},
			expectedError: errors.New("request tx hash is empty"),
		},
		{
			name: "empty-asset",
			msg: MsgSetUnStake{
				RuneAddress:        runeAddress,
				UnstakeBasisPoints: sdk.NewUint(10000),
				Asset:              common.Asset{},
				Tx:                 common.Tx{ID: "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE"},
				Signer:             accountAddr,
			},
			expectedError: errors.New("empty asset"),
		},
		{
			name: "invalid-basis-point",
			msg: MsgSetUnStake{
				RuneAddress:        runeAddress,
				UnstakeBasisPoints: sdk.NewUint(10001),
				Asset:              common.BNBAsset,
				Tx:                 common.Tx{ID: "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE"},
				Signer:             accountAddr,
			},
			expectedError: errors.New("withdraw basis points 10001 is invalid"),
		},
		{
			name: "invalid-pool-notexist",
			msg: MsgSetUnStake{
				RuneAddress:        runeAddress,
				UnstakeBasisPoints: sdk.NewUint(10000),
				Asset:              common.Asset{Chain: common.BNBChain, Ticker: "NOTEXIST", Symbol: "NOTEXIST"},
				Tx:                 common.Tx{ID: "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE"},
				Signer:             accountAddr,
			},
			expectedError: errors.New("pool-BNB.NOTEXIST doesn't exist"),
		},
		{
			name: "all-good",
			msg: MsgSetUnStake{
				RuneAddress:        runeAddress,
				UnstakeBasisPoints: sdk.NewUint(10000),
				Asset:              common.BNBAsset,
				Tx:                 common.Tx{ID: "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE"},
				Signer:             accountAddr,
			},
			expectedError: nil,
		},
	}

	for _, item := range inputs {
		ctx, _ := setupKeeperForTest(c)
		ps := &UnstakeTestKeeper{}
		c.Logf("name:%s", item.name)
		err := validateUnstake(ctx, ps, item.msg)
		if item.expectedError != nil {
			c.Assert(err, NotNil)
			c.Assert(err.Error(), Equals, item.expectedError.Error())
			continue
		}
		c.Assert(err, IsNil)
	}
}

func (UnstakeSuite) TestUnstake(c *C) {
	ps := &UnstakeTestKeeper{}
	accountAddr := GetRandomNodeAccount(NodeWhiteListed).NodeAddress
	runeAddress, err := common.NewAddress("bnb1g0xakzh03tpa54khxyvheeu92hwzypkdce77rm")
	if err != nil {
		c.Error("fail to create new BNB Address")
	}
	testCases := []struct {
		name          string
		msg           MsgSetUnStake
		ps            Keeper
		runeAmount    sdk.Uint
		assetAmount   sdk.Uint
		expectedError error
	}{
		{
			name: "empty-rune-address",
			msg: MsgSetUnStake{
				RuneAddress:        "",
				UnstakeBasisPoints: sdk.NewUint(10000),
				Asset:              common.BNBAsset,
				Tx:                 common.Tx{ID: "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE"},
				Signer:             accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			assetAmount:   sdk.ZeroUint(),
			expectedError: sdk.NewError(DefaultCodespace, CodeUnstakeFailValidation, "empty rune address"),
		},
		{
			name: "empty-withdraw-basis-points",
			msg: MsgSetUnStake{
				RuneAddress:        runeAddress,
				UnstakeBasisPoints: sdk.ZeroUint(),
				Asset:              common.BNBAsset,
				Tx:                 common.Tx{ID: "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE"},
				Signer:             accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			assetAmount:   sdk.ZeroUint(),
			expectedError: sdk.NewError(DefaultCodespace, CodeNoStakeUnitLeft, "nothing to withdraw"),
		},
		{
			name: "empty-request-txhash",
			msg: MsgSetUnStake{
				RuneAddress:        runeAddress,
				UnstakeBasisPoints: sdk.NewUint(10000),
				Asset:              common.BNBAsset,
				Tx:                 common.Tx{},
				Signer:             accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			assetAmount:   sdk.ZeroUint(),
			expectedError: sdk.NewError(DefaultCodespace, CodeUnstakeFailValidation, "request tx hash is empty"),
		},
		{
			name: "empty-asset",
			msg: MsgSetUnStake{
				RuneAddress:        runeAddress,
				UnstakeBasisPoints: sdk.NewUint(10000),
				Asset:              common.Asset{},
				Tx:                 common.Tx{ID: "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE"},
				Signer:             accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			assetAmount:   sdk.ZeroUint(),
			expectedError: sdk.NewError(DefaultCodespace, CodeUnstakeFailValidation, "empty asset"),
		},

		{
			name: "invalid-basis-point",
			msg: MsgSetUnStake{
				RuneAddress:        runeAddress,
				UnstakeBasisPoints: sdk.NewUint(10001),
				Asset:              common.BNBAsset,
				Tx:                 common.Tx{ID: "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE"},
				Signer:             accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			assetAmount:   sdk.ZeroUint(),
			expectedError: sdk.NewError(DefaultCodespace, CodeUnstakeFailValidation, "withdraw basis points 10001 is invalid"),
		},
		{
			name: "invalid-pool-notexist",
			msg: MsgSetUnStake{
				RuneAddress:        runeAddress,
				UnstakeBasisPoints: sdk.NewUint(10000),
				Asset:              common.Asset{Chain: common.BNBChain, Ticker: "NOTEXIST", Symbol: "NOTEXIST"},
				Tx:                 common.Tx{ID: "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE"},
				Signer:             accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			assetAmount:   sdk.ZeroUint(),
			expectedError: sdk.NewError(DefaultCodespace, CodeUnstakeFailValidation, "pool-BNB.NOTEXIST doesn't exist"),
		},
		{
			name: "invalid-pool-staker-notexist",
			msg: MsgSetUnStake{
				RuneAddress:        runeAddress,
				UnstakeBasisPoints: sdk.NewUint(10000),
				Asset:              common.Asset{Chain: common.BNBChain, Ticker: "NOTEXISTSTICKER", Symbol: "NOTEXISTSTICKER"},
				Tx:                 common.Tx{ID: "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE"},
				Signer:             accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			assetAmount:   sdk.ZeroUint(),
			expectedError: sdk.NewError(DefaultCodespace, CodeStakerNotExist, "staker doesn't exist"),
		},
		{
			name: "nothing-to-withdraw",
			msg: MsgSetUnStake{
				RuneAddress:        runeAddress,
				UnstakeBasisPoints: sdk.NewUint(0),
				Asset:              common.BNBAsset,
				Tx:                 common.Tx{ID: "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE"},
				Signer:             accountAddr,
			},
			ps:            ps,
			runeAmount:    sdk.ZeroUint(),
			assetAmount:   sdk.ZeroUint(),
			expectedError: sdk.NewError(DefaultCodespace, CodeNoStakeUnitLeft, "nothing to withdraw"),
		},
		{
			name: "all-good",
			msg: MsgSetUnStake{
				RuneAddress:        runeAddress,
				UnstakeBasisPoints: sdk.NewUint(10000),
				Asset:              common.BNBAsset,
				Tx:                 common.Tx{ID: "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE"},
				Signer:             accountAddr,
			},
			ps:            getUnstakeTestKeeper(c),
			runeAmount:    sdk.NewUint(100 * common.One),
			assetAmount:   sdk.NewUint(100 * common.One).Sub(sdk.NewUint(75000)),
			expectedError: nil,
		},
		{
			name: "all-good-half",
			msg: MsgSetUnStake{
				RuneAddress:        runeAddress,
				UnstakeBasisPoints: sdk.NewUint(5000),
				Asset:              common.BNBAsset,
				Tx:                 common.Tx{ID: "28B40BF105A112389A339A64BD1A042E6140DC9082C679586C6CF493A9FDE3FE"},
				Signer:             accountAddr,
			},
			ps:            getUnstakeTestKeeper(c),
			runeAmount:    sdk.NewUint(50 * common.One),
			assetAmount:   sdk.NewUint(50 * common.One),
			expectedError: nil,
		},
	}
	for _, tc := range testCases {
		ctx, _ := setupKeeperForTest(c)
		c.Logf("name:%s", tc.name)
		version := constants.SWVersion
		versionedEventManagerDummy := NewDummyVersionedEventMgr()
		eventManager, err := versionedEventManagerDummy.GetEventManager(ctx, version)
		c.Assert(err, IsNil)
		r, asset, _, _, err := unstake(ctx, version, tc.ps, tc.msg, eventManager)
		if tc.expectedError != nil {
			c.Assert(err, NotNil)
			c.Check(err.Error(), Equals, tc.expectedError.Error())
			c.Check(r.Uint64(), Equals, tc.runeAmount.Uint64())
			c.Check(asset.Uint64(), Equals, tc.assetAmount.Uint64())
			continue
		}
		c.Assert(err, IsNil)
		c.Check(r.Uint64(), Equals, tc.runeAmount.Uint64())
		c.Check(asset.Uint64(), Equals, tc.assetAmount.Uint64())
	}
}

func getUnstakeTestKeeper(c *C) Keeper {
	runeAddress, err := common.NewAddress("bnb1g0xakzh03tpa54khxyvheeu92hwzypkdce77rm")
	if err != nil {
		c.Error("fail to create new BNB Address")
	}

	ctx, _ := setupKeeperForTest(c)

	store := NewUnstakeTestKeeper()
	pool := Pool{
		BalanceRune:  sdk.NewUint(100 * common.One),
		BalanceAsset: sdk.NewUint(100 * common.One),
		Asset:        common.BNBAsset,
		PoolUnits:    sdk.NewUint(100 * common.One),
		PoolAddress:  runeAddress,
		Status:       PoolEnabled,
	}
	c.Assert(store.SetPool(ctx, pool), IsNil)
	staker := Staker{
		Asset:        pool.Asset,
		RuneAddress:  runeAddress,
		AssetAddress: runeAddress,
		Units:        sdk.NewUint(100 * common.One),
		PendingRune:  sdk.ZeroUint(),
	}
	store.SetStaker(ctx, staker)
	return store
}
