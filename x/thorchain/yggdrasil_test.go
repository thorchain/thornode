package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
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

func (s YggdrasilSuite) TestCalcTargetAmounts2(c *C) {
	// Adding specific test per PR request
	// https://gitlab.com/thorchain/thornode/merge_requests/246#note_241913460
	var pools []Pool
	p := NewPool()
	p.Asset = common.BNBAsset
	p.BalanceRune = sdk.NewUint(1000000 * common.One)
	p.BalanceAsset = sdk.NewUint(1 * common.One)
	pools = append(pools, p)

	totalBond := sdk.NewUint(3000000 * common.One)
	bond := sdk.NewUint(1000000 * common.One)
	coins, err := calcTargetYggCoins(pools, bond, totalBond)
	c.Assert(err, IsNil)
	c.Assert(coins, HasLen, 2)
	c.Check(coins[0].Asset.String(), Equals, common.BNBAsset.String())
	c.Check(coins[0].Amount.Uint64(), Equals, sdk.NewUint(0.16666667*common.One).Uint64(), Commentf("%d vs %d", coins[0].Amount.Uint64(), sdk.NewUint(0.16666667*common.One).Uint64()))
	c.Check(coins[1].Asset.String(), Equals, common.RuneAsset().String())
	c.Check(coins[1].Amount.Uint64(), Equals, sdk.NewUint(166666.66666667*common.One).Uint64(), Commentf("%d vs %d", coins[1].Amount.Uint64(), sdk.NewUint(166666.66666667*common.One).Uint64()))
}
