package binance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	ctypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/cosmos/cosmos-sdk/client/keys"
	cKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/common"

	types2 "gitlab.com/thorchain/thornode/x/thorchain/types"
)

func TestPackage(t *testing.T) { TestingT(t) }

type BinancechainSuite struct {
	thordir  string
	thorKeys *thorclient.Keys
}

var _ = Suite(&BinancechainSuite{})

func (s *BinancechainSuite) SetUpSuite(c *C) {
	var err error
	ns := strconv.Itoa(time.Now().Nanosecond())
	types2.SetupConfigForTest()
	ctypes.Network = ctypes.TestNetwork
	c.Assert(os.Setenv("NET", "testnet"), IsNil)

	s.thordir = filepath.Join(os.TempDir(), ns, ".thorcli")
	cfg := config.ThorchainConfiguration{
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
}

func (s *BinancechainSuite) TearDownSuite(c *C) {
	c.Assert(os.Unsetenv("NET"), IsNil)

	if err := os.RemoveAll(s.thordir); nil != err {
		c.Error(err)
	}
}

const binanceNodeInfo = `{"node_info":{"protocol_version":{"p2p":7,"block":10,"app":0},"id":"7bbe02b44f45fb8f73981c13bb21b19b30e2658d","listen_addr":"10.201.42.4:27146","network":"Binance-Chain-Nile","version":"0.31.5","channels":"3640202122233038","moniker":"Kita","other":{"tx_index":"on","rpc_address":"tcp://0.0.0.0:27147"}},"sync_info":{"latest_block_hash":"BFADEA1DC558D23CB80564AA3C08C863929E4CC93E43C4925D96219114489DC0","latest_app_hash":"1115D879135E2492A947CF3EB9FE055B9813581084EFE3686A6466C2EC12DB7A","latest_block_height":35493230,"latest_block_time":"2019-08-25T00:54:02.906908056Z","catching_up":false},"validator_info":{"address":"E0DD72609CC106210D1AA13936CB67B93A0AEE21","pub_key":[4,34,67,57,104,143,1,46,100,157,228,142,36,24,128,9,46,170,143,106,160,244,241,75,252,249,224,199,105,23,192,182],"voting_power":100000000000}}`

var status = fmt.Sprintf(`{ "jsonrpc": "2.0", "id": "", "result": %s}`, binanceNodeInfo)

func (s *BinancechainSuite) TestNewBinance(c *C) {
	tssCfg := config.TSSConfiguration{
		Scheme: "http",
		Host:   "localhost",
		Port:   0,
	}
	b, err := NewBinance(s.thorKeys, config.BinanceConfiguration{
		RPCHost: "",
	}, false, tssCfg)
	c.Assert(b, IsNil)
	c.Assert(err, NotNil)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		c.Logf("requestUri:%s", req.RequestURI)
		if req.RequestURI == "/status" {
			_, err := rw.Write([]byte(status))
			c.Assert(err, IsNil)
		}
	}))

	b2, err2 := NewBinance(s.thorKeys, config.BinanceConfiguration{
		RPCHost: server.URL,
	}, false, tssCfg)
	c.Assert(err2, IsNil)
	c.Assert(b2, NotNil)
}

