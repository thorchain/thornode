package thorchain

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/helpers"
)

func TestPackage(t *testing.T) { TestingT(t) }

type ThorchainClientSuite struct {
	server             *httptest.Server
	cfg                config.ClientConfiguration
	cleanup            func()
	client             *Client
	authAccountFixture string
}

var _ = Suite(&ThorchainClientSuite{})

func (s *ThorchainClientSuite) SetUpSuite(c *C) {
	s.cfg, _, s.cleanup = helpers.SetupStateChainForTest(c)
	s.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch {
		case strings.HasPrefix(req.RequestURI, AuthAccountEndpoint):
			httpTestHandler(c, rw, s.authAccountFixture)
		case strings.HasPrefix(req.RequestURI, NodeAccountEndpoint):
			httpTestHandler(c, rw, "../../test/fixtures/endpoints/nodeaccount/template.json")
		case strings.HasPrefix(req.RequestURI, LastBlockEndpoint):
			httpTestHandler(c, rw, "../../test/fixtures/endpoints/lastblock/bnb.json")
		case strings.HasPrefix(req.RequestURI, KeysignEndpoint):
			httpTestHandler(c, rw, "../../test/fixtures/endpoints/keysign/template.json")
		}
	}))
	s.cfg.ChainHost = s.server.Listener.Addr().String()

	var err error
	s.client, err = NewClient(s.cfg, helpers.GetMetricForTest(c))
	s.client.httpClient.RetryMax = 1
	c.Assert(err, IsNil)
	c.Assert(s.client, NotNil)
}

func (s *ThorchainClientSuite) TearDownSuite(c *C) {
	s.server.Close()
	s.cleanup()
}

func (s *ThorchainClientSuite) TestGetThorChainURL(c *C) {
	uri := s.client.getThorChainURL("")
	c.Assert(uri, Equals, "http://"+s.server.Listener.Addr().String())
}

func httpTestHandler(c *C, rw http.ResponseWriter, fixture string) {
	var content []byte
	var err error

	switch fixture {
	case "500":
		rw.WriteHeader(http.StatusInternalServerError)
	default:
		content, err = ioutil.ReadFile(fixture)
		if err != nil {
			c.Fatal(err)
		}
	}

	rw.Header().Set("Content-Type", "application/json")
	if _, err := rw.Write(content); err != nil {
		c.Fatal(err)
	}
}

func (s *ThorchainClientSuite) TestGet(c *C) {
	buf, err := s.client.get("")
	c.Assert(err, IsNil)
	c.Assert(buf, NotNil)
}

func (s *ThorchainClientSuite) TestNewClient(c *C) {
	var testFunc = func(cfg config.ClientConfiguration, errChecker Checker, sbChecker Checker) {
		sb, err := NewClient(cfg, helpers.GetMetricForTest(c))
		c.Assert(err, errChecker)
		c.Assert(sb, sbChecker)
	}
	testFunc(config.ClientConfiguration{
		ChainID:         "",
		ChainHost:       "localhost",
		ChainHomeFolder: "~/.thorcli",
		SignerName:      "signer",
		SignerPasswd:    "signerpassword",
	}, NotNil, IsNil)
	testFunc(config.ClientConfiguration{
		ChainID:         "chainid",
		ChainHost:       "",
		ChainHomeFolder: "~/.thorcli",
		SignerName:      "signer",
		SignerPasswd:    "signerpassword",
	}, NotNil, IsNil)
	testFunc(config.ClientConfiguration{
		ChainID:         "chainid",
		ChainHost:       "localhost",
		ChainHomeFolder: "~/.thorcli",
		SignerName:      "",
		SignerPasswd:    "signerpassword",
	}, NotNil, IsNil)
	testFunc(config.ClientConfiguration{
		ChainID:         "chainid",
		ChainHost:       "localhost",
		ChainHomeFolder: "~/.thorcli",
		SignerName:      "signer",
		SignerPasswd:    "",
	}, NotNil, IsNil)
}

func (s *ThorchainClientSuite) TestGetAccountNumberAndSequenceNumber_Success(c *C) {
	s.authAccountFixture = "../../test/fixtures/endpoints/auth/accounts/template.json"
	accNumber, sequence, err := s.client.getAccountNumberAndSequenceNumber()
	c.Assert(err, IsNil)
	c.Assert(accNumber, Equals, uint64(3))
	c.Assert(sequence, Equals, uint64(5))
}

func (s *ThorchainClientSuite) TestGetAccountNumberAndSequenceNumber_Fail(c *C) {
	s.authAccountFixture = ""
	accNumber, sequence, err := s.client.getAccountNumberAndSequenceNumber()
	c.Assert(err, NotNil)
	c.Assert(accNumber, Equals, uint64(0))
	c.Assert(sequence, Equals, uint64(0))
}

func (s *ThorchainClientSuite) TestGetAccountNumberAndSequenceNumber_Fail_500(c *C) {
	s.authAccountFixture = "500"
	accNumber, sequence, err := s.client.getAccountNumberAndSequenceNumber()
	c.Assert(err, NotNil)
	c.Assert(accNumber, Equals, uint64(0))
	c.Assert(sequence, Equals, uint64(0))
}

func (s *ThorchainClientSuite) TestGetAccountNumberAndSequenceNumber_Fail_Unmarshal(c *C) {
	s.authAccountFixture = "../../test/fixtures/endpoints/auth/accounts/malformed.json"
	accNumber, sequence, err := s.client.getAccountNumberAndSequenceNumber()
	c.Assert(err, NotNil)
	c.Assert(true, Equals, strings.HasPrefix(err.Error(), "failed to unmarshal account resp"))
	c.Assert(accNumber, Equals, uint64(0))
	c.Assert(sequence, Equals, uint64(0))
}

func (s *ThorchainClientSuite) TestGetAccountNumberAndSequenceNumber_Fail_AccNumberString(c *C) {
	s.authAccountFixture = "../../test/fixtures/endpoints/auth/accounts/accnumber_string.json"
	accNumber, sequence, err := s.client.getAccountNumberAndSequenceNumber()
	c.Assert(err, NotNil)
	c.Assert(true, Equals, strings.HasPrefix(err.Error(), "failed to unmarshal base account"))
	c.Assert(accNumber, Equals, uint64(0))
	c.Assert(sequence, Equals, uint64(0))
}

func (s *ThorchainClientSuite) TestGetAccountNumberAndSequenceNumber_Fail_SequenceString(c *C) {
	s.authAccountFixture = "../../test/fixtures/endpoints/auth/accounts/seqnumber_string.json"
	accNumber, sequence, err := s.client.getAccountNumberAndSequenceNumber()
	c.Assert(err, NotNil)
	c.Assert(true, Equals, strings.HasPrefix(err.Error(), "failed to unmarshal base account"))
	c.Assert(accNumber, Equals, uint64(0))
	c.Assert(sequence, Equals, uint64(0))
}

func (s *ThorchainClientSuite) TestStart(c *C) {
	s.authAccountFixture = "../../test/fixtures/endpoints/auth/accounts/template.json"
	err := s.client.Start()
	c.Assert(err, IsNil)
	time.Sleep(time.Second)
	err = s.client.Stop()
	c.Assert(err, IsNil)
}
