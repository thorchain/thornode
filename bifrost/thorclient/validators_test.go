package thorclient

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"gitlab.com/thorchain/thornode/bifrost/config"
	. "gopkg.in/check.v1"
)

type ValidatorsSuite struct {
	server  *httptest.Server
	bridge  *ThorchainBridge
	cfg     config.ClientConfiguration
	cleanup func()
	fixture string
}

var _ = Suite(&ValidatorsSuite{})

func (s *ValidatorsSuite) SetUpSuite(c *C) {
	s.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch {
		case strings.HasPrefix(req.RequestURI, ValidatorsEndpoint):
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

func (s *ValidatorsSuite) TearDownSuite(c *C) {
	s.cleanup()
	s.server.Close()
}

func (s *ValidatorsSuite) TestGetValidators(c *C) {
	s.fixture = "../../test/fixtures/endpoints/validators/template.json"
	resp, err := s.bridge.GetValidators()
	c.Assert(err, IsNil)
	c.Assert(resp, NotNil)
	c.Assert(resp.Nominated, IsNil)
	c.Assert(resp.Queued, IsNil)
	c.Assert(resp.RotateWindowOpenAt, Equals, uint64(16081))
	c.Assert(resp.RotateAt, Equals, uint64(17281))
	c.Assert(resp.ActiveNodes, HasLen, 1)
}
