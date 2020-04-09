package blockscanner

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/pkg/chainclients/ethereum/types"
	"gitlab.com/thorchain/thornode/common"
)

func Test(t *testing.T) { TestingT(t) }

type CommonBlockScannerTestSuite struct {
	m *metrics.Metrics
}

var _ = Suite(&CommonBlockScannerTestSuite{})

func (s *CommonBlockScannerTestSuite) TestNewCommonBlockScanner(c *C) {
	mss := NewMockScannerStorage()
	cbs, err := NewCommonBlockScanner(config.BlockScannerConfiguration{
		RPCHost: "",
	}, 0, mss, nil, CosmosSupplemental{})
	c.Check(cbs, IsNil)
	c.Check(err, NotNil)
	cbs, err = NewCommonBlockScanner(config.BlockScannerConfiguration{
		RPCHost: "localhost",
	}, 0, mss, nil, CosmosSupplemental{})
	c.Check(cbs, IsNil)
	c.Check(err, NotNil)
	cbs, err = NewCommonBlockScanner(config.BlockScannerConfiguration{
		RPCHost: "localhost",
	}, 0, mss, m, CosmosSupplemental{})
	c.Check(cbs, NotNil)
	c.Check(err, IsNil)
}

func (s *CommonBlockScannerTestSuite) TestBlockScanner(c *C) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Logf("================>:%s", r.RequestURI)
		switch {
		case strings.HasPrefix(r.RequestURI, "/block"): // trying to get block
			if _, err := w.Write([]byte(blockResult)); err != nil {
				c.Error(err)
			}
		}
	})
	mss := NewMockScannerStorage()
	server := httptest.NewTLSServer(h)
	defer server.Close()
	cbs, err := NewCommonBlockScanner(config.BlockScannerConfiguration{
		RPCHost:                    server.URL,
		StartBlockHeight:           0,
		BlockScanProcessors:        1,
		HttpRequestTimeout:         time.Second,
		HttpRequestReadTimeout:     time.Second * 30,
		HttpRequestWriteTimeout:    time.Second * 30,
		MaxHttpRequestRetry:        3,
		BlockHeightDiscoverBackoff: time.Second,
		BlockRetryInterval:         time.Second,
		ChainID:                    common.BNBChain,
	}, 0, mss, m, CosmosSupplemental{})
	c.Check(cbs, NotNil)
	c.Check(err, IsNil)
	trSkipVerify := &http.Transport{
		MaxIdleConnsPerHost: 10,
		TLSClientConfig: &tls.Config{
			MaxVersion:         tls.VersionTLS11,
			InsecureSkipVerify: true,
		},
	}
	cbs.GetHttpClient().Transport = trSkipVerify
	var counter int
	go func() {
		for item := range cbs.GetMessages() {
			c.Logf("block height:%d", item)
			counter++
		}
	}()
	cbs.Start()
	time.Sleep(time.Second * 1)
	err = cbs.Stop()
	c.Check(err, IsNil)
	// c.Check(counter, Equals, 11)
}

func (s *CommonBlockScannerTestSuite) TestBadBlock(c *C) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Logf("================>:%s", r.RequestURI)
		switch {
		case strings.HasPrefix(r.RequestURI, "/block"): // trying to get block
			if _, err := w.Write([]byte(blockBadResult)); err != nil {
				c.Error(err)
			}
		}
	})
	mss := NewMockScannerStorage()
	server := httptest.NewTLSServer(h)
	defer server.Close()
	cbs, err := NewCommonBlockScanner(config.BlockScannerConfiguration{
		RPCHost:                    server.URL,
		StartBlockHeight:           0,
		BlockScanProcessors:        1,
		HttpRequestTimeout:         time.Second,
		HttpRequestReadTimeout:     time.Second * 30,
		HttpRequestWriteTimeout:    time.Second * 30,
		MaxHttpRequestRetry:        3,
		BlockHeightDiscoverBackoff: time.Second,
		BlockRetryInterval:         time.Second,
		ChainID:                    common.BNBChain,
	}, 0, mss, m, CosmosSupplemental{})
	c.Check(cbs, NotNil)
	c.Check(err, IsNil)
	trSkipVerify := &http.Transport{
		MaxIdleConnsPerHost: 10,
		TLSClientConfig: &tls.Config{
			MaxVersion:         tls.VersionTLS11,
			InsecureSkipVerify: true,
		},
	}
	cbs.GetHttpClient().Transport = trSkipVerify
	cbs.Start()
	time.Sleep(time.Second * 1)
	err = cbs.Stop()
	c.Check(err, IsNil)
	// metric, err := m.GetCounterVec(metrics.CommonBlockScannerError).GetMetricWithLabelValues("fail_unmarshal_block", s.URL+"/block")
	// c.Assert(err, IsNil)
	// c.Check(int(testutil.ToFloat64(metric)), Equals, 1)
}

