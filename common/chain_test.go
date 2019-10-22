package common

import (
	. "gopkg.in/check.v1"
)

type ChainSuite struct{}

var _ = Suite(&ChainSuite{})

func (s ChainSuite) TestChain(c *C) {
	bnbChain, err := NewChain("bnb")
	c.Assert(err, IsNil)
	c.Check(bnbChain.Equals(BNBChain), Equals, true)
	c.Check(IsBNBChain(bnbChain), Equals, true)
	c.Check(bnbChain.IsEmpty(), Equals, false)
	c.Check(bnbChain.String(), Equals, "BNB")

	_, err = NewChain("B") // too short
	c.Assert(err, NotNil)
	_, err = NewChain("BOGUS") // too long
	c.Assert(err, NotNil)
}
