package types

import (
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/common"
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

	msgAck := NewMsgAck(txID, sender, common.BNBChain, signer)
	c.Assert(msgAck.Type(), Equals, "set_ack")
	EnsureMsgBasicCorrect(msgAck, c)

	emptySender := NewMsgAck(txID, "", common.BNBChain, signer)
	c.Assert(emptySender.ValidateBasic(), NotNil)
	emptyHash := NewMsgAck("", sender, common.BNBChain, signer)
	c.Assert(emptyHash.ValidateBasic(), NotNil)
	emptyChain := NewMsgAck(txID, sender, common.EmptyChain, signer)
	c.Assert(emptyChain.ValidateBasic(), NotNil)
}
