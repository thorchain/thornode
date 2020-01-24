package observer

import (
	"testing"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type PubKeyMgrSuite struct{}

var _ = Suite(&PubKeyMgrSuite{})

func (s *PubKeyMgrSuite) TestPubkeyMgr(c *C) {
	pk1 := types.GetRandomPubKey()
	pk2 := types.GetRandomPubKey()
	pk3 := types.GetRandomPubKey()

	pubkeyMgr, err := NewPubKeyManager("localhost:1317", nil)
	c.Assert(err, IsNil)
	c.Check(pubkeyMgr.HasPubKey(pk1), Equals, false)
	pubkeyMgr.AddPubKey(pk1, true)
	c.Check(pubkeyMgr.HasPubKey(pk1), Equals, true)
	c.Check(pubkeyMgr.pubkeys[0].PubKey.Equals(pk1), Equals, true)
	c.Check(pubkeyMgr.pubkeys[0].Signer, Equals, true)

	pubkeyMgr.AddPubKey(pk2, false)
	c.Check(pubkeyMgr.HasPubKey(pk2), Equals, true)
	c.Check(pubkeyMgr.pubkeys[1].PubKey.Equals(pk2), Equals, true)
	c.Check(pubkeyMgr.pubkeys[1].Signer, Equals, false)

	pks := pubkeyMgr.GetPubKeys()
	c.Assert(pks, HasLen, 2)

	pks = pubkeyMgr.GetSignPubKeys()
	c.Assert(pks, HasLen, 1)
	c.Check(pks[0].Equals(pk1), Equals, true)

	addr, err := pk1.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)
	ok, _ := pubkeyMgr.IsValidPoolAddress(addr.String(), common.BNBChain)
	c.Assert(ok, Equals, true)

	addr, err = pk3.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)
	ok, _ = pubkeyMgr.IsValidPoolAddress(addr.String(), common.BNBChain)
	c.Assert(ok, Equals, false)
}
