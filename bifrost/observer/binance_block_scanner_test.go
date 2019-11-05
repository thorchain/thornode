package observer

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"

	btypes "gitlab.com/thorchain/bepswap/thornode/bifrost/binance/types"
	"gitlab.com/thorchain/bepswap/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/bepswap/thornode/bifrost/config"
	"gitlab.com/thorchain/bepswap/thornode/bifrost/metrics"
	"gitlab.com/thorchain/bepswap/thornode/bifrost/statechain/types"
)

func Test(t *testing.T) { TestingT(t) }

type BlockScannerTestSuite struct{}

var _ = Suite(&BlockScannerTestSuite{})

func getConfigForTest(rpcHost string) config.BlockScannerConfiguration {
	return config.BlockScannerConfiguration{
		Scheme:                     "https",
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

func (BlockScannerTestSuite) TestNewBlockScanner(c *C) {
	m, err := metrics.NewMetrics(config.MetricConfiguration{
		Enabled:      false,
		ListenPort:   9000,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	})
	pv := &MockPoolAddressValidator{}
	c.Assert(m, NotNil)
	c.Assert(err, IsNil)
	bs, err := NewBinanceBlockScanner(getConfigForTest(""), blockscanner.NewMockScannerStorage(), true, pv, m)
	c.Assert(err, NotNil)
	c.Assert(bs, IsNil)
	bs, err = NewBinanceBlockScanner(getConfigForTest("127.0.0.1"), blockscanner.NewMockScannerStorage(), true, nil, m)
	c.Assert(err, NotNil)
	c.Assert(bs, IsNil)
	bs, err = NewBinanceBlockScanner(getConfigForTest("127.0.0.1"), nil, true, pv, m)
	c.Assert(err, NotNil)
	c.Assert(bs, IsNil)
	bs, err = NewBinanceBlockScanner(getConfigForTest("127.0.0.1"), blockscanner.NewMockScannerStorage(), true, nil, m)
	c.Assert(err, NotNil)
	c.Assert(bs, IsNil)
	bs, err = NewBinanceBlockScanner(getConfigForTest("127.0.0.1"), blockscanner.NewMockScannerStorage(), true, pv, m)
	c.Assert(err, IsNil)
	c.Assert(bs, NotNil)
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

const (
	blockResult = `{ "jsonrpc": "2.0", "id": "", "result": { "block_meta": { "block_id": { "hash": "D063E5F1562F93D46FD4F01CA24813DD60B919D1C39CC34EF1DBB0EA07D0F7F8", "parts": { "total": "1", "hash": "1D9E042DB7616CCB08AF785134561A9AA3074D6CC45A402DDE81572231FD7C91" } }, "header": { "version": { "block": "10", "app": "0" }, "chain_id": "Binance-Chain-Nile", "height": "10", "time": "2019-08-25T05:11:54.192630044Z", "num_txs": "0", "total_txs": "37507966", "last_block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "last_commit_hash": "03E65648ED376E0FBC5373E94394128AB76928E19A66FD2698D7AC9C8B212D33", "data_hash": "", "validators_hash": "80D9AB0FC10D18CA0E0832D5F4C063C5489EC1443DFB738252D038A82131B27A", "next_validators_hash": "80D9AB0FC10D18CA0E0832D5F4C063C5489EC1443DFB738252D038A82131B27A", "consensus_hash": "294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93", "app_hash": "CB951FFB480BCB8BC6FFB53A0AE4515E45C9873BA5B09B6C1ED59BF8F3D63D11", "last_results_hash": "9C7E9DFB57083B4FA4A9BD57519B6FB7E4B75E7D4CD26648815B6A806215C316", "evidence_hash": "", "proposer_address": "7B343E041CA130000A8BC00C35152BD7E7740037" } }, "block": { "header": { "version": { "block": "10", "app": "0" }, "chain_id": "Binance-Chain-Nile", "height": "10", "time": "2019-08-25T05:11:54.192630044Z", "num_txs": "0", "total_txs": "37507966", "last_block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "last_commit_hash": "03E65648ED376E0FBC5373E94394128AB76928E19A66FD2698D7AC9C8B212D33", "data_hash": "", "validators_hash": "80D9AB0FC10D18CA0E0832D5F4C063C5489EC1443DFB738252D038A82131B27A", "next_validators_hash": "80D9AB0FC10D18CA0E0832D5F4C063C5489EC1443DFB738252D038A82131B27A", "consensus_hash": "294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93", "app_hash": "CB951FFB480BCB8BC6FFB53A0AE4515E45C9873BA5B09B6C1ED59BF8F3D63D11", "last_results_hash": "9C7E9DFB57083B4FA4A9BD57519B6FB7E4B75E7D4CD26648815B6A806215C316", "evidence_hash": "", "proposer_address": "7B343E041CA130000A8BC00C35152BD7E7740037" }, "data": { "txs": null }, "evidence": { "evidence": null }, "last_commit": { "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "precommits": [ { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.201530916Z", "validator_address": "06FD60078EB4C2356137DD50036597DB267CF616", "validator_index": "0", "signature": "xaJAxeJJC+tG4hQOsDQr4uGEw8orINmkWBm6oZ7v92YbzqjTM088P+9o/v+Zg/0L/3tb69YU4QM19eu3OKt8AQ==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.164115501Z", "validator_address": "18E69CC672973992BB5F76D049A5B2C5DDF77436", "validator_index": "1", "signature": "6CWLmG1afETad9ThyFL3UrOx5VCv3a7HGAWMYSvExaJlfW562VjefMlFLqesQYLqgr3BtE4poJ8aFrN/zauvDg==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.164945213Z", "validator_address": "344C39BB8F4512D6CAB1F6AAFAC1811EF9D8AFDF", "validator_index": "2", "signature": "BG6u+vmptI5CXAZ6arH9brXQvtBmcWFUx8c4WzIcrftS+JAK2TuhnpcLNUPl9VNw9LBxatCnX60F7L014pKBBA==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.192630044Z", "validator_address": "37EF19AF29679B368D2B9E9DE3F8769B35786676", "validator_index": "3", "signature": "womUxsg21B/6/lXyweBUv0oz4bP1BHoK9BgtbiXSMKfDpb1iGlkZNmZSITyN03hyXabtjsF2AMjGcIzvW6FqAw==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.235226587Z", "validator_address": "62633D9DB7ED78E951F79913FDC8231AA77EC12B", "validator_index": "4", "signature": "WUCR3OR0d0NN2QlXD8xmdQpZo6vIeSHJUOajIlcj7BWmiqWgBEhrURcOaTDE//Zv99oO11ySDu5vGeEFpxNaCw==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.234781133Z", "validator_address": "7B343E041CA130000A8BC00C35152BD7E7740037", "validator_index": "5", "signature": "ejBogd89wMLUu4wfc24RblmGdZFwTNYlLzcC09tN5+TnrbBjAxeF3NbFd8nAsEtI6IGFngMp+mXdpFa6PNntCA==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.163726597Z", "validator_address": "91844D296BD8E591448EFC65FD6AD51A888D58FA", "validator_index": "6", "signature": "hhDq9bOctfjTScJXOAo+uKOwK/m9mWmykcDsrMPDRJQR5HRSekx8sBi7yvTqwzzePtyxux6NoCG6KKGKVuECAA==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.164570853Z", "validator_address": "B3727172CE6473BC780298A2D66C12F1A14F5B2A", "validator_index": "7", "signature": "0vXT7lOpb1+0/nTHJOLP8USjJl9SG3eGRlxxy2H0fFpPaiCS1cPb8ZyEHmjrZvhwRaNxuvkSFsyC32uuPx7QAw==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.191959099Z", "validator_address": "B6F20C7FAA2B2F6F24518FA02B71CB5F4A09FBA3", "validator_index": "8", "signature": "i3RB4OxsJf+h0nYqXn6xyc17PhN+RD5SSdIfhBGFfWBA2UsoPCCm5MawSSTvgFYDeRvdp5M+09RsSDXh8Dm2Ag==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.237319039Z", "validator_address": "E0DD72609CC106210D1AA13936CB67B93A0AEE21", "validator_index": "9", "signature": "PNdyXGrUK9DtRS3hgCNkFiToGg2QsNrV5Mdakr1/66OVDP6noGz2RaIZY/PHowZlcoWsfPGtSVP5C7U2BRT3Bw==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.234516001Z", "validator_address": "FC3108DC3814888F4187452182BC1BAF83B71BC9", "validator_index": "10", "signature": "yMyPeQTEc7Gu+AZwTROZd8+fMnmSu+MYo9Rf9LBVhtWC2BYGJqfAr3Ctgy9Tn7yngj3jUFPPa5AyPOt3b9bFBA==" } ] } } }}`
)

func (BlockScannerTestSuite) TestSearchTxInABlockFromServer(c *C) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Logf("================>:%s", r.RequestURI)
		switch r.RequestURI {
		case "/block": // trying to get block
			if _, err := w.Write([]byte(blockResult)); nil != err {
				c.Error(err)
			}
		case "/tx_search?page=1&per_page=100&prove=true&query=%22tx.height%3D1%22": // block 1
			if _, err := w.Write([]byte(`{
  "jsonrpc": "2.0",
  "id": "",
  "result": {
    "txs": [],
    "total_count": "0"
  }
}`)); nil != err {
				c.Error(err)
			}
		case "/tx_search?page=1&per_page=100&prove=true&query=%22tx.height%3D2%22": // block 1
			if _, err := w.Write([]byte(`{
  "jsonrpc": "2.0",
  "id": "",
  "result": {
    "txs": [
      {
        "hash": "40C23998DCAF0003D6C4EF04161EAB1BE09DBA83323E9E4FF38AFDF5A1883BAE",
        "height": "35526651",
        "index": 0,
        "tx_result": {
          "data": "eyJvcmRlcl9pZCI6IkU5M0FGQTI1MUY1QTFFRUJCOEUzNzNEREM1MDM5NDMwMjY5NEEyMjAtMTA2NTMxIn0=",
          "log": "Msg 0: ",
          "tags": [
            {
              "key": "YWN0aW9u",
              "value": "b3JkZXJOZXc="
            }
          ]
        },
        "tx": "4QHwYl3uCmfObcBDChTpOvolH1oe67jjc93FA5QwJpSiIBIvRTkzQUZBMjUxRjVBMUVFQkI4RTM3M0REQzUwMzk0MzAyNjk0QTIyMC0xMDY1MzEaC1pDQi1GMDBfQk5CIAIoATD8HjiAqNa5B0ABEnIKJuta6YchAv+Fh/6SCT/nXtoa/c4s1p3cvs0B6EH24jX94NocvGBdEkCXzbtnnErlbokeOUQlZyb0kvegAWpm/Zc12dgdz7Qa/T39gYFOzZF2bLEanu1TZagIuK53qEvN96xO0Fs1U+btGPanKiCiwAY=",
        "proof": {
          "RootHash": "07CB13893EE47F9BAC752B6F0353D9D4BE244E4403D4C1D67D828282B08E4296",
          "Data": "4QHwYl3uCmfObcBDChTpOvolH1oe67jjc93FA5QwJpSiIBIvRTkzQUZBMjUxRjVBMUVFQkI4RTM3M0REQzUwMzk0MzAyNjk0QTIyMC0xMDY1MzEaC1pDQi1GMDBfQk5CIAIoATD8HjiAqNa5B0ABEnIKJuta6YchAv+Fh/6SCT/nXtoa/c4s1p3cvs0B6EH24jX94NocvGBdEkCXzbtnnErlbokeOUQlZyb0kvegAWpm/Zc12dgdz7Qa/T39gYFOzZF2bLEanu1TZagIuK53qEvN96xO0Fs1U+btGPanKiCiwAY=",
          "Proof": {
            "total": "1",
            "index": "0",
            "leaf_hash": "B8sTiT7kf5usdStvA1PZ1L4kTkQD1MHWfYKCgrCOQpY=",
            "aunts": []
          }
        }
      }
    ],
    "total_count": "1"
  }
}`)); nil != err {
				c.Error(err)
			}
		case "/api/v1/tx/40C23998DCAF0003D6C4EF04161EAB1BE09DBA83323E9E4FF38AFDF5A1883BAE?format=json": // return a tx
			if _, err := w.Write([]byte(normalApiTx)); nil != err {
				c.Error(err)
			}
		default:
			if strings.Contains(r.RequestURI, "tx_search?page=1&per_page=100&prove=true") {
				if _, err := w.Write([]byte(`{
  "jsonrpc": "2.0",
  "id": "",
  "result": {
    "txs": [],
    "total_count": "0"
  }
}`)); nil != err {
					c.Error(err)
				}
			}
		}
	})
	s := httptest.NewTLSServer(h)
	defer s.Close()
	m, err := metrics.NewMetrics(config.MetricConfiguration{
		Enabled:      false,
		ListenPort:   9000,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	})
	c.Assert(m, NotNil)
	c.Assert(err, IsNil)
	pv := &MockPoolAddressValidator{}
	bs, err := NewBinanceBlockScanner(getConfigForTest(s.Listener.Addr().String()), blockscanner.NewMockScannerStorage(), true, pv, m)
	c.Assert(err, IsNil)
	c.Assert(bs, NotNil)
	trSkipVerify := &http.Transport{
		MaxIdleConnsPerHost: 10,
		TLSClientConfig: &tls.Config{
			MaxVersion:         tls.VersionTLS11,
			InsecureSkipVerify: true,
		},
	}
	bs.commonBlockScanner.GetHttpClient().Transport = trSkipVerify
	bs.Start()
	// read all the messages
	go func() {
		for item := range bs.GetMessages() {
			c.Logf("got message on block height:%s", item.BlockHeight)
		}
	}()
	// stop
	time.Sleep(time.Second * 5)
	err = bs.Stop()
	c.Assert(err, IsNil)
}

