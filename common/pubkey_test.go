package common

import (
	. "gopkg.in/check.v1"
)

type PubKeyTestSuite struct{}

var _ = Suite(&PubKeyTestSuite{})

// TestPubKey implementation
func (PubKeyTestSuite) TestPubKey(c *C) {
	input := RandStringBytesMask(20)
	pk := NewPubKey([]byte(input))
	hexStr := pk.String()
	c.Assert(len(hexStr) > 0, Equals, true)
	pk1, err := NewPubKeyFromHexString(hexStr)
	c.Assert(err, IsNil)
	c.Assert(pk.Equals(pk1), Equals, true)
}
