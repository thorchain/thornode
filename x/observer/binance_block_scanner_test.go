package observer

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/observe/config"
	btypes "gitlab.com/thorchain/bepswap/observe/x/binance/types"
	"gitlab.com/thorchain/bepswap/observe/x/blockscanner"
	"gitlab.com/thorchain/bepswap/observe/x/metrics"
	"gitlab.com/thorchain/bepswap/observe/x/statechain/types"
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
	c.Assert(m, NotNil)
	c.Assert(err, IsNil)
	bs, err := NewBinanceBlockScanner(getConfigForTest(""), blockscanner.NewMockScannerStorage(), true, common.BnbAddress(""), m)
	c.Assert(err, NotNil)
	c.Assert(bs, IsNil)
	bs, err = NewBinanceBlockScanner(getConfigForTest("127.0.0.1"), blockscanner.NewMockScannerStorage(), true, common.BnbAddress(""), m)
	c.Assert(err, NotNil)
	c.Assert(bs, IsNil)
	bs, err = NewBinanceBlockScanner(getConfigForTest("127.0.0.1"), nil, true, common.BnbAddress(""), m)
	c.Assert(err, NotNil)
	c.Assert(bs, IsNil)
	bs, err = NewBinanceBlockScanner(getConfigForTest("127.0.0.1"), blockscanner.NewMockScannerStorage(), true, common.BnbAddress(""), m)
	c.Assert(err, NotNil)
	c.Assert(bs, IsNil)
	bs, err = NewBinanceBlockScanner(getConfigForTest("127.0.0.1"), blockscanner.NewMockScannerStorage(), true, common.BnbAddress("tbnb1ggdcyhk8rc7fgzp8wa2su220aclcggcsd94ye5"), m)
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
	bs, err := NewBinanceBlockScanner(getConfigForTest(s.Listener.Addr().String()), blockscanner.NewMockScannerStorage(), true, common.BnbAddress("tbnb1ggdcyhk8rc7fgzp8wa2su220aclcggcsd94ye5"), m)
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
		bs, err := NewBinanceBlockScanner(getConfigForTest("127.0.0.1"), blockscanner.NewMockScannerStorage(), true, common.BnbAddress("tbnb1ggdcyhk8rc7fgzp8wa2su220aclcggcsd94ye5"), m)
		c.Assert(err, IsNil)
		c.Assert(bs, NotNil)
		for _, item := range query.Result.Txs {
			txInItem, err := bs.fromTxToTxIn(item.Hash, item.Height, item.Tx)
			c.Check(txInItem, txInItemCheck)
			c.Check(err, errCheck)
			if nil != txInItem {
				return txInItem
			}
		}
		return nil
	}
	testFunc(`
{
  "jsonrpc": "2.0",
  "id": "",
  "result": {
    "txs": [
      {
        "hash": "A8FCD14430FDA557C5744ECC18AA9C9704B739E31FA6FA8328FDD8206F2F47EF",
        "height": "35559022",
        "index": 0,
        "tx_result": {
          "data": "eyJvcmRlcl9pZCI6IkU5M0FGQTI1MUY1QTFFRUJCOEUzNzNEREM1MDM5NDMwMjY5NEEyMjAtMTA5MzU1In0=",
          "log": "Msg 0: ",
          "tags": [
            {
              "key": "YWN0aW9u",
              "value": "b3JkZXJOZXc="
            }
          ]
        },
        "tx": "4QHwYl3uCmfObcBDChTpOvolH1oe67jjc93FA5QwJpSiIBIvRTkzQUZBMjUxRjVBMUVFQkI4RTM3M0REQzUwMzk0MzAyNjk0QTIyMC0xMDkzNTUaC1pDQi1GMDBfQk5CIAIoATCBRTiAqNa5B0ABEnIKJuta6YchAv+Fh/6SCT/nXtoa/c4s1p3cvs0B6EH24jX94NocvGBdEkAkCKvNMD184fFCCK7HQ+BaRQ5NBmW6c7x3Ur2UL6MNswIqN/X9+ZTvRms151aF9speNnyNYZNDrmOrUoyIj8cAGPanKiCq1gY=",
        "proof": {
          "RootHash": "8E6F9DC69873E9F12AC1D84B7D25ED27039924663430042138B3CAA91584E9F6",
          "Data": "4QHwYl3uCmfObcBDChTpOvolH1oe67jjc93FA5QwJpSiIBIvRTkzQUZBMjUxRjVBMUVFQkI4RTM3M0REQzUwMzk0MzAyNjk0QTIyMC0xMDkzNTUaC1pDQi1GMDBfQk5CIAIoATCBRTiAqNa5B0ABEnIKJuta6YchAv+Fh/6SCT/nXtoa/c4s1p3cvs0B6EH24jX94NocvGBdEkAkCKvNMD184fFCCK7HQ+BaRQ5NBmW6c7x3Ur2UL6MNswIqN/X9+ZTvRms151aF9speNnyNYZNDrmOrUoyIj8cAGPanKiCq1gY=",
          "Proof": {
            "total": "1",
            "index": "0",
            "leaf_hash": "jm+dxphz6fEqwdhLfSXtJwOZJGY0MAQhOLPKqRWE6fY=",
            "aunts": []
          }
        }
      }
    ],
    "total_count": "1"
  }
}`, IsNil, IsNil)
	testFunc(`
{
  "jsonrpc": "2.0",
  "id": "",
  "result": {
    "txs": [
      {
        "hash": "10C4E872A5DC842BE72AC8DE9C6A13F97DF6D345336F01B87EBA998F5A3BC36D",
        "height": "35345060",
        "index": 0,
        "tx_result": {
          "log": "Msg 0: ",
          "tags": [
            {
              "key": "c2VuZGVy",
              "value": "dGJuYjFnZ2RjeWhrOHJjN2ZnenA4d2Eyc3UyMjBhY2xjZ2djc2Q5NHllNQ=="
            },
            {
              "key": "cmVjaXBpZW50",
              "value": "dGJuYjF5eWNuNG1oNmZmd3BqZjU4NHQ4bHBwN2MyN2dodTAzZ3B2cWtmag=="
            },
            {
              "key": "YWN0aW9u",
              "value": "c2VuZA=="
            }
          ]
        },
        "tx": "3gHwYl3uClYqLIf6CicKFEIbgl7HHjyUCCd3VQ4pT+4/hCMQEg8KCFJVTkUtQTFGEIDC1y8SJwoUITE67vpKXBkmh6rP8IfYV5F+PigSDwoIUlVORS1BMUYQgMLXLxJwCibrWumHIQOki6+6K5zhbjAndqURWmVv5ZVY+ePXfi/DxUTzcenLWhJAUr5kAtjMfsb+IO+7ligNJRXhpL8WZLkH0IIWeQ2Cb4xEcN8ANIVgKjzU6IQYOKnNYpoCpMWQJTYXFg+Q95ztCBiSsyogFRoMd2l0aGRyYXc6Qk5CIAE=",
        "proof": {
          "RootHash": "A06D7798436C26BAF00177873C901C8A2337F8B0C18A75AAA9D86D615BE24938",
          "Data": "3gHwYl3uClYqLIf6CicKFEIbgl7HHjyUCCd3VQ4pT+4/hCMQEg8KCFJVTkUtQTFGEIDC1y8SJwoUITE67vpKXBkmh6rP8IfYV5F+PigSDwoIUlVORS1BMUYQgMLXLxJwCibrWumHIQOki6+6K5zhbjAndqURWmVv5ZVY+ePXfi/DxUTzcenLWhJAUr5kAtjMfsb+IO+7ligNJRXhpL8WZLkH0IIWeQ2Cb4xEcN8ANIVgKjzU6IQYOKnNYpoCpMWQJTYXFg+Q95ztCBiSsyogFRoMd2l0aGRyYXc6Qk5CIAE=",
          "Proof": {
            "total": "1",
            "index": "0",
            "leaf_hash": "oG13mENsJrrwAXeHPJAciiM3+LDBinWqqdhtYVviSTg=",
            "aunts": []
          }
        }
      }
    ],
    "total_count": "1"
  }
}`, IsNil, IsNil)
	testFunc(`

{
  "jsonrpc": "2.0",
  "id": "",
  "result": {
    "txs": [
      {
        "hash": "523546F263ABA7BDDFFEE82B9A362D0B8BD4F114D58880CF78A77D4D43E7847A",
        "height": "35340678",
        "index": 0,
        "tx_result": {
          "log": "Msg 0: ",
          "tags": [
            {
              "key": "c2VuZGVy",
              "value": "dGJuYjF5eWNuNG1oNmZmd3BqZjU4NHQ4bHBwN2MyN2dodTAzZ3B2cWtmag=="
            },
            {
              "key": "cmVjaXBpZW50",
              "value": "dGJuYjFnZ2RjeWhrOHJjN2ZnenA4d2Eyc3UyMjBhY2xjZ2djc2Q5NHllNQ=="
            },
            {
              "key": "YWN0aW9u",
              "value": "c2VuZA=="
            }
          ]
        },
        "tx": "3gHwYl3uClgqLIf6CigKFCExOu76SlwZJoeqz/CH2FeRfj4oEhAKCFJVTkUtQTFGEICMjZ4CEigKFEIbgl7HHjyUCCd3VQ4pT+4/hCMQEhAKCFJVTkUtQTFGEICMjZ4CEnAKJuta6YchA4qGFSnfOnMFBASpOdYfdpTguZhKJaMZxir4RDzHeb6VEkCoKLgBJFSqxAJxwpeLxumNlKfj3Qtc4V+GVnGooRr/rmKCewweZ5Wc7xT3DqSdkB1oo169zcU5tYpVZm5hmwqJGIe5KiAKGgxPVVRCT1VORDo5NDY=",
        "proof": {
          "RootHash": "1EC05BB121F24DB3E4F04A2EC92710896218B614E94629D4443D3B05065ED46C",
          "Data": "3gHwYl3uClgqLIf6CigKFCExOu76SlwZJoeqz/CH2FeRfj4oEhAKCFJVTkUtQTFGEICMjZ4CEigKFEIbgl7HHjyUCCd3VQ4pT+4/hCMQEhAKCFJVTkUtQTFGEICMjZ4CEnAKJuta6YchA4qGFSnfOnMFBASpOdYfdpTguZhKJaMZxir4RDzHeb6VEkCoKLgBJFSqxAJxwpeLxumNlKfj3Qtc4V+GVnGooRr/rmKCewweZ5Wc7xT3DqSdkB1oo169zcU5tYpVZm5hmwqJGIe5KiAKGgxPVVRCT1VORDo5NDY=",
          "Proof": {
            "total": "1",
            "index": "0",
            "leaf_hash": "HsBbsSHyTbPk8EouyScQiWIYthTpRinURD07BQZe1Gw=",
            "aunts": []
          }
        }
      }
    ],
    "total_count": "1"
  }
}`, NotNil, IsNil)
	txInItem := testFunc(`

{
  "jsonrpc": "2.0",
  "id": "",
  "result": {
    "txs": [
      {
        "hash": "2C199678C1C33CF324DD99E373D5DF9437FBD2BA49E43E35EBB5B0F29180D93F",
        "height": "35339328",
        "index": 0,
        "tx_result": {
          "log": "Msg 0: ",
          "tags": [
            {
              "key": "c2VuZGVy",
              "value": "dGJuYjF5eWNuNG1oNmZmd3BqZjU4NHQ4bHBwN2MyN2dodTAzZ3B2cWtmag=="
            },
            {
              "key": "cmVjaXBpZW50",
              "value": "dGJuYjFnZ2RjeWhrOHJjN2ZnenA4d2Eyc3UyMjBhY2xjZ2djc2Q5NHllNQ=="
            },
            {
              "key": "YWN0aW9u",
              "value": "c2VuZA=="
            }
          ]
        },
        "tx": "3gHwYl3uClgqLIf6CigKFCExOu76SlwZJoeqz/CH2FeRfj4oEhAKCFJVTkUtQTFGEICG2PAkEigKFEIbgl7HHjyUCCd3VQ4pT+4/hCMQEhAKCFJVTkUtQTFGEICG2PAkEnAKJuta6YchA4qGFSnfOnMFBASpOdYfdpTguZhKJaMZxir4RDzHeb6VEkBPCTCewhYLFrS4MD90owM8zRfvBiQaR03HvqX2b9pYyQOXgzNnmy0aYhL4BY/IFHd6Zl8FpgI7pEqP8Ybn6FCSGIe5KiAJGgxPVVRCT1VORDo4MjU=",
        "proof": {
          "RootHash": "DE6B52EF102C6F18F5F417C3B2EA3DED324437AA0F85E3508C2DFC2EE0A97927",
          "Data": "3gHwYl3uClgqLIf6CigKFCExOu76SlwZJoeqz/CH2FeRfj4oEhAKCFJVTkUtQTFGEICG2PAkEigKFEIbgl7HHjyUCCd3VQ4pT+4/hCMQEhAKCFJVTkUtQTFGEICG2PAkEnAKJuta6YchA4qGFSnfOnMFBASpOdYfdpTguZhKJaMZxir4RDzHeb6VEkBPCTCewhYLFrS4MD90owM8zRfvBiQaR03HvqX2b9pYyQOXgzNnmy0aYhL4BY/IFHd6Zl8FpgI7pEqP8Ybn6FCSGIe5KiAJGgxPVVRCT1VORDo4MjU=",
          "Proof": {
            "total": "1",
            "index": "0",
            "leaf_hash": "3mtS7xAsbxj19BfDsuo97TJEN6oPheNQjC38LuCpeSc=",
            "aunts": []
          }
        }
      }
    ],
    "total_count": "1"
  }
}`, NotNil, IsNil)
	c.Check(txInItem, NotNil)
	c.Check(txInItem.Memo, Equals, "OUTBOUND:825")
	c.Check(txInItem.Sender, Equals, "tbnb1yycn4mh6ffwpjf584t8lpp7c27ghu03gpvqkfj")
	c.Check(len(txInItem.Coins), Equals, 1)
	c.Check(txInItem.Coins[0].Denom.String(), Equals, common.RuneA1FTicker.String())
	c.Check(txInItem.Coins[0].Amount.Uint64(), Equals, uint64(9900000000))
}
