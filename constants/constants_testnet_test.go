//go:build testnet
// +build testnet

package constants

import (
	. "gopkg.in/check.v1"
)

type ConstantsTestnetSuite struct{}

var _ = Suite(&ConstantsTestnetSuite{})

func (s *ConstantsTestnetSuite) TestTestNet(c *C) {
	consts := NewConstantValue010()
	c.Check(consts.GetInt64Value(MinimumBondInRune), Equals, int64(100000000))
}
