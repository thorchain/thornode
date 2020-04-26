package types

import (
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type StakerSuite struct{}

var _ = Suite(&StakerSuite{})

func (StakerSuite) TestStaker(c *C) {
	staker := Staker{
		Asset:           common.BNBAsset,
		RuneAddress:     GetRandomBNBAddress(),
		AssetAddress:    GetRandomBTCAddress(),
		LastStakeHeight: 12,
	}
	c.Check(staker.IsValid(), IsNil)
}