func (BlockScannerTestSuite) TestFromTxToTxIn(c *C) {
	testFunc := func(input string, txInItemCheck, errCheck Checker) *types.TxInItem {
		var query btypes.RPCTxSearch
		err := json.Unmarshal([]byte(input), &query)
		c.Check(err, IsNil)
		c.Check(query.Result.Txs, NotNil)
		m, err := metrics.NewMetrics(config.MetricConfiguration{
			Enabled:      false,
			ListenPort:   9000,
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
		})
		c.Assert(m, NotNil)
		c.Assert(err, IsNil)
		pv := NewMockPoolAddressValidator()
		bs, err := NewBinanceBlockScanner(getConfigForTest("127.0.0.1"), blockscanner.NewMockScannerStorage(), true, pv, m)
		c.Assert(err, IsNil)
		c.Assert(bs, NotNil)
		for _, item := range query.Result.Txs {
			txInItem, err := bs.fromTxToTxIn(item.Hash, item.Height, item.Tx)
			c.Logf("hash:%s", item.Hash)
			c.Check(txInItem, txInItemCheck)
			c.Check(err, errCheck)
			if nil != txInItem {
				return txInItem
			}
		}
		return nil
	}
	// NewOrder Transaction on binance chain, we don't care about it
	testFunc(binanceTxNewOrder, IsNil, IsNil)

	// Normal tx send to our pool , withdraw
	testFunc(binanceTxTransferWithdraw, NotNil, IsNil)
	// normal tx outbound from our pool
	testFunc(binanceTxOutboundFromPool, NotNil, IsNil)
	txInItem := testFunc(binanceTxOutboundFromPool1, NotNil, IsNil)
	c.Check(txInItem, NotNil)
	c.Check(txInItem.Memo, Equals, "OUTBOUND:825")
	c.Check(txInItem.Sender, Equals, "tbnb1yycn4mh6ffwpjf584t8lpp7c27ghu03gpvqkfj")
	c.Check(len(txInItem.Coins), Equals, 1)
	c.Check(txInItem.Coins[0].Asset.String(), Equals, common.RuneAsset().String())
	c.Check(txInItem.Coins[0].Amount.Uint64(), Equals, uint64(common.One/10))
	txInItem1 := testFunc(binanceTxSwapLOKToBNB, NotNil, IsNil)
	c.Check(txInItem1, NotNil)
	c.Check(txInItem1.Memo, Equals, "SWAP:BNB")
	c.Check(txInItem1.Sender, Equals, "tbnb190tgp5uchnlcpsk7n7nffypkwlzhcqge27xkfh")
	// observed pool address should be the hex encoded pubkey
	c.Check(txInItem1.ObservedPoolAddress, Equals, "b89c9b697180249e0be37f065fc8aa7a211f2105")
	c.Check(len(txInItem1.Coins), Equals, 1)
	c.Check(txInItem1.Coins[0].Asset.String(), Equals, "BNB.LOK-3C0")
	c.Check(txInItem1.Coins[0].Amount.Uint64(), Equals, uint64(common.One))
}
