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
	c.Check(coin.IsValid(), IsNil)
	c.Check(coin.IsEmpty(), Equals, false)
	c.Check(NoCoin.IsEmpty(), Equals, true)

	c.Check(coin.IsNative(), Equals, false)
	_, err := coin.Native()
	c.Assert(err, NotNil)
	coin = NewCoin(RuneNative, sdk.NewUint(230))
	c.Check(coin.IsNative(), Equals, true)
	sdkCoin, err := coin.Native()
	c.Assert(err, IsNil)
	c.Check(sdkCoin.Denom, Equals, "rune")
	c.Check(sdkCoin.Amount.Equal(sdk.NewInt(230)), Equals, true)
}
