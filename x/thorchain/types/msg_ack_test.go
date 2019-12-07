package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type MsgAckSuite struct{}

var _ = Suite(&MsgAckSuite{})

func (mas *MsgAckSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

func (mas *MsgAckSuite) TestMsgAck(c *C) {
	tx := NewObservedTx(GetRandomTx(), sdk.NewUint(12), GetRandomPubKey())
	sender := GetRandomBNBAddress()
	signer := GetRandomBech32Addr()

	msgAck := NewMsgAck(tx, sender, common.BNBChain, signer)
	c.Assert(msgAck.Type(), Equals, "set_ack")
	EnsureMsgBasicCorrect(msgAck, c)

	emptySender := NewMsgAck(tx, "", common.BNBChain, signer)
	c.Assert(emptySender.ValidateBasic(), NotNil)
	emptyChain := NewMsgAck(tx, sender, common.EmptyChain, signer)
	c.Assert(emptyChain.ValidateBasic(), NotNil)
	tx.Tx.ID = ""
	emptyHash := NewMsgAck(tx, sender, common.BNBChain, signer)
	c.Assert(emptyHash.ValidateBasic(), NotNil)
}
