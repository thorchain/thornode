package common

import (
	. "gopkg.in/check.v1"
)

type AmountSuite struct{}

var _ = Suite(&AmountSuite{})

func (s *AmountSuite) TestAmount(c *C) {
	amt, err := NewAmount("3.45")
	c.Check(err, IsNil)
	c.Check(amt.Equals("3.45"), Equals, true)
	c.Check(amt.String(), Equals, "3.45")
	c.Check(amt.GreaterThen(3), Equals, true)
	c.Check(amt.GreaterThen(4), Equals, false)
	c.Check(amt.LessThen(4), Equals, true)
	c.Check(amt.LessThen(3), Equals, false)
	c.Check(amt.Minus("2").Equals("1.45"), Equals, true)
	c.Check(amt.Plus("2").Equals("5.45"), Equals, true)
	c.Check(Amount("").IsEmpty(), Equals, true)
	c.Check(Amount("100").IsEmpty(), Equals, false)
	c.Check(Amount("-1").IsNegative(), Equals, true)
	c.Check(Amount("1").IsNegative(), Equals, false)
	c.Check(ZeroAmount.Equals(Amount("0")), Equals, true)
	c.Check(NewAmountFromFloat(0).IsZero(), Equals, true)
	c.Check(NewAmountFromFloat(1.45).IsZero(), Equals, false)
	amt, err = NewAmount("bogus")
	c.Assert(err, NotNil)
	c.Check(amt.Equals(ZeroAmount), Equals, true)
}
