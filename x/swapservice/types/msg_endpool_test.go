package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type MsgEndPoolTestSuite struct{}

var _ = Suite(&MsgEndPoolTestSuite{})

func (MsgEndPoolTestSuite) TestMsgEndPool(c *C) {
	ticker := common.BNBTicker
	bnb, err := common.NewAddress("bnb1xlvns0n2mxh77mzaspn2hgav4rr4m8eerfju38")
	c.Assert(err, IsNil)
	addr, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	c.Check(addr.Empty(), Equals, false)
	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	msgEndPool := NewMsgEndPool(ticker, bnb, txID, addr)
	c.Assert(msgEndPool.Route(), Equals, RouterKey)
	c.Assert(msgEndPool.Type(), Equals, "set_poolend")
	c.Assert(msgEndPool.ValidateBasic(), IsNil)
	c.Assert(len(msgEndPool.GetSignBytes()) > 0, Equals, true)
	c.Assert(msgEndPool.GetSigners(), NotNil)
	c.Assert(msgEndPool.GetSigners()[0].String(), Equals, addr.String())

	errEndPool := NewMsgEndPool("", bnb, txID, addr)
	c.Assert(errEndPool.ValidateBasic(), NotNil)
	errEndPool1 := NewMsgEndPool(common.RuneA1FTicker, bnb, txID, addr)
	c.Assert(errEndPool1.ValidateBasic(), NotNil)
	errEndPool2 := NewMsgEndPool(common.BNBTicker, bnb, "", addr)
	c.Assert(errEndPool2.ValidateBasic(), NotNil)
	errEndPool3 := NewMsgEndPool(common.BNBTicker, "", txID, addr)
	c.Assert(errEndPool3.ValidateBasic(), NotNil)

}
