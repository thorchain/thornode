package binance

import (
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	btypes "github.com/cbarraford/go-sdk/common/types"
	cKeys "github.com/cosmos/cosmos-sdk/client/keys"
	cryptoKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	ctypes "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
	resty "gopkg.in/resty.v1"

	"gitlab.com/thorchain/bepswap/thornode/config"
	"gitlab.com/thorchain/bepswap/thornode/x/statechain/types"
)

func TestPackage(t *testing.T) { TestingT(t) }

type MockBinanceClient struct {
	balance  btypes.BalanceAccount
	nodeInfo btypes.ResultStatus
}

func (m MockBinanceClient) GetClosedOrders(query *btypes.ClosedOrdersQuery) (*btypes.CloseOrders, error) {
	return nil, nil
}

func (m MockBinanceClient) GetDepth(query *btypes.DepthQuery) (*btypes.MarketDepth, error) {
	return nil, nil
}

func (m MockBinanceClient) GetKlines(query *btypes.KlineQuery) ([]btypes.Kline, error) {
	return nil, nil
}

func (m MockBinanceClient) GetMarkets(query *btypes.MarketsQuery) ([]btypes.TradingPair, error) {
	return nil, nil
}

func (m MockBinanceClient) GetOrder(orderID string) (*btypes.Order, error) {
	return nil, nil
}

func (m MockBinanceClient) GetOpenOrders(query *btypes.OpenOrdersQuery) (*btypes.OpenOrders, error) {
	return nil, nil
}

func (m MockBinanceClient) GetTicker24h(query *btypes.Ticker24hQuery) ([]btypes.Ticker24h, error) {
	return nil, nil
}

func (m MockBinanceClient) GetTrades(query *btypes.TradesQuery) (*btypes.Trades, error) {
	return nil, nil
}

func (m MockBinanceClient) GetTime() (*btypes.Time, error) {
	return nil, nil
}

func (m MockBinanceClient) GetTokens(query *btypes.TokensQuery) ([]btypes.Token, error) {
	return nil, nil
}

func (m MockBinanceClient) GetAccount(string) (*btypes.BalanceAccount, error) {
	return &m.balance, nil
}

func (m MockBinanceClient) GetNodeInfo() (*btypes.ResultStatus, error) {
	return &m.nodeInfo, nil
}

type BinancechainSuite struct {
	cfg config.BinanceConfiguration
	kb  cryptoKeys.Keybase
}

var _ = Suite(&BinancechainSuite{})

func (s *BinancechainSuite) setupStateChainForTest(c *C) {
	var err error
	thorcliDir := filepath.Join(os.TempDir(), ".thorcli")
	s.cfg = config.BinanceConfiguration{
		DEXHost:         "localhost",
		SignerName:      "bob",
		SignerPasswd:    "password",
		ChainHomeFolder: thorcliDir,
	}
	s.kb, err = cKeys.NewKeyBaseFromDir(thorcliDir)
	c.Assert(err, IsNil)
	_, _, err = s.kb.CreateMnemonic(
		s.cfg.SignerName,
		cryptoKeys.English,
		s.cfg.SignerPasswd,
		cryptoKeys.Secp256k1,
	)
	c.Assert(err, IsNil)
}

func (s *BinancechainSuite) SetUpSuite(c *C) {
	trSkipVerify := &http.Transport{
		MaxIdleConnsPerHost: 10,
		TLSClientConfig: &tls.Config{
			MaxVersion:         tls.VersionTLS11,
			InsecureSkipVerify: true,
		},
	}
	resty.DefaultClient.SetTransport(trSkipVerify)
	c.Assert(os.Setenv("NET", "testnet"), IsNil)

	s.setupStateChainForTest(c)
}

func (s *BinancechainSuite) TearDownSuite(c *C) {
	c.Assert(os.Unsetenv("NET"), IsNil)

	if err := os.RemoveAll(s.cfg.ChainHomeFolder); nil != err {
		c.Error(err)
	}
}

const binanceNodeInfo = `{"node_info":{"protocol_version":{"p2p":7,"block":10,"app":0},"id":"7bbe02b44f45fb8f73981c13bb21b19b30e2658d","listen_addr":"10.201.42.4:27146","network":"Binance-Chain-Nile","version":"0.31.5","channels":"3640202122233038","moniker":"Kita","other":{"tx_index":"on","rpc_address":"tcp://0.0.0.0:27147"}},"sync_info":{"latest_block_hash":"BFADEA1DC558D23CB80564AA3C08C863929E4CC93E43C4925D96219114489DC0","latest_app_hash":"1115D879135E2492A947CF3EB9FE055B9813581084EFE3686A6466C2EC12DB7A","latest_block_height":35493230,"latest_block_time":"2019-08-25T00:54:02.906908056Z","catching_up":false},"validator_info":{"address":"E0DD72609CC106210D1AA13936CB67B93A0AEE21","pub_key":[4,34,67,57,104,143,1,46,100,157,228,142,36,24,128,9,46,170,143,106,160,244,241,75,252,249,224,199,105,23,192,182],"voting_power":100000000000}}`

