package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
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
	msgApply := NewMsgBond(nodeAddr, sdk.NewUint(common.One), txId, bondAddr, signerAddr)
	c.Assert(msgApply.ValidateBasic(), IsNil)
	c.Assert(msgApply.Route(), Equals, RouterKey)
	c.Assert(msgApply.Type(), Equals, "validator_apply")
	c.Assert(msgApply.GetSignBytes(), NotNil)
	c.Assert(len(msgApply.GetSigners()), Equals, 1)
	c.Assert(msgApply.GetSigners()[0].Equals(signerAddr), Equals, true)
	c.Assert(NewMsgBond(sdk.AccAddress{}, sdk.NewUint(common.One), txId, bondAddr, signerAddr).ValidateBasic(), NotNil)
	c.Assert(NewMsgBond(nodeAddr, sdk.ZeroUint(), txId, bondAddr, signerAddr).ValidateBasic(), NotNil)
	c.Assert(NewMsgBond(nodeAddr, sdk.NewUint(common.One), "", bondAddr, signerAddr).ValidateBasic(), NotNil)
	c.Assert(NewMsgBond(nodeAddr, sdk.NewUint(common.One), txId, "", signerAddr).ValidateBasic(), NotNil)
	c.Assert(NewMsgBond(nodeAddr, sdk.NewUint(common.One), txId, bondAddr, sdk.AccAddress{}).ValidateBasic(), NotNil)
}
