package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type QueryResPoolsSuite struct{}

var _ = Suite(&QueryResPoolsSuite{})

func (QueryResPoolsSuite) TestQueryResPools(c *C) {
	p := NewPool()
	p.Ticker = common.BNBTicker
	var qrp QueryResPools
	qrp = append(qrp, p)
	c.Check(qrp.String(), Equals, "BNB")

}

func (QueryResPoolsSuite) TestFromSdkCoins(c *C) {
	coins := sdk.NewCoins(
		sdk.NewCoin("rune", sdk.NewInt(100)))
	statechainCoins, err := FromSdkCoins(coins)
	c.Assert(err, IsNil)
	c.Assert(statechainCoins, NotNil)
	c.Check(len(statechainCoins) > 0, Equals, true)
	c.Check(statechainCoins[0].Denom, Equals, common.Ticker("RUNE"))
	c.Check(statechainCoins[0].Amount.Uint64(), Equals, uint64(100))

	coins1 := sdk.Coins{
		sdk.Coin{
			Denom:  "BN",
			Amount: sdk.NewInt(100),
		},
	}
	statechainCoins1, err := FromSdkCoins(coins1)
	c.Assert(err, NotNil)
	c.Assert(statechainCoins1, IsNil)
}