const accountInfo string = `{
  "jsonrpc": "2.0",
  "id": "",
  "result": {
    "response": {
      "value": "S9xMJwr/CAoUgT5JOfFWeyGXBP/CrU31i94BCHkSDAoHMDAwLTBFMRCiUhIOCgdBQUEtRUI4EJCFogQSEQoIQUdSSS1CRDIQouubj/8CEg4KCEFMSVMtOTVCEIXFPRIRCgdBTk4tNDU3EICQprf5pQISEgoIQVRPTS0yMEMQgIDpg7HeFhIOCgdBVlQtQjc0EIqg/h4SDQoHQkMxLTNBMRCQv28SDQoDQk5CELLzuMXDvhASEQoHQk5OLTQxMRCAkKa3+aUCEhAKCUJUQy5CLTkxOBDwqf41EhIKCUJUTUdMLUM3MhDxx52H+gUSEQoHQ05OLTIxMBCAkKa3+aUCEhUKCkNPU01PUy01ODcQ8Ybm677a6FgSDwoIQ09USS1EMTMQyK7iBBINCgdEQzEtNEI4EJC/bxIRCghEVUlULTMxQxDU+fGWwwMSDgoHRURVLUREMBCM+9lCEg8KB0ZSSS1ENUYQyaiJ9SkSDgoHSUFBLUM4MRDk18AEEg4KB0lCQi04REUQ5NfABBIOCgdJQ0MtNkVGEOTXwAQSDgoHSURELTUxNhDk18AEEg4KB0lFRS1EQ0EQ5NfABBIOCgdJRkYtODA0EOTXwAQSDgoHSUdHLTAxMxDk18AEEg4KB0lISC1ENEUQ5NfABBIOCgdJSUktMjVDEOTXwAQSDgoHSUpKLTY1RRDk18AEEhIKCktPR0U0OC0zNUQQgMivoCUSDQoHTEMxLTdGQxCQv28SDwoHTENRLUFDNRDO5ZyDIhIQCgdNRkgtOUI1ENb6yYbSJBIKCghOQVNDLTEzNxINCgdOQzEtMjc5EJC/bxINCgdOQzItMjQ5EO6TVhIPCgdPQ0ItQjk1EIDIr6AlEhAKB1BJQy1GNDAQouubj/8CEg4KB1BQQy0wMEEQtLDpYRIRCgdRQlgtQUY1EICi/KevmgESDQoHUkJULUNCNxCFxT0SDQoHUkMxLTk0MxCQv28SDQoHUkMxLUExRRCQv28SDQoHUkMxLUY0ORCQv28SDgoHU1ZDLUExNBCi99oIEg0KB1RDMS1GNDMQkL9vEg8KB1RFRC1ERjIQwP3LzgUSEwoIVEVTVC0wNzUQgICE/qbe4RESEAoIVEVTVC01OTkQgJzNymQSEwoIVEVTVC03OEYQgICE/qbe4RESEwoIVEVTVC1EM0YQgICE/qbe4RESDgoHVEZBLTNCNBD8590CEg8KB1RHVC05RkMQ7KCu73sSDgoHVFNULUQ1NxCAhK9fEg4KB1RTVy02RkQQgMLXLxIPCgdVQ1gtQ0M4EIHPg5sFEg8KB1VETy02MzgQwYbx4xISEwoKVVNEVC5CLUI3QxDsxNuFhQQSEAoJV1dXNzYtQThGEJC+mQISDgoHWFNYLTA3MhC1o/AEEg4KB1lMQy1EOEIQ5aq0ZBIPCgdaQ0ItQjM2EIDkl9ASEg4KCVpFQlJBLTE2RBDoBxIOCgdaWlotMjFFEPTl1QYaJuta6YchAhOb3ZXecsIqwqKw+HhTscyi6K35xYpKaJx10yYwE0QaINLlGCh3"
    }
  }
}`

func (s *BinancechainSuite) TestGetHeight(c *C) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		c.Logf("requestUri:%s", req.RequestURI)
		if req.RequestURI == "/status" {
			_, err := rw.Write([]byte(status))
			c.Assert(err, IsNil)
		} else if req.RequestURI == "/abci_info" {
			_, err := rw.Write([]byte(`{ "jsonrpc": "2.0", "id": "", "result": { "response": { "data": "BNBChain", "last_block_height": "123456789", "last_block_app_hash": "pwx4TJjXu3yaF6dNfLQ9F4nwAhjIqmzE8fNa+RXwAzQ=" } } }`))
			c.Assert(err, IsNil)
		}
	}))

	tssCfg := config.TSSConfiguration{
		Scheme: "http",
		Host:   "localhost",
		Port:   0,
	}
	b, err := NewBinance(s.thorKeys, config.BinanceConfiguration{
		RPCHost: server.URL,
	}, false, tssCfg)
	c.Assert(err, IsNil)

	height, err := b.GetHeight()
	c.Assert(err, IsNil)
	c.Check(height, Equals, int64(123456789))
}

