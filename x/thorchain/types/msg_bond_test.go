package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type MsgApplySuite struct{}

var _ = Suite(&MsgApplySuite{})

func (mas *MsgApplySuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

func (MsgApplySuite) TestMsgApply(c *C) {
	nodeAddr := GetRandomBech32Addr()
	txId := GetRandomTxHash()
	c.Check(txId.IsEmpty(), Equals, false)
	signerAddr := GetRandomBech32Addr()
	bondAddr := GetRandomBNBAddress()
	txin := GetRandomTx()
	txinNoID := txin
	txinNoID.ID = ""
	msgApply := NewMsgBond(txin, nodeAddr, sdk.NewUint(common.One), bondAddr, signerAddr)
	c.Assert(msgApply.ValidateBasic(), IsNil)
	c.Assert(msgApply.Route(), Equals, RouterKey)
	c.Assert(msgApply.Type(), Equals, "validator_apply")
	c.Assert(msgApply.GetSignBytes(), NotNil)
	c.Assert(len(msgApply.GetSigners()), Equals, 1)
	c.Assert(msgApply.GetSigners()[0].Equals(signerAddr), Equals, true)
	c.Assert(NewMsgBond(txin, sdk.AccAddress{}, sdk.NewUint(common.One), bondAddr, signerAddr).ValidateBasic(), NotNil)
	c.Assert(NewMsgBond(txin, nodeAddr, sdk.ZeroUint(), bondAddr, signerAddr).ValidateBasic(), NotNil)
	c.Assert(NewMsgBond(txinNoID, nodeAddr, sdk.NewUint(common.One), bondAddr, signerAddr).ValidateBasic(), NotNil)
	c.Assert(NewMsgBond(txin, nodeAddr, sdk.NewUint(common.One), "", signerAddr).ValidateBasic(), NotNil)
	c.Assert(NewMsgBond(txin, nodeAddr, sdk.NewUint(common.One), bondAddr, sdk.AccAddress{}).ValidateBasic(), NotNil)
}
