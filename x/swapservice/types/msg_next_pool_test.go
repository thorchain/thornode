package types

import (
	. "gopkg.in/check.v1"
)

type MsgNextPoolTestSuite struct{}

var _ = Suite(&MsgNextPoolTestSuite{})

func (MsgNextPoolTestSuite) TestMsgNextPool(c *C) {
	sender := GetRandomBNBAddress()
	bnb := GetRandomBNBAddress()
	addr := GetRandomBech32Addr()
	c.Check(addr.Empty(), Equals, false)
	txID := GetRandomTxHash()
	msgNextPool := NewMsgNextPoolAddress(txID, bnb, sender, addr)
	c.Assert(msgNextPool.Type(), Equals, "set_next_pooladdress")
	EnsureMsgBasicCorrect(msgNextPool, c)
	msgNextPool1 := NewMsgNextPoolAddress("", bnb, sender, addr)
	c.Assert(msgNextPool1.ValidateBasic(), NotNil)
	msgNextPool2 := NewMsgNextPoolAddress(txID, "", sender, addr)
	c.Assert(msgNextPool2.ValidateBasic(), NotNil)
	msgNextPool3 := NewMsgNextPoolAddress(txID, bnb, "", addr)
	c.Assert(msgNextPool3.ValidateBasic(), NotNil)
}
