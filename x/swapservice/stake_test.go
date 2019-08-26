package swapservice

import (
	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type StakeSuite struct{}

var _ = Suite(&StakeSuite{})

func (s StakeSuite) TestCalculatePoolUnits(c *C) {
	inputs := []struct {
		name         string
		oldPoolUnits float64
		poolRune     float64
		poolToken    float64
		stakeRune    float64
		stakeToken   float64
		poolUnits    float64
		stakerUnits  float64
		expectedErr  error
	}{
		{
			name:         "negative-poolrune",
			oldPoolUnits: 0,
			poolRune:     -100.0,
			expectedErr:  errors.New("negative RUNE in the pool,likely it is corrupted"),
		},
		{
			name:         "negative-pooltoken",
			oldPoolUnits: 0,
			poolRune:     100.0,
			poolToken:    -100.0,
			expectedErr:  errors.New("negative token in the pool,likely it is corrupted"),
		},
		{
			name:         "negative-stakerune",
			oldPoolUnits: 0,
			poolRune:     100.0,
			poolToken:    100.0,
			stakeRune:    -100,
			expectedErr:  errors.New("you can't stake negative rune"),
		},
		{
			name:         "negative-staketoken",
			oldPoolUnits: 0,
			poolRune:     100.0,
			poolToken:    100.0,
			stakeRune:    100,
			stakeToken:   -100,
			expectedErr:  errors.New("you can't stake negative token"),
		},
		{
			name:         "first-stake-zero-rune",
			oldPoolUnits: 0,
			poolRune:     0.0,
			poolToken:    0.0,
			stakeRune:    0.0,
			stakeToken:   100,
			expectedErr:  errors.New("total RUNE in the pool is zero"),
		},
		{
			name:         "first-stake-zero-token",
			oldPoolUnits: 0,
			poolRune:     0.0,
			poolToken:    0.0,
			stakeRune:    100,
			stakeToken:   0.0,
			expectedErr:  errors.New("total token in the pool is zero"),
		},
		{
			name:         "first-stake",
			oldPoolUnits: 0,
			poolRune:     0.0,
			poolToken:    0.0,
			stakeRune:    100,
			stakeToken:   100,
			poolUnits:    100,
			stakerUnits:  100,
			expectedErr:  nil,
		},
		{
			name:         "second-stake",
			oldPoolUnits: 500.0,
			poolRune:     500.0,
			poolToken:    500.0,
			stakeRune:    345,
			stakeToken:   234,
			poolUnits:    787.0168486,
			stakerUnits:  287.016849,
			expectedErr:  nil,
		},
	}

	for _, item := range inputs {
		poolUnits, stakerUnits, err := calculatePoolUnits(item.oldPoolUnits, item.poolRune, item.poolToken, item.stakeRune, item.stakeToken)
		if item.expectedErr == nil {
			c.Assert(err, IsNil)
		} else {
			c.Assert(err.Error(), Equals, item.expectedErr.Error())
		}
		c.Check(round(item.poolUnits), Equals, round(poolUnits))
		c.Check(round(item.stakerUnits), Equals, round(stakerUnits))
	}
}

func (s StakeSuite) TestValidateAmount(c *C) {
	makePoolStaker := func(total int, avg float64) PoolStaker {
		stakers := make([]StakerUnit, total)
		for i, _ := range stakers {
			stakers[i] = StakerUnit{Units: common.NewAmountFromFloat(avg)}
		}

		return PoolStaker{
			TotalUnits: common.NewAmountFromFloat(avg * float64(total)),
			Stakers:    stakers,
		}
	}

	skrs := makePoolStaker(50, 0.001)
	c.Assert(validateStakeAmount(skrs, 0.001, common.NewAmountFromFloat(100)), IsNil)

	skrs = makePoolStaker(150, 0.0002)
	c.Assert(validateStakeAmount(skrs, 0.0001, common.NewAmountFromFloat(100)), NotNil)
	c.Assert(validateStakeAmount(skrs, 0.0002, common.NewAmountFromFloat(100)), NotNil)
	c.Assert(validateStakeAmount(skrs, 0.0010, common.NewAmountFromFloat(100)), IsNil)

	skrs = makePoolStaker(300, 0.001)
	c.Assert(validateStakeAmount(skrs, 0.0001, common.NewAmountFromFloat(100)), NotNil)
	c.Assert(validateStakeAmount(skrs, 0.002, common.NewAmountFromFloat(100)), NotNil)
	c.Assert(validateStakeAmount(skrs, 0.004, common.NewAmountFromFloat(100)), IsNil)
}

// TestValidateStakeMessage
func (StakeSuite) TestValidateStakeMessage(c *C) {
	ps := NewMockInMemoryPoolStorage()
	ctx := GetCtx("test")
	txId, err := common.NewTxID("4D60A73FEBD42592DB697EF1DA020A214EC3102355D0E1DD07B18557321B106X")
	if nil != err {
		c.Errorf("fail to create tx id,%s", err)
	}
	bnbAddress, err := common.NewBnbAddress("tbnb1c2yvdphs674vlkp2s2e68cw89garykgau2c8vx")
	if nil != err {
		c.Errorf("fail to create bnb address,%s", err)
	}
	c.Assert(validateStakeMessage(ctx, ps, common.Ticker(""), common.NewAmountFromFloat(100), common.NewAmountFromFloat(100), txId, bnbAddress), NotNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BNBTicker, common.Amount(""), common.NewAmountFromFloat(100), txId, bnbAddress), NotNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BNBTicker, common.NewAmountFromFloat(100), common.Amount(""), txId, bnbAddress), NotNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BNBTicker, common.NewAmountFromFloat(100), common.NewAmountFromFloat(100), common.TxID(""), bnbAddress), NotNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BNBTicker, common.NewAmountFromFloat(100), common.NewAmountFromFloat(100), txId, common.NoBnbAddress), NotNil)
	c.Assert(validateStakeMessage(ctx, ps, common.BNBTicker, common.NewAmountFromFloat(100), common.NewAmountFromFloat(100), txId, bnbAddress), NotNil)
	ps.SetPool(ctx, Pool{
		BalanceRune:  common.NewAmountFromFloat(100),
		BalanceToken: common.NewAmountFromFloat(100),
		Ticker:       common.BNBTicker,
		PoolUnits:    common.NewAmountFromFloat(100),
		PoolAddress:  bnbAddress,
		Status:       PoolEnabled,
	})
	c.Assert(validateStakeMessage(ctx, ps, common.BNBTicker, common.NewAmountFromFloat(100), common.NewAmountFromFloat(100), txId, bnbAddress), Equals, nil)
}