func (s *CommonBlockScannerTestSuite) TestBadConnection(c *C) {
	mss := NewMockScannerStorage()
	cbs, err := NewCommonBlockScanner(config.BlockScannerConfiguration{
		RPCHost:                    "localhost:23450",
		StartBlockHeight:           0,
		BlockScanProcessors:        1,
		HttpRequestTimeout:         time.Second,
		HttpRequestReadTimeout:     time.Second * 30,
		HttpRequestWriteTimeout:    time.Second * 30,
		MaxHttpRequestRetry:        3,
		BlockHeightDiscoverBackoff: time.Second,
		BlockRetryInterval:         time.Second,
		ChainID:                    common.BNBChain,
	}, 0, mss, m, CosmosSupplemental{})
	c.Check(cbs, NotNil)
	c.Check(err, IsNil)
	trSkipVerify := &http.Transport{
		MaxIdleConnsPerHost: 10,
		TLSClientConfig: &tls.Config{
			MaxVersion:         tls.VersionTLS11,
			InsecureSkipVerify: true,
		},
	}
	cbs.GetHttpClient().Transport = trSkipVerify
	cbs.Start()
	time.Sleep(time.Second * 1)
	err = cbs.Stop()
	c.Check(err, IsNil)
	// metric, err := m.GetCounterVec(metrics.CommonBlockScannerError).GetMetricWithLabelValues("fail_get_block", "http://localhost:23450/block")
	// c.Assert(err, IsNil)
	// c.Check(int(testutil.ToFloat64(metric)), Equals, 1)
}

func (s *CommonBlockScannerTestSuite) TestGetHttp(c *C) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Logf("================>:%s", r.RequestURI)
		switch r.RequestURI {
		case "/block?height=1": // trying to get block
			if _, err := w.Write([]byte(blockResult)); err != nil {
				c.Error(err)
			}
		case "/block?height=2": // trying to get block
			if _, err := w.Write([]byte(blockError)); err != nil {
				c.Error(err)
			}
		}
	})
	mss := NewMockScannerStorage()
	server := httptest.NewTLSServer(h)
	defer server.Close()
	cbs, err := NewCommonBlockScanner(config.BlockScannerConfiguration{
		RPCHost:                    server.URL,
		StartBlockHeight:           0,
		BlockScanProcessors:        1,
		HttpRequestTimeout:         time.Second,
		HttpRequestReadTimeout:     time.Second * 30,
		HttpRequestWriteTimeout:    time.Second * 30,
		MaxHttpRequestRetry:        3,
		BlockHeightDiscoverBackoff: time.Second,
		BlockRetryInterval:         time.Second,
		ChainID:                    common.BNBChain,
	}, 0, mss, m, CosmosSupplemental{})
	c.Assert(err, IsNil)
	trSkipVerify := &http.Transport{
		MaxIdleConnsPerHost: 10,
		TLSClientConfig: &tls.Config{
			MaxVersion:         tls.VersionTLS11,
			InsecureSkipVerify: true,
		},
	}
	cbs.GetHttpClient().Transport = trSkipVerify

	_, err = cbs.getFromHttp(fmt.Sprintf("%s/block?height=1", server.URL), "")
	c.Assert(err, IsNil)

	_, err = cbs.getFromHttp(fmt.Sprintf("%s/block?height=2", server.URL), "")
	c.Assert(err, NotNil)
}

func (s *CommonBlockScannerTestSuite) TestGetETHBlocks(c *C) {
	h := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		c.Assert(err, IsNil)
		type RPCRequest struct {
			JSONRPC string          `json:"jsonrpc"`
			ID      interface{}     `json:"id"`
			Method  string          `json:"method"`
			Params  json.RawMessage `json:"params"`
		}
		var rpcRequest RPCRequest
		err = json.Unmarshal(body, &rpcRequest)
		c.Assert(err, IsNil)
		if rpcRequest.Method == "eth_getBlockByNumber" {
			_, err := rw.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{
				"difficulty": "0x31962a3fc82b",
				"extraData": "0x4477617266506f6f6c",
				"gasLimit": "0x47c3d8",
				"gasUsed": "0x0",
				"hash": "0x78bfef68fccd4507f9f4804ba5c65eb2f928ea45b3383ade88aaa720f1209cba",
				"logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
				"miner": "0x2a65aca4d5fc5b5c859090a6c34d164135398226",
				"nonce": "0xa5e8fb780cc2cd5e",
				"number": "0x1",
				"parentHash": "0x8b535592eb3192017a527bbf8e3596da86b3abea51d6257898b2ced9d3a83826",
				"receiptsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
				"sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
				"size": "0x20e",
				"stateRoot": "0xdc6ed0a382e50edfedb6bd296892690eb97eb3fc88fd55088d5ea753c48253dc",
				"timestamp": "0x579f4981",
				"totalDifficulty": "0x25cff06a0d96f4bee",
				"transactions": [],
				"transactionsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
				"uncles": [
		]}}`))
			c.Assert(err, IsNil)
		}
	})
	mss := NewMockScannerStorage()
	server := httptest.NewTLSServer(h)
	defer server.Close()
	supp := types.EthereumSupplemental{}
	cbs, err := NewCommonBlockScanner(config.BlockScannerConfiguration{
		RPCHost:                    server.URL,
		StartBlockHeight:           0,
		BlockScanProcessors:        1,
		HttpRequestTimeout:         time.Second,
		HttpRequestReadTimeout:     time.Second * 30,
		HttpRequestWriteTimeout:    time.Second * 30,
		MaxHttpRequestRetry:        3,
		BlockHeightDiscoverBackoff: time.Second,
		BlockRetryInterval:         time.Second,
		ChainID:                    common.ETHChain,
	}, 0, mss, m, supp)
	c.Assert(err, IsNil)
	trSkipVerify := &http.Transport{
		MaxIdleConnsPerHost: 10,
		TLSClientConfig: &tls.Config{
			MaxVersion:         tls.VersionTLS11,
			InsecureSkipVerify: true,
		},
	}
	cbs.GetHttpClient().Transport = trSkipVerify
	_, request := supp.BlockRequest(server.URL, 1)
	_, err = cbs.getFromHttp(server.URL, request)
	c.Assert(err, IsNil)
}
