package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type MsgObservedTxInSuite struct{}

var _ = Suite(&MsgObservedTxInSuite{})

func (s *MsgObservedTxInSuite) TestMsgObservedTxIn(c *C) {
	txs := common.Txs{GetRandomTx()}
	acc := GetRandomBech32Addr()

	m := NewMsgObservedTxIn(txs, acc)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_observed_txin")

	m1 := NewMsgObservedTxIn(nil, acc)
	c.Assert(m1.ValidateBasic(), NotNil)
	m2 := NewMsgObservedTxIn(txs, sdk.AccAddress{})
	c.Assert(m2.ValidateBasic(), NotNil)
}