// TestStake test stake func
func (StakeSuite) TestStake(c *C) {
	ps := NewMockInMemoryPoolStorage()
	ctx := GetCtx("test")
	txId, err := common.NewTxID("4D60A73FEBD42592DB697EF1DA020A214EC3102355D0E1DD07B18557321B106X")
	if nil != err {
		c.Errorf("fail to create tx id,%s", err)
	}
	bnbAddress, err := common.NewBnbAddress("tbnb1c2yvdphs674vlkp2s2e68cw89garykgau2c8vx")
	if nil != err {
		c.Errorf("fail to create bnb address,%s", err)
	}
	_, err = stake(ctx, ps, common.Ticker(""), common.NewAmountFromFloat(100), common.NewAmountFromFloat(100), bnbAddress, txId)
	c.Assert(err, NotNil)
	ps.SetPool(ctx, Pool{
		BalanceRune:  common.NewAmountFromFloat(-100),
		BalanceToken: common.NewAmountFromFloat(100),
		Ticker:       common.BNBTicker,
		PoolUnits:    common.NewAmountFromFloat(100),
		PoolAddress:  bnbAddress,
		Status:       PoolEnabled,
	})
	_, err = stake(ctx, ps, common.BNBTicker, common.NewAmountFromFloat(100), common.NewAmountFromFloat(100), bnbAddress, txId)
	c.Assert(err, NotNil)
	ps.SetPool(ctx, Pool{
		BalanceRune:  common.NewAmountFromFloat(100),
		BalanceToken: common.NewAmountFromFloat(100),
		Ticker:       notExistPoolStakerTicker,
		PoolUnits:    common.NewAmountFromFloat(100),
		PoolAddress:  bnbAddress,
		Status:       PoolEnabled,
	})
	_, err = stake(ctx, ps, notExistPoolStakerTicker, common.NewAmountFromFloat(100), common.NewAmountFromFloat(100), bnbAddress, txId)
	c.Assert(err, NotNil)
	ps.SetPool(ctx, Pool{
		BalanceRune:  common.NewAmountFromFloat(100),
		BalanceToken: common.NewAmountFromFloat(100),
		Ticker:       common.BNBTicker,
		PoolUnits:    common.NewAmountFromFloat(100),
		PoolAddress:  bnbAddress,
		Status:       PoolEnabled,
	})
	makePoolStaker := func(total int, avg float64) PoolStaker {
		stakers := make([]StakerUnit, total)
		for i, _ := range stakers {
			stakers[i] = StakerUnit{Units: common.NewAmountFromFloat(avg)}
		}

		return PoolStaker{
			TotalUnits: common.NewAmountFromFloat(avg * float64(total)),
			Stakers:    stakers,
		}
	}
	skrs := makePoolStaker(150, 0.0002)
	ps.SetPoolStaker(ctx, common.BNBTicker, skrs)
	_, err = stake(ctx, ps, common.BNBTicker, common.NewAmountFromFloat(1), common.NewAmountFromFloat(1), bnbAddress, txId)
	c.Assert(err, NotNil)

	_, err = stake(ctx, ps, common.BNBTicker, common.NewAmountFromFloat(100), common.NewAmountFromFloat(100), notExistStakerPoolAddr, txId)
	c.Assert(err, NotNil)
	ps.SetPool(ctx, Pool{
		BalanceRune:  common.NewAmountFromFloat(100),
		BalanceToken: common.NewAmountFromFloat(100),
		Ticker:       common.BNBTicker,
		PoolUnits:    common.NewAmountFromFloat(100),
		PoolAddress:  bnbAddress,
		Status:       PoolEnabled,
	})
	_, err = stake(ctx, ps, common.BNBTicker, common.NewAmountFromFloat(100), common.NewAmountFromFloat(100), bnbAddress, txId)
	c.Assert(err, IsNil)
	p := ps.GetPool(ctx, common.BNBTicker)

	c.Check(p.PoolUnits.Equals(common.NewAmountFromFloat(200)), Equals, true)
}
