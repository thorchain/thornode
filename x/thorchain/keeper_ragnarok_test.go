package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type KeeperRagnarokSuite struct{}

var _ = Suite(&KeeperRagnarokSuite{})

func (s *KeeperRagnarokSuite) TestRagnarok(c *C) {
	ctx, k := setupKeeperForTest(c)

	height, err := k.GetRagnarokBlockHeight(ctx)
	c.Assert(err, IsNil)
	c.Assert(height.Uint64(), Equals, uint64(0))

	k.SetRagnarokBlockHeight(ctx, sdk.NewUint(45))

	height, err = k.GetRagnarokBlockHeight(ctx)
	c.Assert(err, IsNil)
	c.Assert(height.Uint64(), Equals, uint64(45))
}
