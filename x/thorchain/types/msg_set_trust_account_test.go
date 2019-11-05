package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type MsgSetTrustAccountSuite struct{}

var _ = Suite(&MsgSetTrustAccountSuite{})

func (MsgSetTrustAccountSuite) TestMsgSetTrustAccount(c *C) {
	bnb := GetRandomBNBAddress()
	acc1 := GetRandomBech32Addr()
	c.Assert(acc1.Empty(), Equals, false)
	consensPubKey := GetRandomBech32ConsensusPubKey()
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
