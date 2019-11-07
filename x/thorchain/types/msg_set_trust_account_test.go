package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

type MsgSetTrustAccountSuite struct{}

var _ = Suite(&MsgSetTrustAccountSuite{})

func (MsgSetTrustAccountSuite) TestMsgSetTrustAccount(c *C) {
	acc1 := GetRandomBech32Addr()
	c.Assert(acc1.Empty(), Equals, false)
	consensPubKey := GetRandomBech32ConsensusPubKey()
	pubKeys := common.PubKeys{
		Secp256k1: GetRandomPubKey(),
		Ed25519:   GetRandomPubKey(),
	}
	msgSetTrustAccount := NewMsgSetTrustAccount(pubKeys, consensPubKey, acc1)
	c.Assert(msgSetTrustAccount.Route(), Equals, RouterKey)
	c.Assert(msgSetTrustAccount.Type(), Equals, "set_trust_account")
	c.Assert(msgSetTrustAccount.ValidateBasic(), IsNil)
	c.Assert(len(msgSetTrustAccount.GetSignBytes()) > 0, Equals, true)
	c.Assert(msgSetTrustAccount.GetSigners(), NotNil)
	c.Assert(msgSetTrustAccount.GetSigners()[0].String(), Equals, acc1.String())
	msgUpdateNodeAccount1 := NewMsgSetTrustAccount(pubKeys, "", acc1)
	c.Assert(msgUpdateNodeAccount1.ValidateBasic(), NotNil)

	msgUpdateNodeAccount2 := NewMsgSetTrustAccount(pubKeys, consensPubKey, sdk.AccAddress{})
	c.Assert(msgUpdateNodeAccount2.ValidateBasic(), NotNil)

	emptyPubKeys := NewMsgSetTrustAccount(common.PubKeys{}, consensPubKey, acc1)
	c.Assert(emptyPubKeys.ValidateBasic(), NotNil)

}
