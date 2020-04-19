package types

import (
	. "gopkg.in/check.v1"
)

type MsgSetIPAddressSuite struct{}

var _ = Suite(&MsgSetIPAddressSuite{})

func (MsgSetIPAddressSuite) TestMsgSetIPAddressSuite(c *C) {
	acc1 := GetRandomBech32Addr()
	c.Assert(acc1.Empty(), Equals, false)
	msg := NewMsgSetIPAddress("192.168.0.1", acc1)
	c.Assert(msg.Route(), Equals, RouterKey)
	c.Assert(msg.Type(), Equals, "set_ip_address")
	c.Assert(msg.ValidateBasic(), IsNil)
	c.Assert(len(msg.GetSignBytes()) > 0, Equals, true)
	c.Assert(msg.GetSigners(), NotNil)
	c.Assert(msg.GetSigners()[0].String(), Equals, acc1.String())
}
