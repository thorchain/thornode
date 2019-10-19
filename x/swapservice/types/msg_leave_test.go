package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type MsgLeaveSuite struct{}

var _ = Suite(&MsgLeaveSuite{})

func (*MsgLeaveSuite) SetupSuite(c *C) {
	SetupConfigForTest()
}
func (MsgLeaveSuite) TestMsgLeave(c *C) {
	nodeAddr := GetRandomBech32Addr()
	txId := GetRandomTxHash()
	senderBNBAddr := GetRandomBNBAddress()
	msgLeave := NewMsgLeave(txId, senderBNBAddr, nodeAddr)
	EnsureMsgBasicCorrect(msgLeave, c)
	c.Assert(msgLeave.ValidateBasic(), IsNil)
	c.Assert(msgLeave.Type(), Equals, "validator_leave")

	msgLeave1 := NewMsgLeave(txId, senderBNBAddr, nodeAddr)
	c.Assert(msgLeave1.ValidateBasic(), IsNil)
	msgLeave2 := NewMsgLeave("", senderBNBAddr, nodeAddr)
	c.Assert(msgLeave2.ValidateBasic(), NotNil)
	msgLeave3 := NewMsgLeave(txId, senderBNBAddr, sdk.AccAddress{})
	c.Assert(msgLeave3.ValidateBasic(), NotNil)
	msgLeave4 := NewMsgLeave(txId, "", nodeAddr)
	c.Assert(msgLeave4.ValidateBasic(), NotNil)
}
