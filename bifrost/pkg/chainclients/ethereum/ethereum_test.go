package ethereum

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client/keys"
	cKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/common"

	types2 "gitlab.com/thorchain/thornode/x/thorchain/types"
)

func TestPackage(t *testing.T) { TestingT(t) }

type EthereumSuite struct {
	thordir  string
	thorKeys *thorclient.Keys
	bridge   *thorclient.ThorchainBridge
	m        *metrics.Metrics
}

var _ = Suite(&EthereumSuite{})

var m *metrics.Metrics

func GetMetricForTest(c *C) *metrics.Metrics {
	if m == nil {
		var err error
		m, err = metrics.NewMetrics(config.MetricsConfiguration{
			Enabled:      false,
			ListenPort:   9000,
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
			Chains:       common.Chains{common.ETHChain},
		})
		c.Assert(m, NotNil)
		c.Assert(err, IsNil)
	}
	return m
}

func (s *EthereumSuite) SetUpSuite(c *C) {
	s.m = GetMetricForTest(c)
	c.Assert(s.m, NotNil)
	ns := strconv.Itoa(time.Now().Nanosecond())
	types2.SetupConfigForTest()
	c.Assert(os.Setenv("NET", "testnet"), IsNil)

	s.thordir = filepath.Join(os.TempDir(), ns, ".thorcli")
	cfg := config.ClientConfiguration{
		ChainID:         "thorchain",
		ChainHost:       "localhost",
		SignerName:      "bob",
		SignerPasswd:    "password",
		ChainHomeFolder: s.thordir,
	}

	kb, err := keys.NewKeyBaseFromDir(s.thordir)
	c.Assert(err, IsNil)
	_, _, err = kb.CreateMnemonic(cfg.SignerName, cKeys.English, cfg.SignerPasswd, cKeys.Secp256k1)
	c.Assert(err, IsNil)
	s.thorKeys, err = thorclient.NewKeys(cfg.ChainHomeFolder, cfg.SignerName, cfg.SignerPasswd)
	c.Assert(err, IsNil)
	s.bridge, err = thorclient.NewThorchainBridge(cfg, s.m)
	c.Assert(err, IsNil)
}

func (s *EthereumSuite) TearDownSuite(c *C) {
	c.Assert(os.Unsetenv("NET"), IsNil)

	if err := os.RemoveAll(s.thordir); err != nil {
		c.Error(err)
	}
}

var account = "0xde0b295669a9fd93d5f28d9ec85e40f4cb697bae"

func (s *EthereumSuite) TestNewClient(c *C) {
	e, err := NewClient(s.thorKeys, config.ChainConfiguration{}, nil, s.bridge)
	c.Assert(e, IsNil)
	c.Assert(err, NotNil)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
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
		if rpcRequest.Method == "eth_getBalance" {
			_, err := rw.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x3b9aca00"}`))
			c.Assert(err, IsNil)
		}
		if rpcRequest.Method == "eth_getTransactionCount" {
			_, err := rw.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x0"}`))
			c.Assert(err, IsNil)
		}
		if rpcRequest.Method == "eth_chainId" {
			_, err := rw.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x2"}`))
			c.Assert(err, IsNil)
		}
		if rpcRequest.Method == "eth_gasPrice" {
			_, err := rw.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0xd"}`))
			c.Assert(err, IsNil)
		}
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
	}))
	splitted := strings.SplitAfter(server.URL, ":")
	e2, err2 := NewClient(s.thorKeys, config.ChainConfiguration{RPCHost: "localhost:" + splitted[len(splitted)-1]}, nil, s.bridge)
	c.Assert(err2, IsNil)
	c.Assert(e2, NotNil)

	c.Check(e2.GetChain(), Equals, common.ETHChain)
	height, err := e2.GetHeight()
	c.Assert(err, IsNil)
	c.Check(height, Equals, int64(1))
	c.Check(e2.CheckIsTestNet(), Equals, true)
	gasPrice, err := e2.GetGasPrice()
	c.Assert(err, IsNil)
	c.Check(gasPrice.Uint64(), Equals, uint64(13))

	acct, err := e2.GetAccount(account)
	c.Assert(err, IsNil)
	c.Check(acct.Sequence, Equals, int64(0))
	c.Check(acct.Coins[0].Amount, Equals, uint64(1000000000))
	pk := types2.GetRandomPubKey()
	addr := e2.GetAddress(pk)
	c.Check(len(addr), Equals, 42)
}
