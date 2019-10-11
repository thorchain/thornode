package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type MsgNextPoolTestSuite struct{}

var _ = Suite(&MsgNextPoolTestSuite{})

func (MsgNextPoolTestSuite) TestMsgNextPool(c *C) {
	bnb, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)
	addr, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	c.Check(addr.Empty(), Equals, false)
	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	msgNextPool := NewMsgNextPoolAddress(txID, bnb, addr)
	c.Assert(msgNextPool.Type(), Equals, "set_next_pooladdress")
	EnsureMsgBasicCorrect(msgNextPool, c)
	msgNextPool1 := NewMsgNextPoolAddress("", bnb, addr)
	c.Assert(msgNextPool1.ValidateBasic(), NotNil)
	msgNextPool2 := NewMsgNextPoolAddress(txID, "", addr)
	c.Assert(msgNextPool2.ValidateBasic(), NotNil)
}
