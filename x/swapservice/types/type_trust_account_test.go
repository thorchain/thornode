package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type TrustAccountSuite struct{}

var _ = Suite(&TrustAccountSuite{})

func (TrustAccountSuite) TestTrustAccount(c *C) {
	bnb := GetRandomBNBAddress()
	addr := GetRandomBech32Addr()
	consensusAddr := "bepcpub1zcjduepqrkasznnv37qcguhn6z33v2ndldpq00f7yldamjrtc2a0sc4vqrqqvr9t8t"
	pk, err := sdk.GetConsPubKeyBech32(consensusAddr)
	c.Assert(err, IsNil)
	c.Assert(pk, NotNil)
	c.Check(addr.Empty(), Equals, false)
	bepConsPubKey := `bepcpub1zcjduepq4kn64fcjhf0fp20gp8var0rm25ca9jy6jz7acem8gckh0nkplznq85gdrg`
	trustAccount := NewTrustAccount(bnb, addr, bepConsPubKey)
	err = trustAccount.IsValid()
	c.Assert(err, IsNil)
	c.Assert(trustAccount.ObserverBEPAddress.Equals(addr), Equals, true)
	c.Assert(trustAccount.SignerBNBAddress, Equals, bnb)
	c.Assert(trustAccount.ValidatorBEPConsPubKey, Equals, bepConsPubKey)
	c.Log(trustAccount.String())

	trustAccount1 := NewTrustAccount(common.NoBnbAddress, addr, bepConsPubKey)
	c.Assert(trustAccount1.IsValid(), IsNil)
	c.Assert(NewTrustAccount(bnb, sdk.AccAddress{}, bepConsPubKey).IsValid(), NotNil)
	c.Assert(NewTrustAccount(bnb, addr, "").IsValid(), NotNil)
}
