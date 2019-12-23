package thorchain

import (
	. "gopkg.in/check.v1"
)

type KeeperRagnarokSuite struct{}

var _ = Suite(&KeeperRagnarokSuite{})

func (s *KeeperRagnarokSuite) TestRagnarok(c *C) {
	ctx, k := setupKeeperForTest(c)

	height, err := k.GetRagnarokBlockHeight(ctx)
	c.Assert(err, IsNil)
	c.Assert(height, Equals, int64(0))
	c.Check(k.RagnarokInProgress(ctx), Equals, false)

	k.SetRagnarokBlockHeight(ctx, 45)

	height, err = k.GetRagnarokBlockHeight(ctx)
	c.Assert(err, IsNil)
	c.Assert(height, Equals, int64(45))
	c.Check(k.RagnarokInProgress(ctx), Equals, true)
}
