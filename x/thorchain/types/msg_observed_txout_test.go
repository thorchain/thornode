package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type MsgObservedTxOutSuite struct{}

var _ = Suite(&MsgObservedTxOutSuite{})

func (s *MsgObservedTxOutSuite) TestMsgObservedTxOut(c *C) {
	txs := common.Txs{GetRandomTx()}
	acc := GetRandomBech32Addr()

	m := NewMsgObservedTxOut(txs, acc)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_observed_txout")

	m1 := NewMsgObservedTxOut(nil, acc)
	c.Assert(m1.ValidateBasic(), NotNil)
	m2 := NewMsgObservedTxOut(txs, sdk.AccAddress{})
	c.Assert(m2.ValidateBasic(), NotNil)
}
