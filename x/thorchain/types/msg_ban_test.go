package types

import (
	. "gopkg.in/check.v1"
)

type MsgBanSuite struct{}

var _ = Suite(&MsgBanSuite{})

func (MsgBanSuite) TestMsgBanSuite(c *C) {
	acc1 := GetRandomBech32Addr()
	acc2 := GetRandomBech32Addr()
	c.Assert(acc1.Empty(), Equals, false)
	msg := NewMsgBan(acc1, acc2)
	c.Assert(msg.Route(), Equals, RouterKey)
	c.Assert(msg.Type(), Equals, "ban")
	c.Assert(msg.ValidateBasic(), IsNil)
	c.Assert(len(msg.GetSignBytes()) > 0, Equals, true)
	c.Assert(msg.GetSigners(), NotNil)
	c.Assert(msg.GetSigners()[0].String(), Equals, acc2.String())
}