func (s *BinancechainSuite) TestSignTx(c *C) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		c.Logf("requestUri:%s", req.RequestURI)
		if strings.HasPrefix(req.RequestURI, "/abci_query?") {
			if _, err := rw.Write([]byte(accountInfo)); nil != err {
				c.Error(err)
			}
		} else if strings.HasPrefix(req.RequestURI, "/status") {
			if _, err := rw.Write([]byte(status)); nil != err {
				c.Error(err)
			}
		} else {
			// c.Error(fmt.Errorf("no server path"))
		}
	}))
	tssCfg := config.TSSConfiguration{
		Scheme: "http",
		Host:   "localhost",
		Port:   0,
	}
	b2, err2 := NewBinance(s.thorKeys, config.BinanceConfiguration{
		RPCHost: server.URL,
	}, false, tssCfg)
	c.Assert(err2, IsNil)
	c.Assert(b2, NotNil)
	pk, err := common.NewPubKeyFromCrypto(b2.keyManager.GetPrivKey().PubKey())
	c.Assert(err, IsNil)
	txOut := getTxOutFromJsonInput(`{ "height": "1718", "hash": "", "tx_array": [ { "vault_pubkey":"thorpub1addwnpepq2jgpsw2lalzuk7sgtmyakj7l6890f5cfpwjyfp8k4y4t7cw2vk8vcglsjy","seq_no":"0","to": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj", "coin":  { "denom": "BNB", "amount": "194765912" }  } ]}`, c)
	txOut.TxArray[0].VaultPubKey = pk
	out := txOut.TxArray[0].TxOutItem()
	r, p, err := b2.SignTx(out, 1440)
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)
	c.Assert(p, NotNil)

	err = b2.BroadcastTx(r)
	c.Assert(err, IsNil)
}

func getTxOutFromJsonInput(input string, c *C) types.TxOut {
	var txOut types.TxOut
	err := json.Unmarshal([]byte(input), &txOut)
	c.Check(err, IsNil)
	return txOut
}

func (s *BinancechainSuite) TestBinance_isSignerAddressMatch(c *C) {
	env := os.Getenv("NET")
	if len(env) > 0 {
		c.Assert(os.Setenv("NET", "PROD"), IsNil)
		defer func() {
			c.Assert(os.Setenv("NET", env), IsNil)
		}()
	}

	inputs := []struct {
		poolAddr   common.PubKey
		signerAddr string
		match      bool
	}{
		{
			poolAddr:   common.PubKey("thorpub1addwnpepq2jgpsw2lalzuk7sgtmyakj7l6890f5cfpwjyfp8k4y4t7cw2vk8vcglsjy"),
			signerAddr: "blabab",
			match:      false,
		},
		{
			poolAddr:   common.PubKey("thorpub1addwnpepq2jgpsw2lalzuk7sgtmyakj7l6890f5cfpwjyfp8k4y4t7cw2vk8vcglsjy"),
			signerAddr: "bnb1fds7yhw7qt9rkxw9pn65jyj004x858nymnqwd8",
			match:      true,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		c.Logf("requestUri:%s", req.RequestURI)
		if req.RequestURI == "/status" {
			if _, err := rw.Write([]byte(status)); nil != err {
				c.Error(err)
			}
		}
	}))
	tssCfg := config.TSSConfiguration{
		Scheme: "http",
		Host:   "localhost",
		Port:   0,
	}

	b, err := NewBinance(s.thorKeys, config.BinanceConfiguration{
		RPCHost: server.URL,
	}, false, tssCfg)
	c.Assert(err, IsNil)
	c.Assert(b, NotNil)
	for _, item := range inputs {
		result := b.isSignerAddressMatch(item.poolAddr, item.signerAddr)
		c.Assert(result, Equals, item.match)
	}
}
