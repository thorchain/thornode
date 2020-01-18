package thorchain

import (
	"net/http"
	"net/http/httptest"
	"strings"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/helpers"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

type KeysignSuite struct {
	server  *httptest.Server
	client  *Client
	cfg     config.ClientConfiguration
	cleanup func()
	fixture string
}

var _ = Suite(&KeysignSuite{})

func (s *KeysignSuite) SetUpSuite(c *C) {
	s.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch {
		case strings.HasPrefix(req.RequestURI, AuthAccountEndpoint):
			httpTestHandler(c, rw, "../../test/fixtures/endpoints/auth/accounts/template.json")
		case strings.HasPrefix(req.RequestURI, NodeAccountEndpoint):
			httpTestHandler(c, rw, "../../test/fixtures/endpoints/nodeaccount/template.json")
		case strings.HasPrefix(req.RequestURI, LastBlockEndpoint):
			httpTestHandler(c, rw, "../../test/fixtures/endpoints/lastblock/bnb.json")
		case strings.HasPrefix(req.RequestURI, KeysignEndpoint):
			httpTestHandler(c, rw, s.fixture)
		}
	}))

	s.cfg, _, s.cleanup = helpers.SetupStateChainForTest(c)
	s.cfg.ChainHost = s.server.Listener.Addr().String()
	var err error
	s.client, err = NewClient(s.cfg, helpers.GetMetricForTest(c))
	c.Assert(err, IsNil)
	// fail fast
	s.client.httpClient.RetryMax = 1
}

func (s *KeysignSuite) TearDownSuite(c *C) {
	s.cleanup()
	s.server.Close()
}

func (s *KeysignSuite) TestGetKeysign(c *C) {
	s.fixture = "../../test/fixtures/endpoints/keysign/template.json"
	pk := types.GetRandomPubKey()
	keysign, err := s.client.GetKeysign(1718, pk.String())
	c.Assert(err, IsNil)
	c.Assert(keysign, NotNil)
	c.Assert(keysign.Height, Equals, uint64(1718))
	c.Assert(keysign.TxArray[0].Chain, Equals, common.BNBChain)
}
