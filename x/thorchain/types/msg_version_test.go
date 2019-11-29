package types

import (
	"github.com/blang/semver"
	. "gopkg.in/check.v1"
)

type MsgSetVersionSuite struct{}

var _ = Suite(&MsgSetVersionSuite{})

func (MsgSetVersionSuite) TestMsgSetVersionSuite(c *C) {
	acc1 := GetRandomBech32Addr()
	c.Assert(acc1.Empty(), Equals, false)
	msg := NewMsgSetVersion(semver.MustParse("2.0.0-beta"), acc1)
	c.Assert(msg.Route(), Equals, RouterKey)
	c.Assert(msg.Type(), Equals, "set_version")
	c.Assert(msg.ValidateBasic(), IsNil)
	c.Assert(len(msg.GetSignBytes()) > 0, Equals, true)
	c.Assert(msg.GetSigners(), NotNil)
	c.Assert(msg.GetSigners()[0].String(), Equals, acc1.String())
}
