package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"
)

type YggdrasilSuite struct{}

var _ = Suite(&YggdrasilSuite{})

func (s YggdrasilSuite) TestGetHoldings(c *C) {
	pk := GetRandomPubKey()
	ygg := NewYggdrasil(pk)
	ygg.AddFunds(
		common.Coins{
			common.NewCoin(common.RuneAsset(), sdk.NewUint(50*common.One)),
			common.NewCoin(common.BNBAsset, sdk.NewUint(50*common.One)),
		},
	)

	p := NewPool()
	p.Asset = common.BNBAsset
	p.BalanceRune = sdk.NewUint(1000 * common.One)
	p.BalanceAsset = sdk.NewUint(500 * common.One)

	w := getHandlerTestWrapper(c, 1, true, false)
	w.keeper.SetPool(w.ctx, p)

	amt := getHoldingsValue(w.ctx, w.keeper, ygg)
	expected := sdk.NewUint(150 * common.One).Uint64()
	c.Check(
		amt.Uint64(),
		Equals,
		expected,
		Commentf("%d vs %d", amt.Uint64(), expected),
	)
}

func (s YggdrasilSuite) TestCalcTopUp(c *C) {
	pk := GetRandomPubKey()
	ygg := NewYggdrasil(pk)
	ygg.AddFunds(
		common.Coins{
			common.NewCoin(common.RuneAsset(), sdk.NewUint(50*common.One)),
			common.NewCoin(common.BNBAsset, sdk.NewUint(50*common.One)),
		},
	)

	w := getHandlerTestWrapper(c, 1, true, false)
	addr, err := pk.GetThorAddress()
	c.Assert(err, IsNil)
	w.keeper.SetNodeAccount(w.ctx, NodeAccount{
		NodeAddress: addr,
		Bond:        sdk.NewUint(1500 * common.One),
	})
	p := NewPool()
	p.Asset = common.BNBAsset
	p.BalanceRune = sdk.NewUint(1000 * common.One)
	p.BalanceAsset = sdk.NewUint(500 * common.One)
	w.keeper.SetPool(w.ctx, p)

	p = NewPool()
	p.Asset = common.BTCAsset
	p.BalanceRune = sdk.NewUint(3000 * common.One)
	p.BalanceAsset = sdk.NewUint(225 * common.One)
	w.keeper.SetPool(w.ctx, p)

	target := sdk.NewUint(200 * common.One)
	coins, err := calculateTopUpYgg(w.ctx, w.keeper, target, ygg)
	c.Assert(err, IsNil)
	c.Assert(coins, HasLen, 3)
	c.Check(coins[0].Asset.String(), Equals, common.BNBAsset.String())
	c.Check(coins[1].Asset.String(), Equals, common.BTCAsset.String())
	c.Check(coins[2].Asset.String(), Equals, common.RuneAsset().String())
}
