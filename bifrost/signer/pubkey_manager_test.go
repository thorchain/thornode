package signer

import (
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

type PukKeyManagerSuite struct{}

var _ = Suite(&PukKeyManagerSuite{})

func (s *PukKeyManagerSuite) TestPubKeyManager(c *C) {
	pkm := NewPubKeyManager()

	pk, err := common.NewPubKeyFromBech32("tbnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa", "tbnb")
	c.Assert(err, IsNil)
	c.Assert(pkm.pks, HasLen, 0)

	pkm.Add(pk)
	c.Assert(pkm.pks, HasLen, 1)
	c.Assert(pkm.pks[0].Equals(pk), Equals, true)

	pkm.Remove(pk)
	c.Assert(pkm.pks, HasLen, 0)
}
