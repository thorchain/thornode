package types

import (
	. "gopkg.in/check.v1"
)

type ModeSuite struct{}

var _ = Suite(&ModeSuite{})

func (s ModeSuite) TestMode(c *C) {
	mode, err := NewMode("sync")
	c.Assert(err, IsNil)
	c.Check(mode, Equals, TxSync)
	c.Check(mode.IsValid(), Equals, true)
	c.Check(mode.String(), Equals, "sync")

	mode, err = NewMode("bogus")
	c.Assert(err, NotNil)
	c.Check(mode, Equals, TxUnknown)
	c.Check(mode.IsValid(), Equals, false)
	c.Check(mode.String(), Equals, "unknown")
}
