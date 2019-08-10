package types

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestPackage(t *testing.T) { TestingT(t) }

type TypesSuite struct{}

var _ = Suite(&TypesSuite{})

func (s TypesSuite) TestTicker(c *C) {
	runeTicker, err := NewTicker("rune")
	c.Assert(err, IsNil)
	bnbTicker, err := NewTicker("bnb")
	c.Assert(err, IsNil)
	c.Check(runeTicker.Equals(RuneTicker), Equals, true)
	c.Check(bnbTicker.Equals(RuneTicker), Equals, false)
	c.Check(IsRune(runeTicker), Equals, true)

	c.Check(runeTicker.String(), Equals, "RUNE")

	_, err = NewTicker("t") // too short
	c.Assert(err, NotNil)

	_, err = NewTicker("too long of a token") // too long
	c.Assert(err, NotNil)
}

func (s *TypesSuite) TestBnbAddress(c *C) {
	addr, err := NewBnbAddress("bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(err, IsNil)
	c.Check(addr.Equals(BnbAddress("bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")), Equals, true)
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

	c.Check(NoBnbAddress.Equals(BnbAddress("")), Equals, true)
}
