package thorclient

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
	. "gopkg.in/check.v1"
)

type KeygenSuite struct {
	server  *httptest.Server
	bridge  *ThorchainBridge
	cfg     config.ClientConfiguration
	cleanup func()
	fixture string
}

var _ = Suite(&KeygenSuite{})

func (s *KeygenSuite) SetUpSuite(c *C) {
	s.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch {
		case strings.HasPrefix(req.RequestURI, KeygenEndpoint):
			httpTestHandler(c, rw, s.fixture)
		}
	}))

	s.cfg, _, s.cleanup = SetupStateChainForTest(c)
	s.cfg.ChainHost = s.server.Listener.Addr().String()
	var err error
	s.bridge, err = NewThorchainBridge(s.cfg, GetMetricForTest(c))
	s.bridge.httpClient.RetryMax = 1
	c.Assert(err, IsNil)
	c.Assert(s.bridge, NotNil)
}

func (s *KeygenSuite) TearDownSuite(c *C) {
	s.cleanup()
	s.server.Close()
}

func (s *KeygenSuite) TestGetKeygen(c *C) {
	s.fixture = "../../test/fixtures/endpoints/keygen/template.json"
	pk := types.GetRandomPubKey()
	expectedKey, err := common.NewPubKey("thorpub1addwnpepq2kdyjkm6y9aa3kxl8wfaverka6pvkek2ygrmhx6sj3ec6h0fegwsgeslue")
	c.Assert(err, IsNil)
	keygenBlock, err := s.bridge.GetKeygenBlock(1718, pk.String())
	c.Assert(err, IsNil)
	c.Assert(keygenBlock, NotNil)
	c.Assert(keygenBlock.Height, Equals, int64(1718))
	c.Assert(keygenBlock.Keygens[0].Members[0], Equals, expectedKey)
}
