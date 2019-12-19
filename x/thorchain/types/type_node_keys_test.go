package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type NodeKeysSuite struct{}

var _ = Suite(&NodeKeysSuite{})

func (NodeKeysSuite) TestNodeKeys(c *C) {
	bnb := GetRandomBNBAddress()
	addr := GetRandomBech32Addr()
	consensusAddr := GetRandomBech32ConsensusPubKey()
	pk, err := sdk.GetConsPubKeyBech32(consensusAddr)
	c.Assert(err, IsNil)
	c.Assert(pk, NotNil)
	c.Check(addr.Empty(), Equals, false)
	bepConsPubKey := GetRandomBech32ConsensusPubKey()
	nodeKeys := NewNodeKeys(bnb, addr, bepConsPubKey)
	err = nodeKeys.IsValid()
	c.Assert(err, IsNil)
	c.Assert(nodeKeys.ObserverBEPAddress.Equals(addr), Equals, true)
	c.Assert(nodeKeys.SignerBNBAddress, Equals, bnb)
	c.Assert(nodeKeys.ValidatorBEPConsPubKey, Equals, bepConsPubKey)
	c.Log(nodeKeys.String())

	nodeKeys1 := NewNodeKeys(common.NoAddress, addr, bepConsPubKey)
	c.Assert(nodeKeys1.IsValid(), IsNil)
	c.Assert(NewNodeKeys(bnb, sdk.AccAddress{}, bepConsPubKey).IsValid(), NotNil)
	c.Assert(NewNodeKeys(bnb, addr, "").IsValid(), NotNil)
}
