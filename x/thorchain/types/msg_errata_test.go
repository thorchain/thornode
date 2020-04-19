package types

import (
	common "gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type MsgErrataTxSuite struct{}

var _ = Suite(&MsgErrataTxSuite{})

func (MsgErrataTxSuite) TestMsgErrataTxSuite(c *C) {
	txID := GetRandomTxHash()
	acc1 := GetRandomBech32Addr()
	c.Assert(acc1.Empty(), Equals, false)
	msg := NewMsgErrataTx(txID, common.BNBChain, acc1)
	c.Assert(msg.Route(), Equals, RouterKey)
	c.Assert(msg.Type(), Equals, "errata_tx")
	c.Assert(msg.ValidateBasic(), IsNil)
	c.Assert(len(msg.GetSignBytes()) > 0, Equals, true)
	c.Assert(msg.GetSigners(), NotNil)
	c.Assert(msg.GetSigners()[0].String(), Equals, acc1.String())
}
