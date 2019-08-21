package observer

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/observe/config"
	btypes "gitlab.com/thorchain/bepswap/observe/x/binance/types"
	"gitlab.com/thorchain/bepswap/observe/x/blockscanner"
)

func Test(t *testing.T) { TestingT(t) }

type BlockScannerTestSuite struct{}

var _ = Suite(&BlockScannerTestSuite{})

func getConfigForTest(rpcHost string) config.BlockScannerConfiguration {
	return config.BlockScannerConfiguration{
		RPCHost:                    rpcHost,
		StartBlockHeight:           0,
		BlockScanProcessors:        1,
		HttpRequestTimeout:         time.Second,
		HttpRequestReadTimeout:     time.Second * 30,
		HttpRequestWriteTimeout:    time.Second * 30,
		MaxHttpRequestRetry:        10,
		BlockHeightDiscoverBackoff: time.Second,
		BlockRetryInterval:         time.Second,
	}
}

func (BlockScannerTestSuite) TestNewBlockScanner(c *C) {
	bs, err := NewBinanceBlockScanner(getConfigForTest(""), blockscanner.NewMockScannerStorage(), "", common.BnbAddress(""))
	c.Assert(err, NotNil)
	c.Assert(bs, IsNil)
	bs, err = NewBinanceBlockScanner(getConfigForTest("127.0.0.1"), blockscanner.NewMockScannerStorage(), "", common.BnbAddress(""))
	c.Assert(err, NotNil)
	c.Assert(bs, IsNil)
	bs, err = NewBinanceBlockScanner(getConfigForTest("127.0.0.1"), nil, "127.0.0.1", common.BnbAddress(""))
	c.Assert(err, NotNil)
	c.Assert(bs, IsNil)
	bs, err = NewBinanceBlockScanner(getConfigForTest("127.0.0.1"), blockscanner.NewMockScannerStorage(), "127.0.0.1", common.BnbAddress(""))
	c.Assert(err, NotNil)
	c.Assert(bs, IsNil)
	bs, err = NewBinanceBlockScanner(getConfigForTest("127.0.0.1"), blockscanner.NewMockScannerStorage(), "127.0.0.1", common.BnbAddress("tbnb1ggdcyhk8rc7fgzp8wa2su220aclcggcsd94ye5"))
	c.Assert(err, IsNil)
	c.Assert(bs, NotNil)
}

func (BlockScannerTestSuite) TestFromApiTxToTxInItem(c *C) {
	input := `{"code":0,"hash":"22214C3567DCF0120DA779CC24089C8D18F9B8F217A5B1AD4821EFFFDD2BF92F","height":"34117745","log":"Msg 0: ","ok":true,"tx":{"type":"auth/StdTx","value":{"data":null,"memo":"withdraw:BNB","msg":[{"type":"cosmos-sdk/Send","value":{"inputs":[{"address":"tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj","coins":[{"amount":"100000000","denom":"BNB"}]}],"outputs":[{"address":"tbnb1ggdcyhk8rc7fgzp8wa2su220aclcggcsd94ye5","coins":[{"amount":"100000000","denom":"BNB"}]}]}}],"signatures":[{"account_number":"690006","pub_key":{"type":"tendermint/PubKeySecp256k1","value":"A1MJdnZOD5ji4LbNJtQciz6+HQQjzq0ETiC9mHkVlyXx"},"sequence":"9","signature":"579wbq2otxtKEjd3eVy514LIpeByCg4ak1IF5sHyK/xbfa6WhAswXU+xfntA46sJMc8p2fENOyG6dMrotZjggg=="}],"source":"1"}}}`
	var apiTx btypes.ApiTx
	err := json.Unmarshal([]byte(input), &apiTx)
	c.Assert(err, IsNil)
	bs, err := NewBinanceBlockScanner(getConfigForTest("127.0.0.1"), blockscanner.NewMockScannerStorage(), "127.0.0.1", common.BnbAddress("tbnb1ggdcyhk8rc7fgzp8wa2su220aclcggcsd94ye5"))
	c.Assert(err, IsNil)
	c.Assert(bs, NotNil)
	txIn, err := bs.fromApiTxToTxInItem(apiTx)
	c.Assert(err, IsNil)
	c.Assert(txIn, NotNil)
}

