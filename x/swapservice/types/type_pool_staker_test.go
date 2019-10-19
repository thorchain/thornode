package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type PoolStakerSuite struct{}

var _ = Suite(&PoolStakerSuite{})

func (PoolStakerSuite) TestPoolStaker(c *C) {
	bnb, err := common.NewAddress("bnb1xlvns0n2mxh77mzaspn2hgav4rr4m8eerfju38")
	c.Assert(err, IsNil)

	poolStaker := NewPoolStaker(common.BNBTicker, sdk.NewUint(100))
	c.Assert(poolStaker.Stakers, NotNil)
	stakerUnit := StakerUnit{
		StakerID: bnb,
		Units:    sdk.NewUint(100),
	}
	poolStaker.UpsertStakerUnit(stakerUnit)
	poolStaker.UpsertStakerUnit(stakerUnit)
	c.Logf("poolstakers:%s", poolStaker)
	c.Assert(poolStaker.Stakers, NotNil)
	c.Check(len(poolStaker.Stakers), Equals, 1)
	newStakerUnit := poolStaker.GetStakerUnit(bnb)
	c.Check(newStakerUnit.StakerID, Equals, bnb)
	c.Check(newStakerUnit.Units.Equal(sdk.NewUint(100)), Equals, true)

	poolStaker.RemoveStakerUnit(bnb)
	c.Check(len(poolStaker.Stakers), Equals, 0)

	bnb1, err := common.NewAddress("tbnb1yycn4mh6ffwpjf584t8lpp7c27ghu03gpvqkfj")
	c.Assert(err, IsNil)
	stakerUnit1 := poolStaker.GetStakerUnit(bnb1)
	c.Check(stakerUnit1.Units.IsZero(), Equals, true)

}
