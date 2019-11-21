package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"
)

type YggdrasilSuite struct{}

var _ = Suite(&YggdrasilSuite{})

func (s *YggdrasilSuite) TestYggdrasil(c *C) {
	pk := GetRandomPubKey()

	ygg := Yggdrasil{}
	c.Check(ygg.IsEmpty(), Equals, true)
	c.Check(ygg.IsValid(), NotNil)

	ygg = NewYggdrasil(pk)
	c.Check(ygg.PubKey.Equals(pk), Equals, true)
	c.Check(ygg.HasFunds(), Equals, false)
	c.Check(ygg.IsEmpty(), Equals, false)
	c.Check(ygg.IsValid(), IsNil)

	coins := common.Coins{
		common.NewCoin(common.BNBAsset, sdk.NewUint(500*common.One)),
		common.NewCoin(common.BTCAsset, sdk.NewUint(400*common.One)),
	}

	ygg.AddFunds(coins)
	c.Check(ygg.HasFunds(), Equals, true)
	c.Check(ygg.GetCoin(common.BNBAsset).Amount.Equal(sdk.NewUint(500*common.One)), Equals, true)
	c.Check(ygg.GetCoin(common.BTCAsset).Amount.Equal(sdk.NewUint(400*common.One)), Equals, true)
	ygg.AddFunds(coins)
	c.Check(ygg.GetCoin(common.BNBAsset).Amount.Equal(sdk.NewUint(1000*common.One)), Equals, true)
	c.Check(ygg.GetCoin(common.BTCAsset).Amount.Equal(sdk.NewUint(800*common.One)), Equals, true)
	ygg.SubFunds(coins)
	c.Check(ygg.GetCoin(common.BNBAsset).Amount.Equal(sdk.NewUint(500*common.One)), Equals, true)
	c.Check(ygg.GetCoin(common.BTCAsset).Amount.Equal(sdk.NewUint(400*common.One)), Equals, true)
	ygg.SubFunds(coins)
	c.Check(ygg.GetCoin(common.BNBAsset).Amount.Equal(sdk.ZeroUint()), Equals, true)
	c.Check(ygg.GetCoin(common.BTCAsset).Amount.Equal(sdk.ZeroUint()), Equals, true)
	c.Check(ygg.HasFunds(), Equals, false)
	ygg.SubFunds(coins)
	c.Check(ygg.GetCoin(common.BNBAsset).Amount.Equal(sdk.ZeroUint()), Equals, true)
	c.Check(ygg.GetCoin(common.BTCAsset).Amount.Equal(sdk.ZeroUint()), Equals, true)
	c.Check(ygg.HasFunds(), Equals, false)
}
