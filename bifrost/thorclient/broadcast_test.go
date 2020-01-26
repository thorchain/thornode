package thorclient

import (
	"net/http"
	"net/http/httptest"
	"strings"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
)

type BroadcastSuite struct {
	server  *httptest.Server
	bridge  *ThorchainBridge
	cfg     config.ClientConfiguration
	cleanup func()
	fixture string
}

var _ = Suite(&BroadcastSuite{})

func (s *BroadcastSuite) SetUpSuite(c *C) {
	s.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch {
		case strings.HasPrefix(req.RequestURI, AuthAccountEndpoint):
			httpTestHandler(c, rw, "../../test/fixtures/endpoints/auth/accounts/template.json")
		case strings.HasPrefix(req.RequestURI, LastBlockEndpoint):
			httpTestHandler(c, rw, "../../test/fixtures/endpoints/lastblock/bnb.json")
		case strings.HasPrefix(req.RequestURI, BroadcastTxsEndpoint):
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

func (s *BroadcastSuite) TearDownSuite(c *C) {
	s.cleanup()
	s.server.Close()
}

func (s *BroadcastSuite) TestBroadcast(c *C) {
	s.fixture = "../../test/fixtures/endpoints/txs/template.json"
	stdTx := authtypes.StdTx{}
	mode := types.TxSync
	txID, err := s.bridge.Broadcast(stdTx, mode)
	c.Assert(err, IsNil)
	c.Check(
		txID.String(),
		Equals,
		"E43FA2330C4317ECC084B0C6044DFE75AAE1FAB8F84A66107809E9739D02F80D",
	)
	c.Check(s.bridge.accountNumber, Equals, uint64(3))
	c.Check(s.bridge.seqNumber, Equals, uint64(6))
}
