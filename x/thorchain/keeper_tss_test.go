package thorchain

import (
	. "gopkg.in/check.v1"
)

type KeeperTssSuite struct{}

var _ = Suite(&KeeperTssSuite{})

func (s *KeeperTssSuite) TestTssVoter(c *C) {
	ctx, k := setupKeeperForTest(c)

	pk := GetRandomPubKey()
	voter := NewTssVoter("hello", nil, pk)

	k.SetTssVoter(ctx, voter)
	voter, err := k.GetTssVoter(ctx, voter.ID)
	c.Assert(err, IsNil)
	c.Check(voter.ID, Equals, "hello")
	c.Check(voter.PoolPubKey.Equals(pk), Equals, true)
}
