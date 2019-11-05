package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"
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

func (s StakeSuite) TestValidateAmount(c *C) {
	makePoolStaker := func(total uint64, avg sdk.Uint) PoolStaker {
		stakers := make([]StakerUnit, total)
		for i := range stakers {
			stakers[i] = StakerUnit{Units: avg}
		}

		return PoolStaker{
			TotalUnits: avg.MulUint64(total),
			Stakers:    stakers,
		}
	}

	skrs := makePoolStaker(50, sdk.NewUint(common.One/1000))
	c.Assert(validateStakeAmount(skrs, sdk.NewUint(common.One/1000), common.NewAmountFromFloat(100)), IsNil)

	skrs = makePoolStaker(150, sdk.NewUint(common.One/5000))
	c.Assert(validateStakeAmount(skrs, sdk.NewUint(common.One/10000), common.NewAmountFromFloat(100)), NotNil)
	c.Assert(validateStakeAmount(skrs, sdk.NewUint(common.One/5000), common.NewAmountFromFloat(100)), NotNil)
	c.Assert(validateStakeAmount(skrs, sdk.NewUint(common.One/1000), common.NewAmountFromFloat(100)), IsNil)

	skrs = makePoolStaker(300, sdk.NewUint(common.One/1000))

	c.Assert(validateStakeAmount(skrs, sdk.NewUint(common.One/10000), common.NewAmountFromFloat(100)), NotNil)
	c.Assert(validateStakeAmount(skrs, sdk.NewUint(common.One/500), common.NewAmountFromFloat(100)), NotNil)
	c.Assert(validateStakeAmount(skrs, sdk.NewUint(common.One/250), common.NewAmountFromFloat(100)), IsNil)
}

// TestValidateStakeMessage
func (StakeSuite) TestValidateStakeMessage(c *C) {
	ps := NewMockInMemoryPoolStorage()
	ctx, _ := setupKeeperForTest(c)
	txId := GetRandomTxHash()
	bnbAddress := GetRandomBNBAddress()
	assetAddress := GetRandomBNBAddress()
	c.Assert(validateStakeMessage(ctx, ps, common.Asset{}, txId, bnbAddress, assetAddress), NotNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BNBAsset, txId, bnbAddress, assetAddress), NotNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BNBAsset, txId, bnbAddress, assetAddress), NotNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BNBAsset, common.TxID(""), bnbAddress, assetAddress), NotNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BNBAsset, txId, common.NoAddress, common.NoAddress), NotNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BNBAsset, txId, bnbAddress, assetAddress), NotNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BNBAsset, txId, common.NoAddress, assetAddress), NotNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BTCAsset, txId, bnbAddress, common.NoAddress), NotNil)
	ps.SetPool(ctx, Pool{
		BalanceRune:  sdk.NewUint(100 * common.One),
		BalanceAsset: sdk.NewUint(100 * common.One),
		Asset:        common.BNBAsset,
		PoolUnits:    sdk.NewUint(100 * common.One),
		PoolAddress:  bnbAddress,
		Status:       PoolEnabled,
	})
	c.Assert(validateStakeMessage(ctx, ps, common.BNBAsset, txId, bnbAddress, assetAddress), Equals, nil)
}

