package types

import (
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type MsgAckSuite struct{}

var _ = Suite(&MsgAckSuite{})

func (mas *MsgAckSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

func (mas *MsgAckSuite) TestMsgAck(c *C) {
	tx := GetRandomTx()
	sender := GetRandomBNBAddress()
	signer := GetRandomBech32Addr()

	msgAck := NewMsgAck(tx, sender, common.BNBChain, signer)
	c.Assert(msgAck.Type(), Equals, "set_ack")
	EnsureMsgBasicCorrect(msgAck, c)

	emptySender := NewMsgAck(tx, "", common.BNBChain, signer)
	c.Assert(emptySender.ValidateBasic(), NotNil)
	emptyChain := NewMsgAck(tx, sender, common.EmptyChain, signer)
	c.Assert(emptyChain.ValidateBasic(), NotNil)
	tx.ID = ""
	emptyHash := NewMsgAck(tx, sender, common.BNBChain, signer)
	c.Assert(emptyHash.ValidateBasic(), NotNil)
}
