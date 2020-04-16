package thorchain

import (
	. "gopkg.in/check.v1"
)

type KeeperTxMarkerSuite struct{}

var _ = Suite(&KeeperTxMarkerSuite{})

func (s *KeeperTxMarkerSuite) TestTxMarker(c *C) {
	ctx, k := setupKeeperForTest(c)

	mark1 := NewTxMarker(25, "my memo")
	mark2 := NewTxMarker(30, "my other memo")
	c.Assert(k.AppendTxMarker(ctx, "hash", mark1), IsNil)
	c.Assert(k.AppendTxMarker(ctx, "hash", mark2), IsNil)

	marks, err := k.ListTxMarker(ctx, "hash")
	c.Assert(err, IsNil)
	c.Assert(marks, HasLen, 2)
	c.Assert(marks[0].Height, Equals, mark1.Height)
	c.Assert(marks[0].Memo, Equals, mark1.Memo)
	c.Assert(marks[1].Height, Equals, mark2.Height)
	c.Assert(marks[1].Memo, Equals, mark2.Memo)

	c.Assert(k.SetTxMarkers(ctx, "hash", TxMarkers{mark2}), IsNil)
	marks, err = k.ListTxMarker(ctx, "hash")
	c.Assert(err, IsNil)
	c.Assert(marks, HasLen, 1)
	c.Assert(marks[0].Height, Equals, mark2.Height)
	c.Assert(marks[0].Memo, Equals, mark2.Memo)

	c.Assert(k.SetTxMarkers(ctx, "hash", nil), IsNil)
	marks, err = k.ListTxMarker(ctx, "hash")
	c.Assert(err, IsNil)
	c.Assert(marks, HasLen, 0)
}
