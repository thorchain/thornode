package ethereum

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
)

func Test(t *testing.T) { TestingT(t) }

type BlockScannerTestSuite struct {
	m *metrics.Metrics
}

var _ = Suite(&BlockScannerTestSuite{})

func (s *BlockScannerTestSuite) SetUpSuite(c *C) {
	s.m = GetMetricForTest(c)
	c.Assert(s.m, NotNil)
}

func getConfigForTest(rpcHost string) config.BlockScannerConfiguration {
	return config.BlockScannerConfiguration{
		RPCHost:                    rpcHost,
		StartBlockHeight:           0,
		BlockScanProcessors:        1,
		HttpRequestTimeout:         time.Second,
		HttpRequestReadTimeout:     time.Second * 30,
		HttpRequestWriteTimeout:    time.Second * 30,
		MaxHttpRequestRetry:        3,
		BlockHeightDiscoverBackoff: time.Second,
		BlockRetryInterval:         time.Second,
	}
}

func (s *BlockScannerTestSuite) TestNewBlockScanner(c *C) {
	c.Skip("skip")
	storage, err := blockscanner.NewBlockScannerStorage("")
	c.Assert(err, IsNil)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))
	ctx := context.Background()
	ethClient, err := ethclient.DialContext(ctx, server.URL)
	c.Assert(err, IsNil)
	bs, err := NewBlockScanner(getConfigForTest(""), storage, true, ethClient, s.m)
	c.Assert(err, NotNil)
	c.Assert(bs, IsNil)
	bs, err = NewBlockScanner(getConfigForTest("127.0.0.1"), storage, true, ethClient, s.m)
	c.Assert(err, NotNil)
	c.Assert(bs, IsNil)
	bs, err = NewBlockScanner(getConfigForTest("127.0.0.1"), storage, true, nil, s.m)
	c.Assert(err, NotNil)
	c.Assert(bs, IsNil)
	bs, err = NewBlockScanner(getConfigForTest("127.0.0.1"), storage, true, ethClient, s.m)
	c.Assert(err, NotNil)
	c.Assert(bs, IsNil)
	bs, err = NewBlockScanner(getConfigForTest("127.0.0.1"), storage, true, ethClient, s.m)
	c.Assert(err, IsNil)
	c.Assert(bs, NotNil)
}
