package thorchain

import (
	. "gopkg.in/check.v1"
)

type KeeperPoolAddressesSuite struct{}

var _ = Suite(&KeeperPoolAddressesSuite{})

func (s *KeeperPoolAddressesSuite) TestPoolAddresses(c *C) {
	var err error
	ctx, k := setupKeeperForTest(c)

	prev := GetRandomPoolPubKeys()
	curr := GetRandomPoolPubKeys()
	next := GetRandomPoolPubKeys()

	pa := NewPoolAddresses(prev, curr, next, 5, 10)
	k.SetPoolAddresses(ctx, pa)
	newPa, err := k.GetPoolAddresses(ctx)
	c.Assert(err, IsNil)
	c.Check(newPa.Previous[0].Equals(prev[0]), Equals, true)
	c.Check(newPa.Current[0].Equals(curr[0]), Equals, true)
	c.Check(newPa.Next[0].Equals(next[0]), Equals, true)
	c.Check(newPa.RotateAt, Equals, int64(5))
	c.Check(newPa.RotateWindowOpenAt, Equals, int64(10))
}
