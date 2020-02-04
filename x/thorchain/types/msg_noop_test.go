package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type MsgNoopSuite struct{}

var _ = Suite(&MsgNoopSuite{})

func (MsgNoopSuite) TestMsgNoop(c *C) {
	addr := GetRandomBech32Addr()
	c.Check(addr.Empty(), Equals, false)
	tx := ObservedTx{
		Tx:             GetRandomTx(),
		Status:         Done,
		OutHashes:      nil,
		BlockHeight:    1,
		Signers:        []sdk.AccAddress{addr},
		ObservedPubKey: GetRandomPubKey(),
	}
	m := NewMsgNoOp(tx, addr)
	c.Check(m.ValidateBasic(), IsNil)
	c.Check(m.Type(), Equals, "set_noop")
	EnsureMsgBasicCorrect(m, c)
	mEmpty := NewMsgNoOp(tx, sdk.AccAddress{})
	c.Assert(mEmpty.ValidateBasic(), NotNil)
}
