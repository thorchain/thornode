package common

import (
	. "gopkg.in/check.v1"
)

type CoinSuite struct{}

var _ = Suite(&CoinSuite{})

func (s CoinSuite) TestCoin(c *C) {
	coin := NewCoin("bnb", "2.3")
	c.Check(coin.Denom.Equals(Ticker("BNB")), Equals, true)
	c.Check(coin.Amount.Equals(Amount("2.3")), Equals, true)

	coin = NewCoin("bnb", "-457")
	c.Check(coin.Denom.Equals(Ticker("BNB")), Equals, true)
	c.Check(coin.Amount.Equals(ZeroAmount), Equals, true)
}
