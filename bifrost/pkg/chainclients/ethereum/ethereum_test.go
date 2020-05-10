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
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
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

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		c.Logf("requestUri:%s", req.RequestURI)
		if strings.HasPrefix(req.RequestURI, "/thorchain/keysign") {
			_, err := rw.Write([]byte(`{
			"chains": {
				"ETH": {
					"chain": "ETH",
					"hash": "",
					"height": "1",
					"tx_array": [
						{
							"chain": "ETH",
							"coin": {
								"amount": "194765912",
								"asset": "ETH.ETH"
							},
							"in_hash": "",
							"memo": "",
							"out_hash": "",
							"to": "0xde0b295669a9fd93d5f28d9ec85e40f4cb697bae",
							"vault_pubkey": "thorpub1addwnpepqflvfv08t6qt95lmttd6wpf3ss8wx63e9vf6fvyuj2yy6nnyna5763e2kck"
						}]
					}
				}
			}
			`))
			c.Assert(err, IsNil)
		} else if strings.HasSuffix(req.RequestURI, "/signers") {
			_, err := rw.Write([]byte(`[
				"thorpub1addwnpepqflvfv08t6qt95lmttd6wpf3ss8wx63e9vf6fvyuj2yy6nnyna5763e2kck",
				"thorpub1addwnpepq2flfr96skc5lkwdv0n5xjsnhmuju20x3zndgu42zd8dtkrud9m2v0zl2qu",
				"thorpub1addwnpepqwhnus6xs4208d4ynm05lv493amz3fexfjfx4vptntedd7k0ajlcup0pzgk"
			]`))
			c.Assert(err, IsNil)
		}
	}))
	splitted := strings.SplitAfter(server.URL, ":")

	cfg := config.ClientConfiguration{
		ChainID:         "thorchain",
		ChainHost:       "localhost:" + splitted[len(splitted)-1],
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

func (s *EthereumSuite) TestClient(c *C) {
	e, err := NewClient(s.thorKeys, config.ChainConfiguration{}, nil, s.bridge, s.m)
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
		if rpcRequest.Method == "eth_sendRawTransaction" {
			_, err := rw.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x88df016429689c079f3b2f6ad39fa052532c56795b733da78a91ebe6a713944b"}`))
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
	e2, err2 := NewClient(s.thorKeys, config.ChainConfiguration{
		RPCHost: "http://localhost:" + splitted[len(splitted)-1],
		BlockScanner: config.BlockScannerConfiguration{
			StartBlockHeight: 1, // avoids querying thorchain for block height
		},
	}, nil, s.bridge, s.m)
	c.Assert(err2, IsNil)
	c.Assert(e2, NotNil)

	c.Check(e2.GetChain(), Equals, common.ETHChain)
	height, err := e2.GetHeight()
	c.Assert(err, IsNil)
	c.Check(height, Equals, int64(1))
	gasPrice, err := e2.GetGasPrice()
	c.Assert(err, IsNil)
	c.Check(gasPrice.Uint64(), Equals, uint64(13))

	acct, err := e2.GetAccount(types2.GetRandomPubKey())
	c.Assert(err, IsNil)
	c.Check(acct.Sequence, Equals, int64(0))
	c.Check(acct.Coins[0].Amount, Equals, uint64(1000000000))
	pk := types2.GetRandomPubKey()
	addr := e2.GetAddress(pk)
	c.Check(len(addr), Equals, 42)
	err = e2.BroadcastTx(stypes.TxOutItem{}, []byte(`{
		"from":"0xa7d9ddbe1f17865597fbd27ec712455208b6b76d",
		"gas":"0xc350",
		"gasPrice":"0x4a817c800",
		"input":"0x68656c6c6f21",
		"nonce":"0x15",
		"to":"0xf02c1c8e6114b1dbe8937a39260b5b0a374432bb",
		"transactionIndex":"0x41",
		"value":"0xf3dbb76162000",
		"v":"0x25",
		"r":"0x1b5e176d927f8e9ab405058b2d2457392da3e20f328b16ddabcebc33eaac5fea",
		"s":"0x4ba69724e8f69de52f0125ad8b3c5c2cef33019bac3249e2c0a2192766d1721c"
	}`))
	c.Assert(err, IsNil)

	input := []byte(`{ "height": "1", "hash": "", "tx_array": [ { "vault_pubkey":"thorpub1addwnpepq2jgpsw2lalzuk7sgtmyakj7l6890f5cfpwjyfp8k4y4t7cw2vk8vcglsjy","seq_no":"0","to":"0xde0b295669a9fd93d5f28d9ec85e40f4cb697bae", "coin": { "asset": "ETH", "amount": "194765912" }  } ]}`)
	var txOut stypes.TxOut
	err = json.Unmarshal(input, &txOut)
	c.Check(err, IsNil)

	txOut.TxArray[0].VaultPubKey = e2.kw.GetPubKey()
	c.Logf(txOut.TxArray[0].VaultPubKey.String())
	c.Logf(e2.kw.GetPubKey().String())
	out := txOut.TxArray[0].TxOutItem()

	r, err := e2.SignTx(out, 1)
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	err = e2.BroadcastTx(out, r)
	c.Assert(err, IsNil)
	meta := e2.accts.Get(out.VaultPubKey)
	addr = e2.GetAddress(out.VaultPubKey)
	c.Assert(err, IsNil)
	c.Check(meta.Address, Equals, addr)
	c.Check(meta.Nonce, Equals, uint64(1))
}
