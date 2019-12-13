package common

import (
	. "gopkg.in/check.v1"
)

type AddressSuite struct{}

var _ = Suite(&AddressSuite{})

func (s *AddressSuite) TestAddress(c *C) {
	// BTC address
	addr, err := NewAddress("bc1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3qccfmv3")
	c.Assert(err, IsNil)
	c.Check(addr.String(), Equals, "bc1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3qccfmv3")
	c.Check(addr.IsEmpty(), Equals, false)
	c.Check(addr.IsChain(BTCChain), Equals, true)
	_, err = NewAddress("tb1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3q0sl5k7")
	c.Check(err, IsNil)

	// BNB
	addr, err = NewAddress("bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
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

	// failures
	c.Check(err, NotNil)
	_, err = NewAddress("bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6X")
	c.Check(err, NotNil)
	_, err = NewAddress("bogus")
	c.Check(err, NotNil)
	c.Check(Address("").IsEmpty(), Equals, true)

	c.Check(NoAddress.Equals(Address("")), Equals, true)
	_, err = NewAddress("")
	c.Assert(err, IsNil)
}
