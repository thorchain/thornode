package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type HelperSuite struct{}

var _ = Suite(&HelperSuite{})

func (s *HelperSuite) TestEnableNextPool(c *C) {
	var err error
	ctx, k := setupKeeperForTest(c)

	pool := NewPool()
	pool.Asset = common.BNBAsset
	pool.Status = PoolEnabled
	pool.BalanceRune = sdk.NewUint(100 * common.One)
	pool.BalanceAsset = sdk.NewUint(100 * common.One)
	k.SetPool(ctx, pool)

	pool = NewPool()
	pool.Asset = common.BTCAsset
	pool.Status = PoolBootstrap
	pool.BalanceRune = sdk.NewUint(50 * common.One)
	pool.BalanceAsset = sdk.NewUint(50 * common.One)
	k.SetPool(ctx, pool)

	ethAsset, err := common.NewAsset("ETH.ETH")
	c.Assert(err, IsNil)
	pool = NewPool()
	pool.Asset = ethAsset
	pool.Status = PoolBootstrap
	pool.BalanceRune = sdk.NewUint(40 * common.One)
	pool.BalanceAsset = sdk.NewUint(40 * common.One)
	k.SetPool(ctx, pool)

	xmrAsset, err := common.NewAsset("XMR.XMR")
	c.Assert(err, IsNil)
	pool = NewPool()
	pool.Asset = xmrAsset
	pool.Status = PoolBootstrap
	pool.BalanceRune = sdk.NewUint(40 * common.One)
	pool.BalanceAsset = sdk.NewUint(0 * common.One)
	k.SetPool(ctx, pool)

	// should enable BTC
	enableNextPool(ctx, k)
	pool, err = k.GetPool(ctx, common.BTCAsset)
	c.Check(pool.Status, Equals, PoolEnabled)

	// should enable ETH
	enableNextPool(ctx, k)
	pool, err = k.GetPool(ctx, ethAsset)
	c.Check(pool.Status, Equals, PoolEnabled)

	// should NOT enable XMR, since it has no assets
	enableNextPool(ctx, k)
	pool, err = k.GetPool(ctx, xmrAsset)
	c.Check(pool.Status, Equals, PoolBootstrap)
}
