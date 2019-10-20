package types

import (
	. "gopkg.in/check.v1"
)

type MsgAckSuite struct{}

var _ = Suite(&MsgAckSuite{})

func (mas *MsgAckSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

func (mas *MsgAckSuite) TestMsgAck(c *C) {
	txID := GetRandomTxHash()
	sender := GetRandomBNBAddress()
	signer := GetRandomBech32Addr()
	msgAck := NewMsgAck(txID, sender, signer)
	c.Assert(msgAck.Type(), Equals, "set_ack")
	EnsureMsgBasicCorrect(msgAck, c)

	emptySender := NewMsgAck(txID, "", signer)
	c.Assert(emptySender.ValidateBasic(), NotNil)
	emptyHash := NewMsgAck("", sender, signer)
	c.Assert(emptyHash.ValidateBasic(), NotNil)
}
