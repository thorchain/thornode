package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type MsgObservedTxInSuite struct{}

var _ = Suite(&MsgObservedTxInSuite{})

func (s *MsgObservedTxInSuite) TestMsgObservedTxIn(c *C) {
	var err error
	pk := GetRandomPubKey()
	tx := NewObservedTx(GetRandomTx(), 55, pk)
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

	// will not accept observations with pre-determined signers. This is
	// important to ensure an observer can fake signers from other node accounts
	// *IMPORTANT* DON'T REMOVE THIS CHECK
	tx.Signers = append(tx.Signers, GetRandomBech32Addr())
	m3 := NewMsgObservedTxIn(ObservedTxs{tx}, acc)
	c.Assert(m3.ValidateBasic(), NotNil)
}
