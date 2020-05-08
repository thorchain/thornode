package thorclient

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/cmd"
	"gitlab.com/thorchain/thornode/common"
	stypes "gitlab.com/thorchain/thornode/x/thorchain/types"
)

func TestPackage(t *testing.T) { TestingT(t) }

type ThorchainSuite struct {
	server             *httptest.Server
	cfg                config.ClientConfiguration
	cleanup            func()
	bridge             *ThorchainBridge
	authAccountFixture string
	nodeAccountFixture string
}

var _ = Suite(&ThorchainSuite{})

func (s *ThorchainSuite) SetUpSuite(c *C) {
	cfg2 := sdk.GetConfig()
	cfg2.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	s.cfg, _, s.cleanup = SetupStateChainForTest(c)
	s.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch {
		case strings.HasPrefix(req.RequestURI, AuthAccountEndpoint):
			httpTestHandler(c, rw, s.authAccountFixture)
		case strings.HasPrefix(req.RequestURI, NodeAccountEndpoint):
			httpTestHandler(c, rw, s.nodeAccountFixture)
		case strings.HasPrefix(req.RequestURI, LastBlockEndpoint):
			httpTestHandler(c, rw, "../../test/fixtures/endpoints/lastblock/bnb.json")
		case strings.HasPrefix(req.RequestURI, StatusEndpoint):
			httpTestHandler(c, rw, "../../test/fixtures/endpoints/status/status.json")
		case strings.HasPrefix(req.RequestURI, KeysignEndpoint):
			httpTestHandler(c, rw, "../../test/fixtures/endpoints/keysign/template.json")
		case strings.HasPrefix(req.RequestURI, "/thorchain/vaults") && strings.HasSuffix(req.RequestURI, "/signers"):
			httpTestHandler(c, rw, "../../test/fixtures/endpoints/tss/keysign_party.json")
		case strings.HasPrefix(req.RequestURI, AsgardVault):
			httpTestHandler(c, rw, "../../test/fixtures/endpoints/vaults/asgard.json")
		}
	}))
	s.cfg.ChainHost = s.server.Listener.Addr().String()
	s.cfg.ChainRPC = s.server.Listener.Addr().String()

	var err error
	s.bridge, err = NewThorchainBridge(s.cfg, GetMetricForTest(c))
	s.bridge.httpClient.RetryMax = 1 // fail fast
	c.Assert(err, IsNil)
	c.Assert(s.bridge, NotNil)
}

func (s *ThorchainSuite) TearDownSuite(c *C) {
	s.server.Close()
	s.cleanup()
}

func (s *ThorchainSuite) TestGetThorChainURL(c *C) {
	uri := s.bridge.getThorChainURL("")
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

func (s *ThorchainSuite) TestGet(c *C) {
	buf, status, err := s.bridge.getWithPath("")
	c.Check(status, Equals, http.StatusOK)
	c.Assert(err, IsNil)
	c.Assert(buf, NotNil)
}

func (s *ThorchainSuite) TestSign(c *C) {
	pk := stypes.GetRandomPubKey()
	vaultAddr, err := pk.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)
	tx := stypes.NewObservedTx(
		common.Tx{
			Coins: common.Coins{
				common.NewCoin(common.BNBAsset, sdk.NewUint(123400000)),
			},
			Memo:        "This is my memo!",
			FromAddress: vaultAddr,
			ToAddress:   common.Address("bnb1ntqj0v0sv62ut0ehxt7jqh7lenfrd3hmfws0aq"),
		},
		1,
		pk,
	)

	signedMsg, err := s.bridge.GetObservationsStdTx(stypes.ObservedTxs{tx})
	c.Log(err)
	c.Assert(signedMsg, NotNil)
	c.Assert(err, IsNil)
}

