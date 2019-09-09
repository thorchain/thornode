package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type MsgNoopSuite struct{}

var _ = Suite(&MsgNoopSuite{})

func (MsgNoopSuite) TestMsgNoop(c *C) {
	addr, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	c.Check(addr.Empty(), Equals, false)
	m := NewMsgNoOp(addr)
	c.Check(m.ValidateBasic(), IsNil)
	c.Check(m.Type(), Equals, "set_noop")
	EnsureMsgBasicCorrect(m, c)
	mEmpty := NewMsgNoOp(sdk.AccAddress{})
	c.Assert(mEmpty.ValidateBasic(), NotNil)
}
