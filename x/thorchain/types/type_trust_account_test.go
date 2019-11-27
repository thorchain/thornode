package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type TrustAccountSuite struct{}

var _ = Suite(&TrustAccountSuite{})

func (TrustAccountSuite) TestTrustAccount(c *C) {
	bnb := GetRandomBNBAddress()
	addr := GetRandomBech32Addr()
	consensusAddr := GetRandomBech32ConsensusPubKey()
	pk, err := sdk.GetConsPubKeyBech32(consensusAddr)
	c.Assert(err, IsNil)
	c.Assert(pk, NotNil)
	c.Check(addr.Empty(), Equals, false)
	bepConsPubKey := GetRandomBech32ConsensusPubKey()
	trustAccount := NewTrustAccount(bnb, addr, bepConsPubKey)
	err = trustAccount.IsValid()
	c.Assert(err, IsNil)
	c.Assert(trustAccount.ObserverBEPAddress.Equals(addr), Equals, true)
	c.Assert(trustAccount.SignerBNBAddress, Equals, bnb)
	c.Assert(trustAccount.ValidatorBEPConsPubKey, Equals, bepConsPubKey)
	c.Log(trustAccount.String())

	trustAccount1 := NewTrustAccount(common.NoAddress, addr, bepConsPubKey)
	c.Assert(trustAccount1.IsValid(), IsNil)
	c.Assert(NewTrustAccount(bnb, sdk.AccAddress{}, bepConsPubKey).IsValid(), NotNil)
	c.Assert(NewTrustAccount(bnb, addr, "").IsValid(), NotNil)
}
