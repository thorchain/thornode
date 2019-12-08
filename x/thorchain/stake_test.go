package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type StakeSuite struct{}

var _ = Suite(&StakeSuite{})

func (s StakeSuite) TestCalculatePoolUnits(c *C) {
	inputs := []struct {
		name         string
		oldPoolUnits sdk.Uint
		poolRune     sdk.Uint
		poolAsset    sdk.Uint
		stakeRune    sdk.Uint
		stakeAsset   sdk.Uint
		poolUnits    sdk.Uint
		stakerUnits  sdk.Uint
		expectedErr  error
	}{
		{
			name:         "first-stake-zero-rune",
			oldPoolUnits: sdk.ZeroUint(),
			poolRune:     sdk.ZeroUint(),
			poolAsset:    sdk.ZeroUint(),
			stakeRune:    sdk.ZeroUint(),
			stakeAsset:   sdk.NewUint(100 * common.One),
			poolUnits:    sdk.ZeroUint(),
			stakerUnits:  sdk.ZeroUint(),
			expectedErr:  errors.New("total RUNE in the pool is zero"),
		},
		{
			name:         "first-stake-zero-asset",
			oldPoolUnits: sdk.ZeroUint(),
			poolRune:     sdk.ZeroUint(),
			poolAsset:    sdk.ZeroUint(),
			stakeRune:    sdk.NewUint(100 * common.One),
			stakeAsset:   sdk.ZeroUint(),
			poolUnits:    sdk.ZeroUint(),
			stakerUnits:  sdk.ZeroUint(),
			expectedErr:  errors.New("total asset in the pool is zero"),
		},
		{
			name:         "first-stake",
			oldPoolUnits: sdk.ZeroUint(),
			poolRune:     sdk.ZeroUint(),
			poolAsset:    sdk.ZeroUint(),
			stakeRune:    sdk.NewUint(100 * common.One),
			stakeAsset:   sdk.NewUint(100 * common.One),
			poolUnits:    sdk.NewUint(100 * common.One),
			stakerUnits:  sdk.NewUint(100 * common.One),
			expectedErr:  nil,
		},
		{
			name:         "second-stake",
			oldPoolUnits: sdk.NewUint(500 * common.One),
			poolRune:     sdk.NewUint(500 * common.One),
			poolAsset:    sdk.NewUint(500 * common.One),
			stakeRune:    sdk.NewUint(345 * common.One),
			stakeAsset:   sdk.NewUint(234 * common.One),
			poolUnits:    sdk.NewUint(78701684859),
			stakerUnits:  sdk.NewUint(28701684859),
			expectedErr:  nil,
		},
	}

	for _, item := range inputs {
		poolUnits, stakerUnits, err := calculatePoolUnits(item.oldPoolUnits, item.poolRune, item.poolAsset, item.stakeRune, item.stakeAsset)
		if item.expectedErr == nil {
			c.Assert(err, IsNil)
		} else {
			c.Assert(err.Error(), Equals, item.expectedErr.Error())
		}

		c.Logf("poolUnits:%s,expectedUnits:%s", poolUnits, item.poolUnits)
		c.Check(item.poolUnits.Uint64(), Equals, poolUnits.Uint64())
		c.Logf("stakerUnits:%s,expectedStakerUnits:%s", stakerUnits, item.stakerUnits)
		c.Check(item.stakerUnits.Uint64(), Equals, stakerUnits.Uint64())
	}
}

// TestValidateStakeMessage
func (StakeSuite) TestValidateStakeMessage(c *C) {
	ps := NewMockInMemoryPoolStorage()
	ctx, _ := setupKeeperForTest(c)
	txID := GetRandomTxHash()
	bnbAddress := GetRandomBNBAddress()
	assetAddress := GetRandomBNBAddress()
	c.Assert(validateStakeMessage(ctx, ps, common.Asset{}, txID, bnbAddress, assetAddress), NotNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BNBAsset, txID, bnbAddress, assetAddress), NotNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BNBAsset, txID, bnbAddress, assetAddress), NotNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BNBAsset, common.TxID(""), bnbAddress, assetAddress), NotNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BNBAsset, txID, common.NoAddress, common.NoAddress), NotNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BNBAsset, txID, bnbAddress, assetAddress), NotNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BNBAsset, txID, common.NoAddress, assetAddress), NotNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BTCAsset, txID, bnbAddress, common.NoAddress), NotNil)
	c.Assert(ps.SetPool(ctx, Pool{
		BalanceRune:  sdk.NewUint(100 * common.One),
		BalanceAsset: sdk.NewUint(100 * common.One),
		Asset:        common.BNBAsset,
		PoolUnits:    sdk.NewUint(100 * common.One),
		PoolAddress:  bnbAddress,
		Status:       PoolEnabled,
	}), IsNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BNBAsset, txID, bnbAddress, assetAddress), Equals, nil)
}

