package binance

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	. "gopkg.in/check.v1"
	resty "gopkg.in/resty.v1"

	"gitlab.com/thorchain/bepswap/thornode/bifrost/config"
	"gitlab.com/thorchain/bepswap/thornode/bifrost/thorclient/types"

	types2 "gitlab.com/thorchain/bepswap/thornode/x/thorchain/types"
)

func TestPackage(t *testing.T) { TestingT(t) }

type BinancechainSuite struct{}

var _ = Suite(&BinancechainSuite{})

func (*BinancechainSuite) SetUpSuite(c *C) {
	types2.SetupConfigForTest()
	trSkipVerify := &http.Transport{
		MaxIdleConnsPerHost: 10,
		TLSClientConfig: &tls.Config{
			MaxVersion:         tls.VersionTLS11,
			InsecureSkipVerify: true,
		},
	}
	resty.DefaultClient.SetTransport(trSkipVerify)
	c.Assert(os.Setenv("NET", "testnet"), IsNil)
}

func (*BinancechainSuite) TearDownSuite(c *C) {
	c.Assert(os.Unsetenv("NET"), IsNil)
}

const binanceNodeInfo = `{"node_info":{"protocol_version":{"p2p":7,"block":10,"app":0},"id":"7bbe02b44f45fb8f73981c13bb21b19b30e2658d","listen_addr":"10.201.42.4:27146","network":"Binance-Chain-Nile","version":"0.31.5","channels":"3640202122233038","moniker":"Kita","other":{"tx_index":"on","rpc_address":"tcp://0.0.0.0:27147"}},"sync_info":{"latest_block_hash":"BFADEA1DC558D23CB80564AA3C08C863929E4CC93E43C4925D96219114489DC0","latest_app_hash":"1115D879135E2492A947CF3EB9FE055B9813581084EFE3686A6466C2EC12DB7A","latest_block_height":35493230,"latest_block_time":"2019-08-25T00:54:02.906908056Z","catching_up":false},"validator_info":{"address":"E0DD72609CC106210D1AA13936CB67B93A0AEE21","pub_key":[4,34,67,57,104,143,1,46,100,157,228,142,36,24,128,9,46,170,143,106,160,244,241,75,252,249,224,199,105,23,192,182],"voting_power":100000000000}}`

var status = fmt.Sprintf(`{ "jsonrpc": "2.0", "id": "", "result": %s}`, binanceNodeInfo)

func (BinancechainSuite) TestNewBinance(c *C) {
	tssCfg := config.TSSConfiguration{
		Scheme: "http",
		Host:   "localhost",
		Port:   0,
	}
	b, err := NewBinance(config.BinanceConfiguration{
		RPCHost:    "",
		PrivateKey: "91a2f0e5b1495cf51b0792a009b49c54ce8ae52d0dada711e73d98b22e6698ea",
	}, false, tssCfg)
	c.Assert(b, IsNil)
	c.Assert(err, NotNil)
	b1, err1 := NewBinance(config.BinanceConfiguration{
		RPCHost:    "http://localhost",
		PrivateKey: "",
	}, false, tssCfg)
	c.Assert(b1, IsNil)
	c.Assert(err1, NotNil)

	server := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		c.Logf("requestUri:%s", req.RequestURI)
		if req.RequestURI == "/api/v1/node-info" {
			_, err := rw.Write([]byte(binanceNodeInfo))
			c.Assert(err, IsNil)
		} else if req.RequestURI == "/status" {
			_, err := rw.Write([]byte(status))
			c.Assert(err, IsNil)
		}
	}))

	b2, err2 := NewBinance(config.BinanceConfiguration{
		RPCHost:    server.URL,
		PrivateKey: "91a2f0e5b1495cf51b0792a009b49c54ce8ae52d0dada711e73d98b22e6698ea",
	}, false, tssCfg)
	c.Assert(err2, IsNil)
	c.Assert(b2, NotNil)
	return
	b3, err3 := NewBinance(config.BinanceConfiguration{
		RPCHost:    "http://localhost",
		PrivateKey: "asdfsdfdsf",
	}, false, tssCfg)
	c.Assert(b3, IsNil)
	c.Assert(err3, NotNil)
	b4, err4 := NewBinance(config.BinanceConfiguration{
		RPCHost:    "http://localhost",
		PrivateKey: "91a2f0e5b1495cf51b0792a009b49c54ce8ae52d0dada711e73d98b22e6698ea",
	}, false, tssCfg)
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

