package blockscanner

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/bifrost/config"
	"gitlab.com/thorchain/bepswap/thornode/bifrost/metrics"
)

func Test(t *testing.T) { TestingT(t) }

type CommonBlockScannerTestSuite struct{}

var _ = Suite(&CommonBlockScannerTestSuite{})

func (CommonBlockScannerTestSuite) TestNewCommonBlockScanner(c *C) {
	mss := NewMockScannerStorage()
	m, err := metrics.NewMetrics(config.MetricConfiguration{
		Enabled:      false,
		ListenPort:   8080,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	})
	c.Check(m, NotNil)
	c.Check(err, IsNil)
	cbs, err := NewCommonBlockScanner(config.BlockScannerConfiguration{
		RPCHost: "",
	}, mss, m)
	c.Check(cbs, IsNil)
	c.Check(err, NotNil)
	cbs, err = NewCommonBlockScanner(config.BlockScannerConfiguration{
		RPCHost: "http://localhost",
	}, mss, nil)
	c.Check(cbs, IsNil)
	c.Check(err, NotNil)
	cbs, err = NewCommonBlockScanner(config.BlockScannerConfiguration{
		RPCHost: "http://localhost",
	}, mss, m)
	c.Check(cbs, NotNil)
	c.Check(err, IsNil)
}

const (
	blockResult = `{ "jsonrpc": "2.0", "id": "", "result": { "block_meta": { "block_id": { "hash": "D063E5F1562F93D46FD4F01CA24813DD60B919D1C39CC34EF1DBB0EA07D0F7F8", "parts": { "total": "1", "hash": "1D9E042DB7616CCB08AF785134561A9AA3074D6CC45A402DDE81572231FD7C91" } }, "header": { "version": { "block": "11", "app": "0" }, "chain_id": "Binance-Chain-Nile", "height": "11", "time": "2019-08-25T05:11:54.192630044Z", "num_txs": "0", "total_txs": "37507966", "last_block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "last_commit_hash": "03E65648ED376E0FBC5373E94394128AB76928E19A66FD2698D7AC9C8B212D33", "data_hash": "", "validators_hash": "80D9AB0FC10D18CA0E0832D5F4C063C5489EC1443DFB738252D038A82131B27A", "next_validators_hash": "80D9AB0FC10D18CA0E0832D5F4C063C5489EC1443DFB738252D038A82131B27A", "consensus_hash": "294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93", "app_hash": "CB951FFB480BCB8BC6FFB53A0AE4515E45C9873BA5B09B6C1ED59BF8F3D63D11", "last_results_hash": "9C7E9DFB57083B4FA4A9BD57519B6FB7E4B75E7D4CD26648815B6A806215C316", "evidence_hash": "", "proposer_address": "7B343E041CA130000A8BC00C35152BD7E7740037" } }, "block": { "header": { "version": { "block": "11", "app": "0" }, "chain_id": "Binance-Chain-Nile", "height": "11", "time": "2019-08-25T05:11:54.192630044Z", "num_txs": "0", "total_txs": "37507966", "last_block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "last_commit_hash": "03E65648ED376E0FBC5373E94394128AB76928E19A66FD2698D7AC9C8B212D33", "data_hash": "", "validators_hash": "80D9AB0FC10D18CA0E0832D5F4C063C5489EC1443DFB738252D038A82131B27A", "next_validators_hash": "80D9AB0FC10D18CA0E0832D5F4C063C5489EC1443DFB738252D038A82131B27A", "consensus_hash": "294D8FBD0B94B767A7EBA9840F299A3586DA7FE6B5DEAD3B7EECBA193C400F93", "app_hash": "CB951FFB480BCB8BC6FFB53A0AE4515E45C9873BA5B09B6C1ED59BF8F3D63D11", "last_results_hash": "9C7E9DFB57083B4FA4A9BD57519B6FB7E4B75E7D4CD26648815B6A806215C316", "evidence_hash": "", "proposer_address": "7B343E041CA130000A8BC00C35152BD7E7740037" }, "data": { "txs": null }, "evidence": { "evidence": null }, "last_commit": { "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "precommits": [ { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.201530916Z", "validator_address": "06FD60078EB4C2356137DD50036597DB267CF616", "validator_index": "0", "signature": "xaJAxeJJC+tG4hQOsDQr4uGEw8orINmkWBm6oZ7v92YbzqjTM088P+9o/v+Zg/0L/3tb69YU4QM19eu3OKt8AQ==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.164115501Z", "validator_address": "18E69CC672973992BB5F76D049A5B2C5DDF77436", "validator_index": "1", "signature": "6CWLmG1afETad9ThyFL3UrOx5VCv3a7HGAWMYSvExaJlfW562VjefMlFLqesQYLqgr3BtE4poJ8aFrN/zauvDg==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.164945213Z", "validator_address": "344C39BB8F4512D6CAB1F6AAFAC1811EF9D8AFDF", "validator_index": "2", "signature": "BG6u+vmptI5CXAZ6arH9brXQvtBmcWFUx8c4WzIcrftS+JAK2TuhnpcLNUPl9VNw9LBxatCnX60F7L014pKBBA==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.192630044Z", "validator_address": "37EF19AF29679B368D2B9E9DE3F8769B35786676", "validator_index": "3", "signature": "womUxsg21B/6/lXyweBUv0oz4bP1BHoK9BgtbiXSMKfDpb1iGlkZNmZSITyN03hyXabtjsF2AMjGcIzvW6FqAw==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.235226587Z", "validator_address": "62633D9DB7ED78E951F79913FDC8231AA77EC12B", "validator_index": "4", "signature": "WUCR3OR0d0NN2QlXD8xmdQpZo6vIeSHJUOajIlcj7BWmiqWgBEhrURcOaTDE//Zv99oO11ySDu5vGeEFpxNaCw==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.234781133Z", "validator_address": "7B343E041CA130000A8BC00C35152BD7E7740037", "validator_index": "5", "signature": "ejBogd89wMLUu4wfc24RblmGdZFwTNYlLzcC09tN5+TnrbBjAxeF3NbFd8nAsEtI6IGFngMp+mXdpFa6PNntCA==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.163726597Z", "validator_address": "91844D296BD8E591448EFC65FD6AD51A888D58FA", "validator_index": "6", "signature": "hhDq9bOctfjTScJXOAo+uKOwK/m9mWmykcDsrMPDRJQR5HRSekx8sBi7yvTqwzzePtyxux6NoCG6KKGKVuECAA==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.164570853Z", "validator_address": "B3727172CE6473BC780298A2D66C12F1A14F5B2A", "validator_index": "7", "signature": "0vXT7lOpb1+0/nTHJOLP8USjJl9SG3eGRlxxy2H0fFpPaiCS1cPb8ZyEHmjrZvhwRaNxuvkSFsyC32uuPx7QAw==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.191959099Z", "validator_address": "B6F20C7FAA2B2F6F24518FA02B71CB5F4A09FBA3", "validator_index": "8", "signature": "i3RB4OxsJf+h0nYqXn6xyc17PhN+RD5SSdIfhBGFfWBA2UsoPCCm5MawSSTvgFYDeRvdp5M+09RsSDXh8Dm2Ag==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.237319039Z", "validator_address": "E0DD72609CC106210D1AA13936CB67B93A0AEE21", "validator_index": "9", "signature": "PNdyXGrUK9DtRS3hgCNkFiToGg2QsNrV5Mdakr1/66OVDP6noGz2RaIZY/PHowZlcoWsfPGtSVP5C7U2BRT3Bw==" }, { "type": 2, "height": "35526651", "round": "0", "block_id": { "hash": "7B67CE1848B7BF3127218B8A27C178968FA2AC7C1EB49C7042E5622189EDD4FA", "parts": { "total": "1", "hash": "0FFFBD74F728723EEFBFDCE0322D164385D30E17753DC8513C695ED83217A740" } }, "timestamp": "2019-08-25T05:11:54.234516001Z", "validator_address": "FC3108DC3814888F4187452182BC1BAF83B71BC9", "validator_index": "10", "signature": "yMyPeQTEc7Gu+AZwTROZd8+fMnmSu+MYo9Rf9LBVhtWC2BYGJqfAr3Ctgy9Tn7yngj3jUFPPa5AyPOt3b9bFBA==" } ] } } }}`
)