// TestStake test stake func
func (StakeSuite) TestStake(c *C) {
	ps := NewMockInMemoryPoolStorage()
	ctx, _ := setupKeeperForTest(c)
	txID := GetRandomTxHash()

	bnbAddress := GetRandomBNBAddress()
	assetAddress := GetRandomBNBAddress()
	btcAddress, err := common.NewAddress("bc1qwqdg6squsna38e46795at95yu9atm8azzmyvckulcc7kytlcckxswvvzej")
	c.Assert(err, IsNil)

	_, err = stake(ctx, ps, common.Asset{}, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One), bnbAddress, assetAddress, txID)
	c.Assert(err, NotNil)
	c.Assert(ps.SetPool(ctx, Pool{
		BalanceRune:  sdk.ZeroUint(),
		BalanceAsset: sdk.NewUint(100 * common.One),
		Asset:        common.BNBAsset,
		PoolUnits:    sdk.NewUint(100 * common.One),
		PoolAddress:  bnbAddress,
		Status:       PoolEnabled,
	}), IsNil)
	stakerUnit, err := stake(ctx, ps, common.BNBAsset, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One), bnbAddress, assetAddress, txID)
	c.Assert(stakerUnit.Equal(sdk.NewUint(11250000000)), Equals, true)
	c.Assert(err, IsNil)

	c.Assert(ps.SetPool(ctx, Pool{
		BalanceRune:  sdk.NewUint(100 * common.One),
		BalanceAsset: sdk.NewUint(100 * common.One),
		Asset:        notExistPoolStakerAsset,
		PoolUnits:    sdk.NewUint(100 * common.One),
		PoolAddress:  bnbAddress,
		Status:       PoolEnabled,
	}), IsNil)
	// stake asymmetically
	_, err = stake(ctx, ps, common.BNBAsset, sdk.NewUint(100*common.One), sdk.ZeroUint(), bnbAddress, assetAddress, txID)
	c.Assert(err, IsNil)
	_, err = stake(ctx, ps, common.BNBAsset, sdk.ZeroUint(), sdk.NewUint(100*common.One), bnbAddress, assetAddress, txID)
	c.Assert(err, IsNil)

	_, err = stake(ctx, ps, notExistPoolStakerAsset, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One), bnbAddress, assetAddress, txID)
	c.Assert(err, NotNil)
	c.Assert(ps.SetPool(ctx, Pool{
		BalanceRune:  sdk.NewUint(100 * common.One),
		BalanceAsset: sdk.NewUint(100 * common.One),
		Asset:        common.BNBAsset,
		PoolUnits:    sdk.NewUint(100 * common.One),
		PoolAddress:  bnbAddress,
		Status:       PoolEnabled,
	}), IsNil)
	makePoolStaker := func(total int, avg sdk.Uint) PoolStaker {
		stakers := make([]StakerUnit, total)
		for i := range stakers {
			stakers[i] = StakerUnit{Units: avg}
		}

		return PoolStaker{
			Asset:      common.BNBAsset,
			TotalUnits: avg.MulUint64(uint64(total)),
			Stakers:    stakers,
		}
	}
	skrs := makePoolStaker(150, sdk.NewUint(common.One/5000))
	ps.SetPoolStaker(ctx, skrs)
	_, err = stake(ctx, ps, common.BNBAsset, sdk.NewUint(common.One), sdk.NewUint(common.One), bnbAddress, assetAddress, txID)
	c.Assert(err, IsNil)

	_, err = stake(ctx, ps, common.BNBAsset, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One), notExistStakerPoolAddr, notExistStakerPoolAddr, txID)
	c.Assert(err, NotNil)
	c.Assert(ps.SetPool(ctx, Pool{
		BalanceRune:  sdk.NewUint(100 * common.One),
		BalanceAsset: sdk.NewUint(100 * common.One),
		Asset:        common.BNBAsset,
		PoolUnits:    sdk.NewUint(100 * common.One),
		PoolAddress:  bnbAddress,
		Status:       PoolEnabled,
	}), IsNil)
	_, err = stake(ctx, ps, common.BNBAsset, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One), bnbAddress, assetAddress, txID)
	c.Assert(err, IsNil)
	p, err := ps.GetPool(ctx, common.BNBAsset)
	c.Assert(err, IsNil)

	c.Check(p.PoolUnits.Equal(sdk.NewUint(200*common.One)), Equals, true)

	// Test atomic cross chain staking
	// create BTC pool
	c.Assert(ps.SetPool(ctx, Pool{
		BalanceRune:  sdk.ZeroUint(),
		BalanceAsset: sdk.ZeroUint(),
		Asset:        common.BTCAsset,
		PoolUnits:    sdk.ZeroUint(),
		PoolAddress:  btcAddress,
		Status:       PoolEnabled,
	}), IsNil)

	// stake rune
	stakerUnit, err = stake(ctx, ps, common.BTCAsset, sdk.NewUint(100*common.One), sdk.ZeroUint(), bnbAddress, btcAddress, txID)
	c.Assert(err, IsNil)
	c.Check(stakerUnit.IsZero(), Equals, true)
	// stake btc
	stakerUnit, err = stake(ctx, ps, common.BTCAsset, sdk.ZeroUint(), sdk.NewUint(100*common.One), bnbAddress, btcAddress, txID)
	c.Assert(err, IsNil)
	c.Check(stakerUnit.IsZero(), Equals, false)
	p, err = ps.GetPool(ctx, common.BTCAsset)
	c.Assert(err, IsNil)
	c.Check(p.BalanceAsset.Equal(sdk.NewUint(100*common.One)), Equals, true, Commentf("%d", p.BalanceAsset.Uint64()))
	c.Check(p.BalanceRune.Equal(sdk.NewUint(100*common.One)), Equals, true, Commentf("%d", p.BalanceRune.Uint64()))
	c.Check(p.PoolUnits.Equal(sdk.NewUint(100*common.One)), Equals, true, Commentf("%d", p.PoolUnits.Uint64()))
}
