package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/statechain/cmd"
)

type MsgApplySuite struct{}

var _ = Suite(&MsgApplySuite{})

func (mas *MsgApplySuite) SetUpSuite(c *C) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(cmd.Bech32PrefixValAddr, cmd.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(cmd.Bech32PrefixConsAddr, cmd.Bech32PrefixConsPub)
}

func (MsgApplySuite) TestMsgApply(c *C) {
	nodeAddr, err := sdk.AccAddressFromBech32("bep180xs5jx2szhww4jq4xfmvpza7kzr6rwu9408dm")
	c.Assert(err, IsNil)
	txId, err := common.NewTxID("712882AC9587198FA46F8D79BDFF013E77A89B12882702F03FA60FD298C517A4")
	c.Assert(err, IsNil)
	c.Check(txId.IsEmpty(), Equals, false)
	signerAddr, err := sdk.AccAddressFromBech32("bep1n93wghyzlksfxyxvrejfc9eh3dfqkdzfs7k8fg")
	c.Assert(err, IsNil)
	msgApply := NewMsgApply(nodeAddr, sdk.NewUint(common.One), txId, signerAddr)
	c.Assert(msgApply.ValidateBasic(), IsNil)
	c.Assert(msgApply.Route(), Equals, RouterKey)
	c.Assert(msgApply.Type(), Equals, "validator_apply")
	c.Assert(msgApply.GetSignBytes(), NotNil)
	c.Assert(len(msgApply.GetSigners()), Equals, 1)
	c.Assert(msgApply.GetSigners()[0].Equals(signerAddr), Equals, true)
	c.Assert(NewMsgApply(sdk.AccAddress{}, sdk.NewUint(common.One), txId, signerAddr).ValidateBasic(), NotNil)
	c.Assert(NewMsgApply(nodeAddr, sdk.ZeroUint(), txId, signerAddr).ValidateBasic(), NotNil)
	c.Assert(NewMsgApply(nodeAddr, sdk.NewUint(common.One), "", signerAddr).ValidateBasic(), NotNil)
	c.Assert(NewMsgApply(nodeAddr, sdk.NewUint(common.One), txId, sdk.AccAddress{}).ValidateBasic(), NotNil)
}
