package pubkeymanager

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

func Test(t *testing.T) { TestingT(t) }

type PubKeyMgrSuite struct{}

var _ = Suite(&PubKeyMgrSuite{})

func (s *PubKeyMgrSuite) TestPubkeyMgr(c *C) {
	pk1 := types.GetRandomPubKey()
	pk2 := types.GetRandomPubKey()
	pk3 := types.GetRandomPubKey()
	pk4 := types.GetRandomPubKey()

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

	// remove a pubkey
	pubkeyMgr.RemovePubKey(pk2)
	c.Check(pubkeyMgr.HasPubKey(pk2), Equals, false)

	addr, err := pk1.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)
	ok, _ := pubkeyMgr.IsValidPoolAddress(addr.String(), common.BNBChain)
	c.Assert(ok, Equals, true)

	addr, err = pk3.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)
	pubkeyMgr.AddNodePubKey(pk4)
	c.Check(pubkeyMgr.GetNodePubKey().String(), Equals, pk4.String())
	ok, _ = pubkeyMgr.IsValidPoolAddress(addr.String(), common.BNBChain)
	c.Assert(ok, Equals, false)
}

func (s *PubKeyMgrSuite) TestFetchKeys(c *C) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Logf("================>:%s", r.RequestURI)
		switch r.RequestURI {
		case "/thorchain/vaults/pubkeys":
			if _, err := w.Write([]byte(`{"asgard": [], "yggdrasil": []}`)); err != nil {
				c.Error(err)
			}
		}
	}))
	pubkeyMgr, err := NewPubKeyManager(server.URL[7:], nil)
	c.Assert(err, IsNil)
	err = pubkeyMgr.Start()
	c.Assert(err, IsNil)
	pubkeyMgr.FetchPubKeys()
	pubKeys := pubkeyMgr.GetPubKeys()
	c.Check(len(pubKeys), Equals, 0)
	err = pubkeyMgr.Stop()
	c.Assert(err, IsNil)
}
