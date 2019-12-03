package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperLastHeightSuite struct{}

var _ = Suite(&KeeperLastHeightSuite{})

func (s *KeeperLastHeightSuite) TestLastHeight(c *C) {
	ctx, k := setupKeeperForTest(c)

	k.SetLastSignedHeight(ctx, sdk.NewUint(12))
	last, err := k.GetLastSignedHeight(ctx)
	c.Assert(err, IsNil)
	c.Check(last.Uint64(), Equals, uint64(12))

	err = k.SetLastChainHeight(ctx, common.BNBChain, sdk.NewUint(14))
	c.Assert(err, IsNil)
	last, err = k.GetLastChainHeight(ctx, common.BNBChain)
	c.Assert(err, IsNil)
	c.Check(last.Uint64(), Equals, uint64(14))
}
