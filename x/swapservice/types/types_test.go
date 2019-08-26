package types

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestPackage(t *testing.T) { TestingT(t) }

type TypesSuite struct{}

var _ = Suite(&TypesSuite{})

func (s TypesSuite) TestHasMajority(c *C) {
	// happy path
	c.Check(HasMajority(3, 4), Equals, true)
	c.Check(HasMajority(4, 4), Equals, true)

	// unhappy path
	c.Check(HasMajority(2, 4), Equals, false)
	c.Check(HasMajority(9, 4), Equals, false)
	c.Check(HasMajority(-9, 4), Equals, false)
	c.Check(HasMajority(9, -4), Equals, false)
	c.Check(HasMajority(0, 0), Equals, false)
	c.Check(HasMajority(3, 0), Equals, false)
}
