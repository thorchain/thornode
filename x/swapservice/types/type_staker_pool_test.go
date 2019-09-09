package types

import (
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type StakerPoolSuite struct{}

var _ = Suite(&StakerPoolSuite{})

func (StakerPoolSuite) TestStakerPool(c *C) {
	bnb, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)

	stakerPool := NewStakerPool(bnb)
	c.Check(stakerPool.StakerID, Equals, bnb)
	c.Assert(stakerPool.PoolUnits, NotNil)

	stakerPoolItem := &StakerPoolItem{
		Ticker: common.BNBTicker,
		Units:  common.NewAmountFromFloat(100),
	}
	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	txID2, err := common.NewTxID("47B4FE474A63DDF79DF2790C1C5162F4C213484750AB8292CFE7342E4B0B40E2")
	c.Assert(err, IsNil)
	stakerPoolItem.AddStakerTxDetail(txID, common.NewAmountFromFloat(100), common.NewAmountFromFloat(100))
	stakerPool.UpsertStakerPoolItem(stakerPoolItem)
	c.Assert(stakerPool.PoolUnits, NotNil)
	c.Assert(len(stakerPool.PoolUnits), Equals, 1)
	stakerPoolItem.AddStakerTxDetail(txID2, common.NewAmountFromFloat(50), common.NewAmountFromFloat(50))
	c.Assert(len(stakerPool.PoolUnits), Equals, 1)
	c.Assert(stakerPoolItem.StakeDetails, NotNil)
	spi := stakerPool.GetStakerPoolItem(common.BNBTicker)
	c.Assert(spi, NotNil)
	c.Assert(len(spi.StakeDetails), Equals, 2)
	stakerPool.RemoveStakerPoolItem(common.RuneA1FTicker)
	stakerPool.RemoveStakerPoolItem(common.BNBTicker)
	c.Assert(len(stakerPool.PoolUnits), Equals, 0)
	stakerPool.UpsertStakerPoolItem(stakerPoolItem)
	stakerPoolItem1 := &StakerPoolItem{
		Ticker: common.RuneB1ATicker,
		Units:  common.NewAmountFromFloat(100),
	}
	stakerPool.UpsertStakerPoolItem(stakerPoolItem1)
	stakerPool.UpsertStakerPoolItem(stakerPoolItem)
	stakerPool.RemoveStakerPoolItem(common.RuneB1ATicker)
	c.Assert(len(stakerPool.PoolUnits), Equals, 1)
	c.Log(stakerPool.String())
	spi1 := stakerPool.GetStakerPoolItem(common.RuneA1FTicker)
	c.Assert(spi1.Units, Equals, common.ZeroAmount)
}
