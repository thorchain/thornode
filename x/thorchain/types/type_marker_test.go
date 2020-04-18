package types

import (
	. "gopkg.in/check.v1"
)

type TxMarkerSuite struct{}

var _ = Suite(&TxMarkerSuite{})

func (s *TxMarkerSuite) TestTxMarker(c *C) {
	mark := TxMarker{0, "bar"}
	c.Check(mark.IsEmpty(), Equals, true)
	mark = TxMarker{5, ""}
	c.Check(mark.IsEmpty(), Equals, true)
	mark = NewTxMarker(12, "bar")
	c.Check(mark.IsEmpty(), Equals, false)
}

func (s *TxMarkerSuite) TestTxMarkers(c *C) {
	marks := TxMarkers{
		NewTxMarker(10, "foo"),
		NewTxMarker(100, "bar"),
		NewTxMarker(1000, "baz"),
	}

	var mark TxMarker
	mark, marks = marks.Pop()
	c.Check(mark.Height, Equals, int64(10))
	c.Assert(marks, HasLen, 2)

	marks = marks.FilterByMinHeight(101)
	c.Assert(marks, HasLen, 1)
	c.Assert(marks[0].Height, Equals, int64(1000))
}
