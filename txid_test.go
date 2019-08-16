package types

import (
	. "gopkg.in/check.v1"
)

type TxIDSuite struct{}

var _ = Suite(&TxIDSuite{})

func (s TxIDSuite) TestTxID(c *C) {
	ID := "A7DA8FF1B7C290616D68A276F30AC618315E6CCE982EB8F7A79339E163798F49"
	tx, err := NewTxID(ID)
	c.Assert(err, IsNil)
	c.Check(tx.String(), Equals, ID)
	c.Check(tx.IsEmpty(), Equals, false)
	c.Check(tx.Equals(TxID(ID)), Equals, true)

	_, err = NewTxID("bogus")
	c.Check(err, NotNil)
}
