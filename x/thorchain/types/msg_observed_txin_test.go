package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type MsgObservedTxInSuite struct{}

var _ = Suite(&MsgObservedTxInSuite{})

func (s *MsgObservedTxInSuite) TestMsgObservedTxIn(c *C) {
	tx := NewObservedTx(GetRandomTx(), sdk.NewUint(55), GetRandomPubKey())
	acc := GetRandomBech32Addr()

	m := NewMsgObservedTxIn(ObservedTxs{tx}, acc)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_observed_txin")

	m1 := NewMsgObservedTxIn(nil, acc)
	c.Assert(m1.ValidateBasic(), NotNil)
	m2 := NewMsgObservedTxIn(ObservedTxs{tx}, sdk.AccAddress{})
	c.Assert(m2.ValidateBasic(), NotNil)
}
