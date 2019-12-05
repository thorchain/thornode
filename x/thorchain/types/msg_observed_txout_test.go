package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type MsgObservedTxOutSuite struct{}

var _ = Suite(&MsgObservedTxOutSuite{})

func (s *MsgObservedTxOutSuite) TestMsgObservedTxOut(c *C) {
	tx := NewObservedTx(GetRandomTx(), sdk.NewUint(55), GetRandomPubKey())
	acc := GetRandomBech32Addr()

	m := NewMsgObservedTxOut(ObservedTxs{tx}, acc)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_observed_txout")

	m1 := NewMsgObservedTxOut(nil, acc)
	c.Assert(m1.ValidateBasic(), NotNil)
	m2 := NewMsgObservedTxOut(ObservedTxs{tx}, sdk.AccAddress{})
	c.Assert(m2.ValidateBasic(), NotNil)
}