func (BinancechainSuite) TestSignTx(c *C) {
	c.Skip("skipping")
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
		case "status":
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
	b2, err2 := NewBinance(config.BinanceConfiguration{
		RPCHost:    server.Listener.Addr().String(),
		PrivateKey: "91a2f0e5b1495cf51b0792a009b49c54ce8ae52d0dada711e73d98b22e6698ea",
	}, false, tssCfg)
	c.Assert(err2, IsNil)
	c.Assert(b2, NotNil)
	txOut := getTxOutFromJsonInput(`{ "height": "1440", "hash": "", "tx_array": [ { "pool_address":"thorpub1addwnpepqd5r97je7uw94e3t27r2jhxdxuglp5q5dr2muhckcpek96365dutx8frl9w","seq_no":"0","to": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj", "coins": null } ]}`, c)
	r, p, err := b2.SignTx(txOut.TxArray[0], 1440)
	c.Assert(r, IsNil)
	c.Assert(p, IsNil)
	c.Assert(err, IsNil)
	txOut1 := getTxOutFromJsonInput(`{ "height": "1718", "hash": "", "tx_array": [ { "pool_address":"thorpub1addwnpepq2jgpsw2lalzuk7sgtmyakj7l6890f5cfpwjyfp8k4y4t7cw2vk8vcglsjy","seq_no":"0","to": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj", "coin":  { "denom": "BNB", "amount": "194765912" }  } ]}`, c)
	r1, p1, err1 := b2.SignTx(txOut1.TxArray[0], 1718)
	c.Assert(r1, NotNil)
	c.Assert(p1, NotNil)
	c.Assert(err1, IsNil)
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

func (BinancechainSuite) TestBinance_isSignerAddressMatch(c *C) {
	c.Skip("skipping")
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
			poolAddr:   "thorpub1addwnpepq2jgpsw2lalzuk7sgtmyakj7l6890f5cfpwjyfp8k4y4t7cw2vk8vcglsjy",
			signerAddr: "blabab",
			match:      false,
		},
		{
			poolAddr:   "thorpub1addwnpepq2jgpsw2lalzuk7sgtmyakj7l6890f5cfpwjyfp8k4y4t7cw2vk8vcglsjy",
			signerAddr: "bnb1fds7yhw7qt9rkxw9pn65jyj004x858nymnqwd8",
			match:      true,
		},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		c.Logf("requestUri:%s", req.RequestURI)
		if req.RequestURI == "/api/v1/node-info" {
			if _, err := rw.Write([]byte(binanceNodeInfo)); nil != err {
				c.Error(err)
			} else if req.RequestURI == "/status" {
				if _, err := rw.Write([]byte(status)); nil != err {
					c.Error(err)
				}
			}
		}
	}))
	tssCfg := config.TSSConfiguration{
		Scheme: "https",
		Host:   "127.0.0.1",
		Port:   0,
	}

	b, err := NewBinance(config.BinanceConfiguration{
		RPCHost:    server.URL,
		PrivateKey: "91a2f0e5b1495cf51b0792a009b49c54ce8ae52d0dada711e73d98b22e6698ea",
	}, false, tssCfg)
	c.Assert(err, IsNil)
	c.Assert(b, NotNil)
	for _, item := range inputs {
		result := b.isSignerAddressMatch(item.poolAddr, item.signerAddr)
		c.Assert(result, Equals, item.match)
	}
}
