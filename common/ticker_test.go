package common

import (
	. "gopkg.in/check.v1"
)

type TickerSuite struct{}

var _ = Suite(&TickerSuite{})

func (s TickerSuite) TestTicker(c *C) {
	runeTicker, err := NewTicker("rune")
	c.Assert(err, IsNil)
	bnbTicker, err := NewTicker("bnb")
	c.Assert(err, IsNil)
	c.Check(runeTicker.IsEmpty(), Equals, false)
	c.Check(runeTicker.Equals(RuneTicker), Equals, true)
	c.Check(bnbTicker.Equals(RuneTicker), Equals, false)
	c.Check(IsRune(runeTicker), Equals, true)
	c.Check(IsBNB(runeTicker), Equals, false)
	c.Check(IsBNB(bnbTicker), Equals, true)
	c.Check(runeTicker.String(), Equals, "RUNE")
	runeB1aTicker, err := NewTicker("rune-b1a")
	c.Assert(err, IsNil)
	c.Check(runeB1aTicker.Equals(RuneB1ATicker), Equals, true)
	c.Check(IsRune(runeB1aTicker), Equals, true)
	c.Check(runeB1aTicker.String(), Equals, RuneB1ATicker.String())

	runeA1fTicker, err := NewTicker("rune-a1f")
	c.Assert(err, IsNil)
	c.Check(runeA1fTicker.Equals(RuneA1FTicker), Equals, true)
	c.Check(IsRune(runeA1fTicker), Equals, true)
	c.Check(runeA1fTicker.String(), Equals, RuneA1FTicker.String())

	tomobTicker, err := NewTicker("TOMOB-1E1")
	c.Assert(err, IsNil)
	c.Assert(tomobTicker.String(), Equals, "TOMOB-1E1")
	_, err = NewTicker("t") // too short
	c.Assert(err, NotNil)

	maxCharacterTicker, err := NewTicker("TICKER789-XXX")
	c.Assert(err, IsNil)
	c.Assert(maxCharacterTicker.IsEmpty(), Equals, false)
	_, err = NewTicker("too long of a token") // too long
	c.Assert(err, NotNil)
}
