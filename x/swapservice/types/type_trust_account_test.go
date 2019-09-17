package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type TrustAccountSuite struct{}

var _ = Suite(&TrustAccountSuite{})

func (TrustAccountSuite) TestTrustAccount(c *C) {
	bnb, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)
	addr, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	c.Check(addr.Empty(), Equals, false)
	trustAccount, err := NewTrustAccount("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v", bnb)
	c.Assert(err, IsNil)
	c.Assert(trustAccount.BepAddress.Equals(addr), Equals, true)
	c.Assert(trustAccount.BnbAddress, Equals, bnb)
	c.Log(trustAccount.String())
}
