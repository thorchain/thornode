package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperLiquidityFeesSuite struct{}

var _ = Suite(&KeeperLiquidityFeesSuite{})

func (s *KeeperLiquidityFeesSuite) TestLiquidityFees(c *C) {
	ctx, k := setupKeeperForTest(c)

	ctx = ctx.WithBlockHeight(10)
	height := uint64(ctx.BlockHeight())
	err := k.AddToLiquidityFees(ctx, common.BTCAsset, sdk.NewUint(200))
	c.Assert(err, IsNil)
	err = k.AddToLiquidityFees(ctx, common.BNBAsset, sdk.NewUint(300))
	c.Assert(err, IsNil)

	i, err := k.GetTotalLiquidityFeesInBlock(ctx, height)
	c.Assert(err, IsNil)
	c.Check(i.Uint64(), Equals, uint64(500))

	i, err = k.GetPoolLiquidityFeesInBlock(ctx, height, common.BTCAsset)
	c.Assert(err, IsNil)
	c.Check(i.Uint64(), Equals, uint64(200), Commentf("%d", i.Uint64()))

	i, err = k.GetPoolLiquidityFeesInBlock(ctx, height, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Check(i.Uint64(), Equals, uint64(300), Commentf("%d", i.Uint64()))

	i, err = k.GetTotalLiquidityFees(ctx)
	c.Assert(err, IsNil)
	c.Check(i.Uint64(), Equals, uint64(500))

	i, err = k.GetPoolLiquidityFees(ctx, common.BTCAsset)
	c.Assert(err, IsNil)
	c.Check(i.Uint64(), Equals, uint64(200), Commentf("%d", i.Uint64()))

	i, err = k.GetPoolLiquidityFees(ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Check(i.Uint64(), Equals, uint64(300), Commentf("%d", i.Uint64()))

	ctx = ctx.WithBlockHeight(11)
	err = k.AddToLiquidityFees(ctx, common.BTCAsset, sdk.NewUint(200))
	c.Assert(err, IsNil)
	err = k.AddToLiquidityFees(ctx, common.BNBAsset, sdk.NewUint(300))
	c.Assert(err, IsNil)

	i, err = k.GetTotalLiquidityFees(ctx)
	c.Assert(err, IsNil)
	c.Check(i.Uint64(), Equals, uint64(1000))

	i, err = k.GetPoolLiquidityFees(ctx, common.BTCAsset)
	c.Assert(err, IsNil)
	c.Check(i.Uint64(), Equals, uint64(400), Commentf("%d", i.Uint64()))

	i, err = k.GetPoolLiquidityFees(ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Check(i.Uint64(), Equals, uint64(600), Commentf("%d", i.Uint64()))

}
