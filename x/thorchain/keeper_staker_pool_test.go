package thorchain

import (
	. "gopkg.in/check.v1"
)

type KeeperStakerPoolSuite struct{}

var _ = Suite(&KeeperStakerPoolSuite{})

func (s *KeeperStakerPoolSuite) TestStakerPool(c *C) {
	ctx, k := setupKeeperForTest(c)

	addr := GetRandomBNBAddress()
	staker := NewStakerPool(addr)

	k.SetStakerPool(ctx, staker)
	staker, err := k.GetStakerPool(ctx, addr)
	c.Assert(err, IsNil)
	c.Check(staker.RuneAddress.Equals(addr), Equals, true)
}
