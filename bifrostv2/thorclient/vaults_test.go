package thorclient

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/x/thorchain"
)

func TestVaults(t *testing.T) {
	TestingT(t)
}

type VaultsSuite struct {
	server *httptest.Server
	client *Client
}

var _ = Suite(&VaultsSuite{})

func (s *VaultsSuite) SetUpSuite(c *C) {
	fmt.Println("SetUpSuite!!")
	thorchain.SetupConfigForTest()
	s.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.RequestURI {
		case "/thorchain/vaults/pubkeys":
			vaultsHandle(c, rw)
		}
	}))

	cfg, _, cleanup := SetupStateChainForTest(c)
	defer cleanup()
	cfg.ChainHost = s.server.URL
	var err error
	s.client, err = NewClient(cfg, getMetricForTest(c))
	c.Assert(err, IsNil)
	c.Assert(s.client, NotNil)
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
