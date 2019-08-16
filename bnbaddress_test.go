package common

import (
	. "gopkg.in/check.v1"
)

type BnBAddressSuite struct{}

var _ = Suite(&BnBAddressSuite{})

func (s *BnBAddressSuite) TestBnbAddress(c *C) {
	addr, err := NewBnbAddress("bnbblejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(err, IsNil)
	c.Check(addr.IsEmpty(), Equals, false)
	c.Check(addr.Equals(BnbAddress("bnbblejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")), Equals, true)
	c.Check(addr.String(), Equals, "bnbblejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	_, err = NewBnbAddress("bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(err, IsNil)
	_, err = NewBnbAddress("tbnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(err, IsNil)
	_, err = NewBnbAddress("1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(err, NotNil)
	_, err = NewBnbAddress("bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6X")
	c.Check(err, NotNil)
	_, err = NewBnbAddress("bogus")
	c.Check(err, NotNil)
	c.Check(BnbAddress("").IsEmpty(), Equals, true)

	c.Check(NoBnbAddress.Equals(BnbAddress("")), Equals, true)
	_, err = NewBnbAddress("")
	c.Assert(err, IsNil)
}
