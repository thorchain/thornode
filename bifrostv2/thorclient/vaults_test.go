package thorclient

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/helpers"
)

type VaultsSuite struct {
	server  *httptest.Server
	client  *Client
	cfg     config.ThorChainConfiguration
	cleanup func()
}

var _ = Suite(&VaultsSuite{})

func (s *VaultsSuite) SetUpSuite(c *C) {
	s.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.RequestURI {
		case "/thorchain/vaults/pubkeys":
			vaultsHandle(c, rw)
		}
	}))

	s.cfg, _, s.cleanup = helpers.SetupStateChainForTest(c)
	s.cfg.ChainHost = s.server.Listener.Addr().String()
	var err error
	s.client, err = NewClient(s.cfg, helpers.GetMetricForTest(c))
	c.Assert(err, IsNil)
	c.Assert(s.client, NotNil)
}

func (s *VaultsSuite) TearDownSuite(c *C) {
	s.cleanup()
	s.server.Close()
}

func vaultsHandle(c *C, rw http.ResponseWriter) {
	content, err := ioutil.ReadFile("../../test/fixtures/endpoints/vaults/pubKeys.json")
	if err != nil {
		c.Fatal(err)
	}

	rw.Header().Set("Content-Type", "application/json")
	if _, err := rw.Write(content); err != nil {
		c.Fatal(err)
	}
}

func (s *VaultsSuite) TestGetVaults(c *C) {
	vaults, err := s.client.GetVaults()
	c.Assert(err, IsNil)
	c.Assert(vaults, NotNil)
	c.Assert(vaults.Asgard[0].String(), Equals, "thorpub1addwnpepqflvfv08t6qt95lmttd6wpf3ss8wx63e9vf6fvyuj2yy6nnyna5763e2kck")
}