func (ThorchainSuite) TestNewThorchainBridge(c *C) {
	testFunc := func(cfg config.ClientConfiguration, errChecker, sbChecker Checker) {
		sb, err := NewThorchainBridge(cfg, m)
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

func (s *ThorchainSuite) TestGetAccountNumberAndSequenceNumber_Success(c *C) {
	s.nodeAccountFixture = "../../test/fixtures/endpoints/nodeaccount/template.json"
	s.authAccountFixture = "../../test/fixtures/endpoints/auth/accounts/template.json"
	accNumber, sequence, err := s.bridge.getAccountNumberAndSequenceNumber()
	c.Assert(err, IsNil)
	c.Assert(accNumber, Equals, uint64(3))
	c.Assert(sequence, Equals, uint64(5))
}

func (s *ThorchainSuite) TestGetAccountNumberAndSequenceNumber_Fail(c *C) {
	s.nodeAccountFixture = "../../test/fixtures/endpoints/nodeaccount/template.json"
	s.authAccountFixture = ""
	accNumber, sequence, err := s.bridge.getAccountNumberAndSequenceNumber()
	c.Assert(err, NotNil)
	c.Assert(accNumber, Equals, uint64(0))
	c.Assert(sequence, Equals, uint64(0))
}

func (s *ThorchainSuite) TestGetAccountNumberAndSequenceNumber_Fail_500(c *C) {
	s.nodeAccountFixture = "../../test/fixtures/endpoints/nodeaccount/template.json"
	s.authAccountFixture = "500"
	accNumber, sequence, err := s.bridge.getAccountNumberAndSequenceNumber()
	c.Assert(err, NotNil)
	c.Assert(accNumber, Equals, uint64(0))
	c.Assert(sequence, Equals, uint64(0))
}

func (s *ThorchainSuite) TestGetAccountNumberAndSequenceNumber_Fail_Unmarshal(c *C) {
	s.nodeAccountFixture = "../../test/fixtures/endpoints/nodeaccount/template.json"
	s.authAccountFixture = "../../test/fixtures/endpoints/auth/accounts/malformed.json"
	accNumber, sequence, err := s.bridge.getAccountNumberAndSequenceNumber()
	c.Assert(err, NotNil)
	c.Assert(true, Equals, strings.HasPrefix(err.Error(), "failed to unmarshal account resp"))
	c.Assert(accNumber, Equals, uint64(0))
	c.Assert(sequence, Equals, uint64(0))
}

func (s *ThorchainSuite) TestGetAccountNumberAndSequenceNumber_Fail_AccNumberString(c *C) {
	s.nodeAccountFixture = "../../test/fixtures/endpoints/nodeaccount/template.json"
	s.authAccountFixture = "../../test/fixtures/endpoints/auth/accounts/accnumber_string.json"
	accNumber, sequence, err := s.bridge.getAccountNumberAndSequenceNumber()
	c.Assert(err, NotNil)
	c.Assert(true, Equals, strings.HasPrefix(err.Error(), "failed to parse account number"))
	c.Assert(accNumber, Equals, uint64(0))
	c.Assert(sequence, Equals, uint64(0))
}

func (s *ThorchainSuite) TestGetAccountNumberAndSequenceNumber_Fail_SequenceString(c *C) {
	s.nodeAccountFixture = "../../test/fixtures/endpoints/nodeaccount/template.json"
	s.authAccountFixture = "../../test/fixtures/endpoints/auth/accounts/seqnumber_string.json"
	accNumber, sequence, err := s.bridge.getAccountNumberAndSequenceNumber()
	c.Assert(err, NotNil)
	c.Assert(true, Equals, strings.HasPrefix(err.Error(), "failed to parse sequence number"))
	c.Assert(accNumber, Equals, uint64(0))
	c.Assert(sequence, Equals, uint64(0))
}

func (s *ThorchainSuite) TestEnsureNodeWhitelisted_Success(c *C) {
	s.authAccountFixture = "../../test/fixtures/endpoints/auth/accounts/template.json"
	s.nodeAccountFixture = "../../test/fixtures/endpoints/nodeaccount/template.json"
	err := s.bridge.EnsureNodeWhitelisted()
	c.Assert(err, IsNil)
}

func (s *ThorchainSuite) TestEnsureNodeWhitelisted_Fail(c *C) {
	s.authAccountFixture = "../../test/fixtures/endpoints/auth/accounts/template.json"
	s.nodeAccountFixture = "../../test/fixtures/endpoints/nodeaccount/disabled.json"
	err := s.bridge.EnsureNodeWhitelisted()
	c.Assert(err, NotNil)
}

func (s *ThorchainSuite) TestGetKeysignParty(c *C) {
	pubKey := stypes.GetRandomPubKey()
	pubKeys, err := s.bridge.GetKeysignParty(pubKey)
	c.Assert(err, IsNil)
	c.Assert(pubKeys, HasLen, 3)
}

func (s *ThorchainSuite) TestIsCatchingUp(c *C) {
	ok, err := s.bridge.IsCatchingUp()
	c.Assert(err, IsNil)
	c.Assert(ok, Equals, false)
}

func (s *ThorchainSuite) TestGetAsgards(c *C) {
	vaults, err := s.bridge.GetAsgards()
	c.Assert(err, IsNil)
	c.Assert(vaults, NotNil)
}
