package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"
)

type YggdrasilSuite struct{}

var _ = Suite(&YggdrasilSuite{})

func (s YggdrasilSuite) TestCalcTargetAmounts(c *C) {
	var pools []Pool
	p := NewPool()
	p.Asset = common.BNBAsset
	p.BalanceRune = sdk.NewUint(1000 * common.One)
	p.BalanceAsset = sdk.NewUint(500 * common.One)
	pools = append(pools, p)

	p = NewPool()
	p.Asset = common.BTCAsset
	p.BalanceRune = sdk.NewUint(3000 * common.One)
	p.BalanceAsset = sdk.NewUint(225 * common.One)
	pools = append(pools, p)

	totalBond := sdk.NewUint(8000 * common.One)
	bond := sdk.NewUint(200 * common.One)
	coins, err := calcTargetYggCoins(pools, bond, totalBond)
	c.Assert(err, IsNil)
	c.Assert(coins, HasLen, 3)
	c.Check(coins[0].Asset.String(), Equals, common.BNBAsset.String())
	c.Check(coins[0].Amount.Uint64(), Equals, sdk.NewUint(6.25*common.One).Uint64(), Commentf("%d vs %d", coins[0].Amount.Uint64(), sdk.NewUint(6.25*common.One).Uint64()))
	c.Check(coins[1].Asset.String(), Equals, common.BTCAsset.String())
	c.Check(coins[1].Amount.Uint64(), Equals, sdk.NewUint(2.8125*common.One).Uint64(), Commentf("%d vs %d", coins[1].Amount.Uint64(), sdk.NewUint(2.8125*common.One).Uint64()))
	c.Check(coins[2].Asset.String(), Equals, common.RuneAsset().String())
	c.Check(coins[2].Amount.Uint64(), Equals, sdk.NewUint(50*common.One).Uint64(), Commentf("%d vs %d", coins[2].Amount.Uint64(), sdk.NewUint(50*common.One).Uint64()))
}
