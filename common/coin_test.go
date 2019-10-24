package common

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type CoinSuite struct{}

var _ = Suite(&CoinSuite{})

func (s CoinSuite) TestCoin(c *C) {
	coin := NewCoin(BNBAsset, sdk.NewUint(230000000))
	c.Check(coin.Asset.Equals(BNBAsset), Equals, true)
	c.Check(coin.Amount.Uint64(), Equals, uint64(230000000))
	c.Check(coin.Valid(), IsNil)
}
