package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type MsgLeaveSuite struct{}

var _ = Suite(&MsgLeaveSuite{})

func (*MsgLeaveSuite) SetupSuite(c *C) {
	SetupConfigForTest()
}
func (MsgLeaveSuite) TestMsgLeave(c *C) {
	nodeAddr, err := sdk.AccAddressFromBech32("bep180xs5jx2szhww4jq4xfmvpza7kzr6rwu9408dm")
	c.Assert(err, IsNil)
	txId, err := common.NewTxID("712882AC9587198FA46F8D79BDFF013E77A89B12882702F03FA60FD298C517A4")
	c.Assert(err, IsNil)
	bnbAddress, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)
	senderBNBAddr, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqb")
	msgLeave := NewMsgLeave(bnbAddress, txId, senderBNBAddr, nodeAddr)
	EnsureMsgBasicCorrect(msgLeave, c)
	c.Assert(msgLeave.ValidateBasic(), IsNil)
	c.Assert(msgLeave.Type(), Equals, "validator_leave")

	msgLeave1 := NewMsgLeave("", txId, senderBNBAddr, nodeAddr)
	c.Assert(msgLeave1.ValidateBasic(), NotNil)
	msgLeave2 := NewMsgLeave(bnbAddress, "", senderBNBAddr, nodeAddr)
	c.Assert(msgLeave2.ValidateBasic(), NotNil)
	msgLeave3 := NewMsgLeave(bnbAddress, txId, senderBNBAddr, sdk.AccAddress{})
	c.Assert(msgLeave3.ValidateBasic(), NotNil)
	msgLeave4 := NewMsgLeave(bnbAddress, txId, "", nodeAddr)
	c.Assert(msgLeave4.ValidateBasic(), NotNil)
}
