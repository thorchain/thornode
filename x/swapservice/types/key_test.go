package types

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestPackage(t *testing.T) { TestingT(t) }

type KeySuite struct{}

var _ = Suite(&KeySuite{})

func (x KeySuite) TestIsRune(c *C) {
	c.Check(IsRune("RUNE"), Equals, true)
	c.Check(IsRune("rune"), Equals, true)
	c.Check(IsRune("RUNE2"), Equals, false)
	c.Check(IsRune(""), Equals, false)
}
