package thorchain

import (
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperPoolSuite struct{}

var _ = Suite(&KeeperPoolSuite{})

func (s *KeeperPoolSuite) TestPool(c *C) {
	ctx, k := setupKeeperForTest(c)

	pool := NewPool()
	pool.Asset = common.BNBAsset

	k.SetPool(ctx, pool)
	pool, err := k.GetPool(ctx, pool.Asset)
	c.Assert(err, IsNil)
	c.Check(pool.Asset.Equals(common.BNBAsset), Equals, true)
	c.Check(k.PoolExist(ctx, common.BNBAsset), Equals, true)
	c.Check(k.PoolExist(ctx, common.BTCAsset), Equals, false)
}