func (BlockScannerTestSuite) TestFromApiTxToTxInItemNotExist(c *C) {
	input := `{"code":0,"hash":"22214C3567DCF0120DA779CC24089C8D18F9B8F217A5B1AD4821EFFFDD2BF92F","height":"34117745","log":"Msg 0: ","ok":true,"tx":{"type":"auth/StdTx","value":{"data":null,"memo":"withdraw:BNB","msg":[{"type":"cosmos-sdk/Send","value":{"inputs":[{"address":"tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj","coins":[{"amount":"100000000","denom":"BNB"}]}],"outputs":[{"address":"tbnb1ggdcyhk8rc7fgzp8wa2su220aclcggcsd94yex","coins":[{"amount":"100000000","denom":"BNB"}]}]}}],"signatures":[{"account_number":"690006","pub_key":{"type":"tendermint/PubKeySecp256k1","value":"A1MJdnZOD5ji4LbNJtQciz6+HQQjzq0ETiC9mHkVlyXx"},"sequence":"9","signature":"579wbq2otxtKEjd3eVy514LIpeByCg4ak1IF5sHyK/xbfa6WhAswXU+xfntA46sJMc8p2fENOyG6dMrotZjggg=="}],"source":"1"}}}`
	var apiTx btypes.ApiTx
	err := json.Unmarshal([]byte(input), &apiTx)
	c.Assert(err, IsNil)
	bs, err := NewBinanceBlockScanner(getConfigForTest("127.0.0.1"), blockscanner.NewMockScannerStorage(), "127.0.0.1", common.BnbAddress("tbnb1ggdcyhk8rc7fgzp8wa2su220aclcggcsd94ye5"))
	c.Assert(err, IsNil)
	c.Assert(bs, NotNil)
	txIn, err := bs.fromApiTxToTxInItem(apiTx)
	c.Assert(err, IsNil)
	c.Assert(txIn, IsNil)
	input1 := `{"code":0,"hash":"22214C3567DCF0120DA779CC24089C8D18F9B8F217A5B1AD4821EFFFDD2BF92F","height":"34117745","log":"Msg 0: ","ok":true,"tx":{"type":"auth/StdTx","value":{"data":null,"memo":"withdraw:BNB","msg":[{"type":"cosmos-sdk/Send","value":{"inputs":[{"address":"tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj","coins":[{"amount":"100000000","denom":"ABCDEFGHIJK"}]}],"outputs":[{"address":"tbnb1ggdcyhk8rc7fgzp8wa2su220aclcggcsd94ye5","coins":[{"amount":"100000000","denom":"ABCDEFGHG"}]}]}}],"signatures":[{"account_number":"690006","pub_key":{"type":"tendermint/PubKeySecp256k1","value":"A1MJdnZOD5ji4LbNJtQciz6+HQQjzq0ETiC9mHkVlyXx"},"sequence":"9","signature":"579wbq2otxtKEjd3eVy514LIpeByCg4ak1IF5sHyK/xbfa6WhAswXU+xfntA46sJMc8p2fENOyG6dMrotZjggg=="}],"source":"1"}}}`
	var apiTx1 btypes.ApiTx
	err1 := json.Unmarshal([]byte(input1), &apiTx1)
	c.Assert(err1, IsNil)
	txIn1, err2 := bs.fromApiTxToTxInItem(apiTx1)
	c.Assert(err2, NotNil)
	c.Assert(txIn1, IsNil)

	input2 := `{"code":0,"hash":"22214C3567DCF0120DA779CC24089C8D18F9B8F217A5B1AD4821EFFFDD2BF92F","height":"34117745","log":"Msg 0: ","ok":true,"tx":{"type":"auth/StdTx","value":{"data":null,"memo":"withdraw:BNB","msg":[{"type":"cosmos-sdk/Send","value":{"inputs":[{"address":"tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj","coins":[{"amount":"eeee","denom":"BNB"}]}],"outputs":[{"address":"tbnb1ggdcyhk8rc7fgzp8wa2su220aclcggcsd94ye5","coins":[{"amount":"xdcf","denom":"BNB"}]}]}}],"signatures":[{"account_number":"690006","pub_key":{"type":"tendermint/PubKeySecp256k1","value":"A1MJdnZOD5ji4LbNJtQciz6+HQQjzq0ETiC9mHkVlyXx"},"sequence":"9","signature":"579wbq2otxtKEjd3eVy514LIpeByCg4ak1IF5sHyK/xbfa6WhAswXU+xfntA46sJMc8p2fENOyG6dMrotZjggg=="}],"source":"1"}}}`
	var apiTx2 btypes.ApiTx
	err3 := json.Unmarshal([]byte(input2), &apiTx2)
	c.Assert(err3, IsNil)
	txIn2, err4 := bs.fromApiTxToTxInItem(apiTx2)
	c.Assert(err4, NotNil)
	c.Assert(txIn2, IsNil)
}

