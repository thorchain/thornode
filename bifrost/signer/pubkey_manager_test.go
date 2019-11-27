package signer

import (
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

type PukKeyManagerSuite struct{}

var _ = Suite(&PukKeyManagerSuite{})

func (s *PukKeyManagerSuite) SetUpSuite(c *C) {
	types.SetupConfigForTest()
}
func (s *PukKeyManagerSuite) TestPubKeyManager(c *C) {
	pkm := NewPubKeyManager()
	pk, err := common.NewPubKey("thorpub1addwnpepqfzklu4h6ztp39ndcvyys2v9ljqh496wnhrkgpawd5w3yqxxhvm8sjp85z9")
	c.Assert(err, IsNil)
	c.Assert(pkm.pks, HasLen, 0)

	pkm.Add(pk)
	c.Assert(pkm.pks, HasLen, 1)
	c.Assert(pkm.pks[0].Equals(pk), Equals, true)

	pkm.Remove(pk)
	c.Assert(pkm.pks, HasLen, 0)
}