func (s *BinancechainSuite) TestNewBinance(c *C) {
	cfg := s.cfg
	cfg.DEXHost = ""
	b, err := NewBinance(cfg)
	c.Assert(b, IsNil)
	c.Assert(err, NotNil)
	b1, err1 := NewBinance(s.cfg)
	c.Assert(b1, IsNil)
	c.Assert(err1, NotNil)

	server := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		c.Logf("requestUri:%s", req.RequestURI)
		if req.RequestURI == "/api/v1/node-info" {
			if _, err := rw.Write([]byte(binanceNodeInfo)); nil != err {
				c.Error(err)
			}
		}
	}))

	cfg = s.cfg
	cfg.DEXHost = server.Listener.Addr().String()
	b2, err2 := NewBinance(cfg)
	c.Assert(err2, IsNil)
	c.Assert(b2, NotNil)
	b3, err3 := NewBinance(s.cfg)
	c.Assert(b3, IsNil)
	c.Assert(err3, NotNil)
	b4, err4 := NewBinance(s.cfg)
	c.Assert(b4, IsNil)
	c.Assert(err4, NotNil)
}

const accountInfo string = `{
"type": "cosmos-sdk/Account",
"value": {
"address": "",
"coins": [],
"public_key": null,
"account_number": "5",
"sequence": "6"
}
}`

func (s *BinancechainSuite) TestSignTx(c *C) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		c.Logf("requestUri:%s", req.RequestURI)
		switch req.RequestURI {
		case "/api/v1/node-info":
			if _, err := rw.Write([]byte(binanceNodeInfo)); nil != err {
				c.Error(err)
			}
		case "/api/v1/account/tbnb1fds7yhw7qt9rkxw9pn65jyj004x858ny4xf2dk":
			if _, err := rw.Write([]byte(accountInfo)); nil != err {
				c.Error(err)
			}
		case "/api/v1/broadcast?sync=true":
			if _, err := rw.Write([]byte(`[
    {
        "ok":true,
        "hash":"E43FA2330C4317ECC084B0C6044DFE75AAE1FAB8F84A66107809E9739D02F80D",
        "code":0
    }
]`)); nil != err {
				c.Error(err)
			}
		}
	}))
	cfg := s.cfg
	cfg.DEXHost = server.Listener.Addr().String()
	b2, err2 := NewBinance(cfg)
	c.Assert(err2, IsNil)
	b2.queryClient = MockBinanceClient{
		balance: btypes.BalanceAccount{
			Sequence: 12,
			Number:   33,
		},
	}
	c.Assert(b2, NotNil)
	r, p, err := b2.SignTx(getTxOutFromJsonInput(`{ "height": "1440", "hash": "", "tx_array": [ { "pool_address":"4b61e25dde02ca3b19c50cf549124f7d4c7a1e64","to": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj", "coins": null } ]}`, c))
	c.Assert(r, IsNil)
	c.Assert(p, IsNil)
	c.Assert(err, IsNil)

	fmt.Println("Hello")
	addrByte, err := ctypes.GetFromBech32(b2.GetAddress(), "tbnb")
	c.Assert(err, IsNil)
	addr := hex.EncodeToString(addrByte)
	fmt.Println(addr)
	j := fmt.Sprintf(`{ "height": "1718", "hash": "", "tx_array": [ { "pool_address":"%s","to": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj", "coins": [ { "denom": "BNB", "amount": "194765912" } ] } ]}`, addr)
	fmt.Println(j)
	r1, p1, err1 := b2.SignTx(getTxOutFromJsonInput(j, c))
	c.Assert(err1, IsNil)
	c.Assert(r1, NotNil)
	c.Assert(p1, NotNil)
	result, err := b2.BroadcastTx(r1, p1)
	c.Assert(result, NotNil)
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
		poolAddr   string
		signerAddr string
		match      bool
	}{
		{
			poolAddr:   "whatever",
			signerAddr: "blabab",
			match:      false,
		},
		{
			poolAddr:   "fe6431ad7d2e103a953cbfacbe460d6df2f4a7ce",
			signerAddr: "blabab",
			match:      false,
		},
		{
			poolAddr:   "fe6431ad7d2e103a953cbfacbe460d6df2f4a7ce",
			signerAddr: "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6",
			match:      true,
		},
	}
	server := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		c.Logf("requestUri:%s", req.RequestURI)
		if req.RequestURI == "/api/v1/node-info" {
			if _, err := rw.Write([]byte(binanceNodeInfo)); nil != err {
				c.Error(err)
			}
		}
	}))

	cfg := s.cfg
	cfg.DEXHost = server.Listener.Addr().String()
	b, err := NewBinance(cfg)
	c.Assert(err, IsNil)
	c.Assert(b, NotNil)
	for _, item := range inputs {
		result := b.isSignerAddressMatch(item.poolAddr, item.signerAddr)
		c.Assert(result, Equals, item.match)
	}
}
