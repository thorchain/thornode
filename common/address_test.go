package common

import (
	. "gopkg.in/check.v1"

	atypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/tendermint/tendermint/crypto"
)

type AddressSuite struct{}

var _ = Suite(&AddressSuite{})

func (s *AddressSuite) TestAddress(c *C) {
	addr, err := NewAddress("bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Assert(err, IsNil)
	c.Check(addr.IsEmpty(), Equals, false)
	c.Check(addr.Equals(Address("bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")), Equals, true)
	c.Check(addr.String(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(addr.IsChain(BNBChain), Equals, true)
	_, err = NewAddress("bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(err, IsNil)
	_, err = NewAddress("tbnb12ymaslcrhnkj0tvmecyuejdvk25k2nnurqjvyp")
	c.Check(err, IsNil)
	_, err = NewAddress("1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(err, NotNil)
	_, err = NewAddress("bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6X")
	c.Check(err, NotNil)
	_, err = NewAddress("bogus")
	c.Check(err, NotNil)
	c.Check(Address("").IsEmpty(), Equals, true)

	c.Check(NoAddress.Equals(Address("")), Equals, true)
	_, err = NewAddress("")
	c.Assert(err, IsNil)

	_, pubKey, _ := atypes.KeyTestPubAddr()
	inputBytes := crypto.AddressHash(pubKey.Bytes())
	pk := NewPubKey(inputBytes)
	addr, err = pk.GetAddress(BNBChain)
	c.Assert(err, IsNil)
	pk2, err := addr.PubKey()
	c.Assert(err, IsNil)
	c.Check(pk.String(), Equals, pk2.String())
}
