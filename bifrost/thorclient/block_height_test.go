package thorclient

import (
	"net/http"
	"net/http/httptest"
	"strings"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/helpers"
)

type BlockHeightSuite struct {
	server  *httptest.Server
	bridge  *ThorchainBridge
	cfg     config.ThorchainConfiguration
	cleanup func()
	fixture string
}

var _ = Suite(&BlockHeightSuite{})

func (s *BlockHeightSuite) SetUpSuite(c *C) {
	s.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch {
		case strings.HasPrefix(req.RequestURI, LastBlockEndpoint):
			httpTestHandler(c, rw, s.fixture)
		}
	}))

	s.cfg, _, s.cleanup = helpers.SetupStateChainForTest(c)
	s.cfg.ChainHost = s.server.Listener.Addr().String()
	var err error
	s.bridge, err = NewThorchainBridge(s.cfg, helpers.GetMetricForTest(c))
	s.bridge.httpClient.RetryMax = 1
	c.Assert(err, IsNil)
	c.Assert(s.bridge, NotNil)
}

func (s *BlockHeightSuite) TearDownSuite(c *C) {
	s.cleanup()
	s.server.Close()
}

func (s *BlockHeightSuite) TestGetBlockHeight(c *C) {
	s.fixture = "../../test/fixtures/endpoints/lastblock/bnb.json"
	height, err := s.bridge.GetBlockHeight()
	c.Assert(err, IsNil)
	c.Assert(height, NotNil)
}

func (s *BlockHeightSuite) TestGetBlockHeight(c *C) {
	s.fixture = "../../test/fixtures/endpoints/lastblock/bnb.json"
	height, err := s.bridge.GetBlockHeight()
	c.Assert(err, IsNil)
	c.Assert(height, NotNil)
}
