package thorclient

import (
	"net/http"
	"net/http/httptest"
	"strings"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/helpers"
)

type NodeAccountSuite struct {
	server  *httptest.Server
	client  *Client
	cfg     config.ThorChainConfiguration
	cleanup func()
	fixture string
}

var _ = Suite(&NodeAccountSuite{})

func (s *NodeAccountSuite) SetUpSuite(c *C) {
	s.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch {
		case strings.HasPrefix(req.RequestURI, NodeAccountEndpoint):
			httpTestHandler(c, rw, s.fixture)
		}
	}))

	s.cfg, _, s.cleanup = helpers.SetupStateChainForTest(c)
	s.cfg.ChainHost = s.server.Listener.Addr().String()
	var err error
	s.client, err = NewClient(s.cfg, helpers.GetMetricForTest(c))
	c.Assert(err, IsNil)
	c.Assert(s.client, NotNil)
}

func (s *NodeAccountSuite) TearDownSuite(c *C) {
	s.cleanup()
	s.server.Close()
}

func (s *NodeAccountSuite) TestGetNodeAccount(c *C) {
	s.fixture = "../../test/fixtures/endpoints/nodeaccount/template.json"
	na, err := s.client.GetNodeAccount(s.client.keys.GetSignerInfo().GetAddress().String())
	c.Assert(err, IsNil)
	c.Assert(na, NotNil)
}
