package types

import (
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type PoolStakerSuite struct{}

var _ = Suite(&PoolStakerSuite{})

func (PoolStakerSuite) TestPoolStaker(c *C) {
	bnb, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)

	poolStaker := NewPoolStaker(common.BNBTicker, common.NewAmountFromFloat(100))
	c.Assert(poolStaker.Stakers, NotNil)
	stakerUnit := StakerUnit{
		StakerID: bnb,
		Units:    common.NewAmountFromFloat(100),
	}
	poolStaker.UpsertStakerUnit(stakerUnit)
	poolStaker.UpsertStakerUnit(stakerUnit)
	c.Logf("poolstakers:%s", poolStaker)
	c.Assert(poolStaker.Stakers, NotNil)
	c.Check(len(poolStaker.Stakers), Equals, 1)
	newStakerUnit := poolStaker.GetStakerUnit(bnb)
	c.Check(newStakerUnit.StakerID, Equals, bnb)
	c.Check(newStakerUnit.Units, Equals, common.NewAmountFromFloat(100))

	poolStaker.RemoveStakerUnit(bnb)
	c.Check(len(poolStaker.Stakers), Equals, 0)

	bnb1, err := common.NewBnbAddress("tbnb1yycn4mh6ffwpjf584t8lpp7c27ghu03gpvqkfj")
	c.Assert(err, IsNil)
	stakerUnit1 := poolStaker.GetStakerUnit(bnb1)
	c.Check(stakerUnit1.Units, Equals, common.ZeroAmount)

}
