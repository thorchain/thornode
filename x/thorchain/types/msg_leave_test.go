package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
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
	tx := common.Tx{ID: txId, FromAddress: senderBNBAddr}
	msgLeave := NewMsgLeave(tx, nodeAddr)
	EnsureMsgBasicCorrect(msgLeave, c)
	c.Assert(msgLeave.ValidateBasic(), IsNil)
	c.Assert(msgLeave.Type(), Equals, "validator_leave")

	msgLeave1 := NewMsgLeave(tx, nodeAddr)
	c.Assert(msgLeave1.ValidateBasic(), IsNil)
	msgLeave2 := NewMsgLeave(common.Tx{ID: "", FromAddress: senderBNBAddr}, nodeAddr)
	c.Assert(msgLeave2.ValidateBasic(), NotNil)
	msgLeave3 := NewMsgLeave(tx, sdk.AccAddress{})
	c.Assert(msgLeave3.ValidateBasic(), NotNil)
	msgLeave4 := NewMsgLeave(common.Tx{ID: txId, FromAddress: ""}, nodeAddr)
	c.Assert(msgLeave4.ValidateBasic(), NotNil)
}
