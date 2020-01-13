package thorchain

import (
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperLastHeightSuite struct{}

var _ = Suite(&KeeperLastHeightSuite{})

func (s *KeeperLastHeightSuite) TestLastHeight(c *C) {
	ctx, k := setupKeeperForTest(c)

	k.SetLastSignedHeight(ctx, 12)
	last, err := k.GetLastSignedHeight(ctx)
	c.Assert(err, IsNil)
	c.Check(last, Equals, int64(12))

	err = k.SetLastChainHeight(ctx, common.BNBChain, 14)
	c.Assert(err, IsNil)
	last, err = k.GetLastChainHeight(ctx, common.BNBChain)
	c.Assert(err, IsNil)
	c.Check(last, Equals, int64(14))
}
