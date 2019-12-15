package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type MsgTssPoolSuite struct{}

var _ = Suite(&MsgTssPoolSuite{})

func (s *MsgTssPoolSuite) TestMsgTssPool(c *C) {
	var err error
	pk := GetRandomPubKey()
	tx := NewObservedTx(GetRandomTx(), sdk.NewUint(55), pk)
	acc := GetRandomBech32Addr()
	tx.Tx.ToAddress, err = pk.GetAddress(tx.Tx.Coins[0].Asset.Chain)
	c.Assert(err, IsNil)

	m := NewMsgObservedTxIn(ObservedTxs{tx}, acc)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_observed_txin")

	m1 := NewMsgObservedTxIn(nil, acc)
	c.Assert(m1.ValidateBasic(), NotNil)
	m2 := NewMsgObservedTxIn(ObservedTxs{tx}, sdk.AccAddress{})
	c.Assert(m2.ValidateBasic(), NotNil)
}
