package thorchain

import (
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperLastHeightSuite struct{}

var _ = Suite(&KeeperLastHeightSuite{})

func (s *KeeperLastHeightSuite) TestLastHeight_EmptyKeeper(c *C) {
	ctx, k := setupKeeperForTest(c)

	last, err := k.GetLastSignedHeight(ctx)
	c.Assert(err, IsNil)
	c.Check(last, Equals, int64(0))

	last, err = k.GetLastChainHeight(ctx, common.BNBChain)
	c.Assert(err, IsNil)
	c.Check(last, Equals, int64(0))
}

func (s *KeeperLastHeightSuite) TestLastHeight_SetKeeperSingleChain(c *C) {
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

func (s *KeeperLastHeightSuite) TestLastHeight_SetKeeperMultipleChains(c *C) {
	ctx, k := setupKeeperForTest(c)

	err := k.SetLastChainHeight(ctx, common.BTCChain, 23)
	c.Assert(err, IsNil)
	err = k.SetLastChainHeight(ctx, common.BNBChain, 14)
	c.Assert(err, IsNil)
	last, err := k.GetLastChainHeight(ctx, common.BTCChain)
	c.Assert(err, IsNil)
	c.Check(last, Equals, int64(23))
	last, err = k.GetLastChainHeight(ctx, common.BNBChain)
	c.Assert(err, IsNil)
	c.Check(last, Equals, int64(14))
}
