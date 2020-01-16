package thorclient

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/helpers"
	"gitlab.com/thorchain/thornode/common"
)

type KeygenSuite struct {
	server  *httptest.Server
	client  *Client
	cfg     config.ThorChainConfiguration
	cleanup func()
	fixture string
}

var _ = Suite(&KeygenSuite{})

func (s *KeygenSuite) SetUpSuite(c *C) {
	s.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch {
		case strings.HasPrefix(req.RequestURI, AuthAccountEndpoint):
			httpTestHandler(c, rw, "../../test/fixtures/endpoints/auth/accounts/template.json")
		case strings.HasPrefix(req.RequestURI, NodeAccountEndpoint):
			httpTestHandler(c, rw, "../../test/fixtures/endpoints/nodeaccount/template.json")
		case strings.HasPrefix(req.RequestURI, LastBlockEndpoint):
			httpTestHandler(c, rw, "../../test/fixtures/endpoints/lastblock/bnb.json")
		case strings.HasPrefix(req.RequestURI, KeygenEndpoint):
			httpTestHandler(c, rw, s.fixture)
		}
	}))

	s.cfg, _, s.cleanup = helpers.SetupStateChainForTest(c)
	s.cfg.ChainHost = s.server.Listener.Addr().String()
	var err error
	s.client, err = NewClient(s.cfg, helpers.GetMetricForTest(c))
	// fail fast
	s.client.client.RetryMax = 1
	c.Assert(err, IsNil)
	c.Assert(s.client, NotNil)
}

func (s *KeygenSuite) TearDownSuite(c *C) {
	s.cleanup()
	s.server.Close()
}

func keygenHandle(c *C, rw http.ResponseWriter) {
	content, err := ioutil.ReadFile("../../test/fixtures/endpoints/vaults/pubKeys.json")
	if err != nil {
		c.Fatal(err)
	}

	rw.Header().Set("Content-Type", "application/json")
	if _, err := rw.Write(content); err != nil {
		c.Fatal(err)
	}
}

func (s *KeygenSuite) TestGetKeygen(c *C) {
	s.fixture = "../../test/fixtures/endpoints/keygen/template.json"
	err := s.client.getPubKeys()
	c.Assert(err, IsNil)
	pk := s.client.pkm.GetPks()[0]
	expectedKey, err := common.NewPubKey("thorpub1addwnpepq2kdyjkm6y9aa3kxl8wfaverka6pvkek2ygrmhx6sj3ec6h0fegwsgeslue")
	c.Assert(err, IsNil)
	keygens, err := s.client.GetKeygens(1718, pk.String())
	c.Assert(err, IsNil)
	c.Assert(keygens, NotNil)
	c.Assert(keygens.Height, Equals, uint64(1718))
	c.Assert(keygens.Keygens[0][0], Equals, expectedKey)
}
