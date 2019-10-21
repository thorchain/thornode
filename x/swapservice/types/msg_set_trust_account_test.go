package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type MsgSetTrustAccountSuite struct{}

var _ = Suite(&MsgSetTrustAccountSuite{})

func (MsgSetTrustAccountSuite) TestMsgSetTrustAccount(c *C) {
	bnb, err := common.NewAddress("bnb1xlvns0n2mxh77mzaspn2hgav4rr4m8eerfju38")
	c.Assert(err, IsNil)
	acc1, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	c.Assert(acc1.Empty(), Equals, false)
	consensPubKey := `bepcpub1zcjduepqtqr6c2nks2220zt2n2p7v2ph4fzpn85aydgnn0mvx3n6gaqkpg0qqkfr5a`
	trustAccount := NewTrustAccount(bnb, acc1, consensPubKey)
	c.Assert(trustAccount.IsValid(), IsNil)
	msgSetTrustAccount := NewMsgSetTrustAccount(trustAccount, acc1)
	c.Assert(msgSetTrustAccount.Route(), Equals, RouterKey)
	c.Assert(msgSetTrustAccount.Type(), Equals, "set_trust_account")
	c.Assert(msgSetTrustAccount.ValidateBasic(), IsNil)
	c.Assert(len(msgSetTrustAccount.GetSignBytes()) > 0, Equals, true)
	c.Assert(msgSetTrustAccount.GetSigners(), NotNil)
	c.Assert(msgSetTrustAccount.GetSigners()[0].String(), Equals, acc1.String())
	msgUpdateNodeAccount1 := NewMsgSetTrustAccount(NewTrustAccount(bnb, acc1, ""), acc1)
	c.Assert(msgUpdateNodeAccount1.ValidateBasic(), NotNil)
	msgUpdateNodeAccount2 := NewMsgSetTrustAccount(trustAccount, sdk.AccAddress{})
	c.Assert(msgUpdateNodeAccount2.ValidateBasic(), NotNil)

}