func (CommonBlockScannerTestSuite) TestBlockScanner(c *C) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Logf("================>:%s", r.RequestURI)
		switch r.RequestURI {
		case "/block": // trying to get block
			if _, err := w.Write([]byte(blockResult)); nil != err {
				c.Error(err)
			}
		}
	})
	mss := NewMockScannerStorage()
	s := httptest.NewTLSServer(h)
	defer s.Close()
	m, err := metrics.NewMetrics(config.MetricConfiguration{
		Enabled:      false,
		ListenPort:   8080,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	})
	c.Check(m, NotNil)
	c.Check(err, IsNil)
	cbs, err := NewCommonBlockScanner(config.BlockScannerConfiguration{
		RPCHost:                    s.URL,
		StartBlockHeight:           0,
		BlockScanProcessors:        1,
		HttpRequestTimeout:         time.Second,
		HttpRequestReadTimeout:     time.Second * 30,
		HttpRequestWriteTimeout:    time.Second * 30,
		MaxHttpRequestRetry:        3,
		BlockHeightDiscoverBackoff: time.Second,
		BlockRetryInterval:         time.Second,
	}, mss, m)
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
	time.Sleep(time.Second * 10)
	err = cbs.Stop()
	c.Check(err, IsNil)
	c.Check(counter, Equals, 10)

}
