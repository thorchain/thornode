package thorchain

import (
	. "gopkg.in/check.v1"
)

type KeeperBanSuite struct{}

var _ = Suite(&KeeperBanSuite{})

func (s *KeeperBanSuite) TestBanVoter(c *C) {
	ctx, k := setupKeeperForTest(c)

	addr := GetRandomBech32Addr()
	voter := NewBanVoter(addr)

	k.SetBanVoter(ctx, voter)
	voter, err := k.GetBanVoter(ctx, addr)
	c.Assert(err, IsNil)
	c.Check(voter.NodeAddress.Equals(addr), Equals, true)
}