// TestStake test stake func
func (StakeSuite) TestStake(c *C) {
	ps := NewMockInMemoryPoolStorage()
	ctx, _ := setupKeeperForTest(c)
	txId := GetRandomTxHash()

	bnbAddress := GetRandomBNBAddress()
	assetAddress := GetRandomBNBAddress()
	btcAddress, err := common.NewAddress("bc1qwqdg6squsna38e46795at95yu9atm8azzmyvckulcc7kytlcckxswvvzej")
	c.Assert(err, IsNil)

	_, err = stake(ctx, ps, common.Asset{}, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One), bnbAddress, assetAddress, txId)
	c.Assert(err, NotNil)
	ps.SetPool(ctx, Pool{
		BalanceRune:  sdk.ZeroUint(),
		BalanceAsset: sdk.NewUint(100 * common.One),
		Asset:        common.BNBAsset,
		PoolUnits:    sdk.NewUint(100 * common.One),
		PoolAddress:  bnbAddress,
		Status:       PoolEnabled,
	})
	stakerUnit, err := stake(ctx, ps, common.BNBAsset, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One), bnbAddress, assetAddress, txId)
	c.Assert(stakerUnit.Equal(sdk.NewUint(11250000000)), Equals, true)
	c.Assert(err, IsNil)
	ps.SetPool(ctx, Pool{
		BalanceRune:  sdk.NewUint(100 * common.One),
		BalanceAsset: sdk.NewUint(100 * common.One),
		Asset:        notExistPoolStakerAsset,
		PoolUnits:    sdk.NewUint(100 * common.One),
		PoolAddress:  bnbAddress,
		Status:       PoolEnabled,
	})
	// stake asymmetically
	_, err = stake(ctx, ps, common.BNBAsset, sdk.NewUint(100*common.One), sdk.ZeroUint(), bnbAddress, assetAddress, txId)
	c.Assert(err, IsNil)
	_, err = stake(ctx, ps, common.BNBAsset, sdk.ZeroUint(), sdk.NewUint(100*common.One), bnbAddress, assetAddress, txId)
	c.Assert(err, IsNil)

	_, err = stake(ctx, ps, notExistPoolStakerAsset, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One), bnbAddress, assetAddress, txId)
	c.Assert(err, NotNil)
	ps.SetPool(ctx, Pool{
		BalanceRune:  sdk.NewUint(100 * common.One),
		BalanceAsset: sdk.NewUint(100 * common.One),
		Asset:        common.BNBAsset,
		PoolUnits:    sdk.NewUint(100 * common.One),
		PoolAddress:  bnbAddress,
		Status:       PoolEnabled,
	})
	makePoolStaker := func(total int, avg sdk.Uint) PoolStaker {
		stakers := make([]StakerUnit, total)
		for i := range stakers {
			stakers[i] = StakerUnit{Units: avg}
		}

		return PoolStaker{
			TotalUnits: avg.MulUint64(uint64(total)),
			Stakers:    stakers,
		}
	}
	skrs := makePoolStaker(150, sdk.NewUint(common.One/5000))
	ps.SetPoolStaker(ctx, common.BNBAsset, skrs)
	_, err = stake(ctx, ps, common.BNBAsset, sdk.NewUint(common.One), sdk.NewUint(common.One), bnbAddress, assetAddress, txId)
	c.Assert(err, NotNil)

	_, err = stake(ctx, ps, common.BNBAsset, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One), notExistStakerPoolAddr, notExistStakerPoolAddr, txId)
	c.Assert(err, NotNil)
	ps.SetPool(ctx, Pool{
		BalanceRune:  sdk.NewUint(100 * common.One),
		BalanceAsset: sdk.NewUint(100 * common.One),
		Asset:        common.BNBAsset,
		PoolUnits:    sdk.NewUint(100 * common.One),
		PoolAddress:  bnbAddress,
		Status:       PoolEnabled,
	})
	_, err = stake(ctx, ps, common.BNBAsset, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One), bnbAddress, assetAddress, txId)
	c.Assert(err, IsNil)
	p := ps.GetPool(ctx, common.BNBAsset)

	c.Check(p.PoolUnits.Equal(sdk.NewUint(200*common.One)), Equals, true)

	// Test atomic cross chain staking
	// create BTC pool
	ps.SetPool(ctx, Pool{
		BalanceRune:  sdk.ZeroUint(),
		BalanceAsset: sdk.ZeroUint(),
		Asset:        common.BTCAsset,
		PoolUnits:    sdk.ZeroUint(),
		PoolAddress:  btcAddress,
		Status:       PoolEnabled,
	})

	// stake rune
	stakerUnit, err = stake(ctx, ps, common.BTCAsset, sdk.NewUint(100*common.One), sdk.ZeroUint(), bnbAddress, btcAddress, txId)
	c.Assert(err, IsNil)
	c.Check(stakerUnit.IsZero(), Equals, true)
	// stake btc
	stakerUnit, err = stake(ctx, ps, common.BTCAsset, sdk.ZeroUint(), sdk.NewUint(100*common.One), bnbAddress, btcAddress, txId)
	c.Assert(err, IsNil)
	c.Check(stakerUnit.IsZero(), Equals, false)
	p = ps.GetPool(ctx, common.BTCAsset)
	c.Check(p.BalanceAsset.Equal(sdk.NewUint(100*common.One)), Equals, true, Commentf("%d", p.BalanceAsset.Uint64()))
	c.Check(p.BalanceRune.Equal(sdk.NewUint(100*common.One)), Equals, true, Commentf("%d", p.BalanceRune.Uint64()))
	c.Check(p.PoolUnits.Equal(sdk.NewUint(100*common.One)), Equals, true, Commentf("%d", p.PoolUnits.Uint64()))
}
