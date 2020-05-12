package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type MsgMimirSuite struct{}

var _ = Suite(&MsgMimirSuite{})

func (MsgMimirSuite) TestMsgMimir(c *C) {
	addr := GetRandomBech32Addr()
	m := NewMsgMimir("key", 12, addr)
	c.Check(m.ValidateBasic(), IsNil)
	c.Check(m.Type(), Equals, "set_mimir_attr")
	EnsureMsgBasicCorrect(m, c)
	mEmpty := NewMsgMimir("", 0, sdk.AccAddress{})
	c.Assert(mEmpty.ValidateBasic(), NotNil)
}