const (
	normalApiTx = `{
    "code": 0,
    "hash": "22214C3567DCF0120DA779CC24089C8D18F9B8F217A5B1AD4821EFFFDD2BF92F",
    "height": "34117745",
    "log": "Msg 0: ",
    "ok": true,
    "tx": {
        "type": "auth/StdTx",
        "value": {
            "data": null,
            "memo": "withdraw:BNB",
            "msg": [
                {
                    "type": "cosmos-sdk/Send",
                    "value": {
                        "inputs": [
                            {
                                "address": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj",
                                "coins": [
                                    {
                                        "amount": "100000000",
                                        "denom": "BNB"
                                    }
                                ]
                            }
                        ],
                        "outputs": [
                            {
                                "address": "tbnb1ggdcyhk8rc7fgzp8wa2su220aclcggcsd94ye5",
                                "coins": [
                                    {
                                        "amount": "100000000",
                                        "denom": "BNB"
                                    }
                                ]
                            }
                        ]
                    }
                }
            ],
            "signatures": [
                {
                    "account_number": "690006",
                    "pub_key": {
                        "type": "tendermint/PubKeySecp256k1",
                        "value": "A1MJdnZOD5ji4LbNJtQciz6+HQQjzq0ETiC9mHkVlyXx"
                    },
                    "sequence": "9",
                    "signature": "579wbq2otxtKEjd3eVy514LIpeByCg4ak1IF5sHyK/xbfa6WhAswXU+xfntA46sJMc8p2fENOyG6dMrotZjggg=="
                }
            ],
            "source": "1"
        }
    }
}`
)

func (BlockScannerTestSuite) TestGetOneTxFromServer(c *C) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(normalApiTx)); nil != err {
			c.Error(err)
		}
	})
	s := httptest.NewTLSServer(h)
	defer s.Close()
	addr := s.Listener.Addr().String()
	bs, err := NewBinanceBlockScanner(getConfigForTest(addr), blockscanner.NewMockScannerStorage(), addr, common.BnbAddress("tbnb1ggdcyhk8rc7fgzp8wa2su220aclcggcsd94ye5"))
	c.Assert(err, IsNil)
	c.Assert(bs, NotNil)
	singleTxUrl := bs.getSingleTxUrl("22214C3567DCF0120DA779CC24089C8D18F9B8F217A5B1AD4821EFFFDD2BF92F")
	bs.commonBlockScanner.GetHttpClient().TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	inItem, err := bs.getOneTxFromServer("22214C3567DCF0120DA779CC24089C8D18F9B8F217A5B1AD4821EFFFDD2BF92F", singleTxUrl)
	c.Assert(err, IsNil)
	c.Assert(inItem, NotNil)
}
