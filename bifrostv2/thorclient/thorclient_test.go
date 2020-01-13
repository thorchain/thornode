package thorclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/helpers"
)

func TestPackage(t *testing.T) { TestingT(t) }

type ThorClientSuite struct {
	server  *httptest.Server
	cfg     config.ThorChainConfiguration
	cleanup func()
	client  *Client
}

var _ = Suite(&ThorClientSuite{})

func (s *ThorClientSuite) SetUpSuite(c *C) {
	s.cfg, _, s.cleanup = helpers.SetupStateChainForTest(c)
	s.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))
	s.cfg.ChainHost = s.server.Listener.Addr().String()

	var err error
	s.client, err = NewClient(s.cfg, helpers.GetMetricForTest(c))
	c.Assert(err, IsNil)
	c.Assert(s.client, NotNil)
}

func (s *ThorClientSuite) TearDownSuite(c *C) {
	s.server.Close()
	s.cleanup()
}

func (s *ThorClientSuite) TestGetThorChainUrl(c *C) {
	uri := s.client.getThorChainUrl("")
	c.Assert(uri, Equals, "http://"+s.server.Listener.Addr().String())
}

func (s *ThorClientSuite) TestGet(c *C) {
	buf, err := s.client.get("")
	c.Assert(err, IsNil)
	c.Assert(buf, NotNil)
}

func (s *ThorClientSuite) TestNewStateChainBridge(c *C) {
	var testFunc = func(cfg config.ThorChainConfiguration, errChecker Checker, sbChecker Checker) {
		sb, err := NewClient(cfg, helpers.GetMetricForTest(c))
		c.Assert(err, errChecker)
		c.Assert(sb, sbChecker)
	}
	testFunc(config.ThorChainConfiguration{
		ChainID:         "",
		ChainHost:       "localhost",
		ChainHomeFolder: "~/.thorcli",
		SignerName:      "signer",
		SignerPasswd:    "signerpassword",
	}, NotNil, IsNil)
	testFunc(config.ThorChainConfiguration{
		ChainID:         "chainid",
		ChainHost:       "",
		ChainHomeFolder: "~/.thorcli",
		SignerName:      "signer",
		SignerPasswd:    "signerpassword",
	}, NotNil, IsNil)
	testFunc(config.ThorChainConfiguration{
		ChainID:         "chainid",
		ChainHost:       "localhost",
		ChainHomeFolder: "~/.thorcli",
		SignerName:      "",
		SignerPasswd:    "signerpassword",
	}, NotNil, IsNil)
	testFunc(config.ThorChainConfiguration{
		ChainID:         "chainid",
		ChainHost:       "localhost",
		ChainHomeFolder: "~/.thorcli",
		SignerName:      "signer",
		SignerPasswd:    "",
	}, NotNil, IsNil)
}
