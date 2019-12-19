package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type MsgSetNodeKeysSuite struct{}

var _ = Suite(&MsgSetNodeKeysSuite{})

func (MsgSetNodeKeysSuite) TestMsgSetNodeKeys(c *C) {
	acc1 := GetRandomBech32Addr()
	c.Assert(acc1.Empty(), Equals, false)
	consensPubKey := GetRandomBech32ConsensusPubKey()
	pubKeys := common.PubKeys{
		Secp256k1: GetRandomPubKey(),
		Ed25519:   GetRandomPubKey(),
	}
	msgSetNodeKeys := NewMsgSetNodeKeys(pubKeys, consensPubKey, acc1)
	c.Assert(msgSetNodeKeys.Route(), Equals, RouterKey)
	c.Assert(msgSetNodeKeys.Type(), Equals, "set_node_keys")
	c.Assert(msgSetNodeKeys.ValidateBasic(), IsNil)
	c.Assert(len(msgSetNodeKeys.GetSignBytes()) > 0, Equals, true)
	c.Assert(msgSetNodeKeys.GetSigners(), NotNil)
	c.Assert(msgSetNodeKeys.GetSigners()[0].String(), Equals, acc1.String())
	msgUpdateNodeAccount1 := NewMsgSetNodeKeys(pubKeys, "", acc1)
	c.Assert(msgUpdateNodeAccount1.ValidateBasic(), NotNil)

	msgUpdateNodeAccount2 := NewMsgSetNodeKeys(pubKeys, consensPubKey, sdk.AccAddress{})
	c.Assert(msgUpdateNodeAccount2.ValidateBasic(), NotNil)

	emptyPubKeys := NewMsgSetNodeKeys(common.PubKeys{}, consensPubKey, acc1)
	c.Assert(emptyPubKeys.ValidateBasic(), NotNil)

}
