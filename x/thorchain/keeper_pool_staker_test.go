package thorchain

import (
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperPoolStakerSuite struct{}

var _ = Suite(&KeeperPoolStakerSuite{})

func (s *KeeperPoolStakerSuite) TestPoolStaker(c *C) {
	ctx, k := setupKeeperForTest(c)

	chains, err := k.GetChains(ctx)
	c.Assert(err, IsNil)
	c.Assert(chains, HasLen, 0)

	chains = append(chains, common.BNBChain)
	k.SetChains(ctx, chains)

	chains, err = k.GetChains(ctx)
	c.Assert(err, IsNil)
	c.Assert(chains, HasLen, 1)
	c.Check(chains[0].IsBNB(), Equals, true)
}
