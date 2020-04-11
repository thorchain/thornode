package bitcoin

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/btcsuite/btcd/btcjson"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

func TestPackage(t *testing.T) { TestingT(t) }

type BitcoinSuite struct {
	client  *Client
	server  *httptest.Server
	cfg     config.ChainConfiguration
	cleanup func()
}

var _ = Suite(
	&BitcoinSuite{},
)

func (s *BitcoinSuite) SetUpSuite(c *C) {
	s.cfg = config.ChainConfiguration{
		ChainID:     "BTC",
		UserName:    "bob",
		Password:    "password",
		DisableTLS:  true,
		HTTPostMode: true,
	}
	s.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r := struct {
			Method string `json:"method"`
		}{}
		json.NewDecoder(req.Body).Decode(&r)
		switch {
		case r.Method == "getblockhash":
			httpTestHandler(c, rw, "../../../../test/fixtures/btc/blockhash.json")
		case r.Method == "getblock":
			httpTestHandler(c, rw, "../../../../test/fixtures/btc/block.json")
		case r.Method == "getrawtransaction":
			httpTestHandler(c, rw, "../../../../test/fixtures/btc/tx.json")
		}
	}))

	s.cfg.ChainHost = s.server.Listener.Addr().String()
	var err error
	s.client, err = NewClient(s.cfg)
	c.Assert(err, IsNil)
	c.Assert(s.client, NotNil)
}

func (s *BitcoinSuite) TearDownSuite(c *C) {
	s.server.Close()
}

func httpTestHandler(c *C, rw http.ResponseWriter, fixture string) {
	content, err := ioutil.ReadFile(fixture)
	if err != nil {
		c.Fatal(err)
	}
	rw.Header().Set("Content-Type", "application/json")
	if _, err := rw.Write(content); err != nil {
		c.Fatal(err)
	}
}

func (s *BitcoinSuite) TestGetBlock(c *C) {
	block, err := s.client.getBlock(1696761)
	c.Assert(err, IsNil)
	c.Assert(block.Hash, Equals, "000000008de7a25f64f9780b6c894016d2c63716a89f7c9e704ebb7e8377a0c8")
	c.Assert(block.RawTx[0].Txid, Equals, "31f8699ce9028e9cd37f8a6d58a79e614a96e3fdd0f58be5fc36d2d95484716f")
	c.Assert(len(block.RawTx), Equals, 110)
}

func (s *BitcoinSuite) TestFetchTxs(c *C) {
	txs, err := s.client.FetchTxs(0)
	c.Assert(err, IsNil)
	c.Assert(txs.BlockHeight, Equals, "1696761")
	c.Assert(txs.Chain, Equals, common.BTCChain)
	c.Assert(txs.TxArray[0].Tx, Equals, "3075045b8fe31659634d57084c9c8979f8c91029994dc9ab0b9444f1e793603a:0")
	c.Assert(txs.TxArray[0].Sender, Equals, "tb1qdxxlx4r4jk63cve3rjpj428m26xcukjn5yegff")
	c.Assert(txs.TxArray[0].To, Equals, "tb1qkq7weysjn6ljc2ywmjmwp8ttcckg8yyxjdz5k6")
	c.Assert(txs.TxArray[0].Coins.Equals(common.Coins{common.NewCoin(common.BTCAsset, sdk.NewUint(1090601))}), Equals, true)
	c.Assert(len(txs.TxArray), Equals, 4)
}

func (s *BitcoinSuite) TestGetSender(c *C) {
	tx := btcjson.TxRawResult{
		Vin: []btcjson.Vin{
			btcjson.Vin{
				Txid: "31f8699ce9028e9cd37f8a6d58a79e614a96e3fdd0f58be5fc36d2d95484716f",
				Vout: 0,
			},
		},
	}
	sender, err := s.client.getSender(&tx)
	c.Assert(err, IsNil)
	c.Assert(sender, Equals, "n3jYBjCzgGNydQwf83Hz6GBzGBhMkKfgL1")

	tx.Vin[0].Vout = 1
	sender, err = s.client.getSender(&tx)
	c.Assert(err, IsNil)
	c.Assert(sender, Equals, "tb1qdxxlx4r4jk63cve3rjpj428m26xcukjn5yegff")
}

func (s *BitcoinSuite) TestGetMemo(c *C) {
	tx := btcjson.TxRawResult{
		Vout: []btcjson.Vout{
			btcjson.Vout{
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm: "OP_RETURN 74686f72636861696e3a636f6e736f6c6964617465",
				},
			},
		},
	}
	memo, err := s.client.getMemo(&tx)
	c.Assert(err, IsNil)
	c.Assert(memo, Equals, "thorchain:consolidate")

	tx = btcjson.TxRawResult{
		Vout: []btcjson.Vout{
			btcjson.Vout{
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm: "OP_RETURN 737761703a6574682e3078633534633135313236393646334541373935366264396144343130383138654563414443466666663a30786335346331353132363936463345413739353662643961443431",
				},
			},
			btcjson.Vout{
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm: "OP_RETURN 30383138654563414443466666663a3130303030303030303030",
				},
			},
		},
	}
	memo, err = s.client.getMemo(&tx)
	c.Assert(err, IsNil)
	c.Assert(memo, Equals, "swap:eth.0xc54c1512696F3EA7956bd9aD410818eEcADCFfff:0xc54c1512696F3EA7956bd9aD410818eEcADCFfff:10000000000")
}

func (s *BitcoinSuite) TestIgnoreTx(c *C) {
	// valid tx that will NOT be ignored
	tx := btcjson.TxRawResult{
		Vin: []btcjson.Vin{
			btcjson.Vin{
				Txid: "24ed2d26fd5d4e0e8fa86633e40faf1bdfc8d1903b1cd02855286312d48818a2",
				Vout: 0,
			},
		},
		Vout: []btcjson.Vout{
			btcjson.Vout{
				Value: 0.12345678,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Addresses: []string{"tb1qkq7weysjn6ljc2ywmjmwp8ttcckg8yyxjdz5k6"},
				},
			},
			btcjson.Vout{
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm: "OP_RETURN 74686f72636861696e3a636f6e736f6c6964617465",
				},
			},
		},
	}
	ignored := s.client.ignoreTx(&tx)
	c.Assert(ignored, Equals, false)

	// invalid tx missing Vout
	tx = btcjson.TxRawResult{
		Vin: []btcjson.Vin{
			btcjson.Vin{
				Txid: "24ed2d26fd5d4e0e8fa86633e40faf1bdfc8d1903b1cd02855286312d48818a2",
				Vout: 0,
			},
		},
		Vout: []btcjson.Vout{},
	}
	ignored = s.client.ignoreTx(&tx)
	c.Assert(ignored, Equals, true)

	// invalid tx missing vout[0].Value == no coins
	tx = btcjson.TxRawResult{
		Vin: []btcjson.Vin{
			btcjson.Vin{
				Txid: "24ed2d26fd5d4e0e8fa86633e40faf1bdfc8d1903b1cd02855286312d48818a2",
				Vout: 0,
			},
		},
		Vout: []btcjson.Vout{
			btcjson.Vout{
				Value: 0,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Addresses: []string{"tb1qkq7weysjn6ljc2ywmjmwp8ttcckg8yyxjdz5k6"},
				},
			},
			btcjson.Vout{
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm: "OP_RETURN 74686f72636861696e3a636f6e736f6c6964617465",
				},
			},
		},
	}
	ignored = s.client.ignoreTx(&tx)
	c.Assert(ignored, Equals, true)

	// invalid tx missing vin[0].Txid means coinbase
	tx = btcjson.TxRawResult{
		Vin: []btcjson.Vin{
			btcjson.Vin{
				Txid: "",
				Vout: 0,
			},
		},
		Vout: []btcjson.Vout{
			btcjson.Vout{
				Value: 0.1234565,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Addresses: []string{"tb1qkq7weysjn6ljc2ywmjmwp8ttcckg8yyxjdz5k6"},
				},
			},
			btcjson.Vout{
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm: "OP_RETURN 74686f72636861696e3a636f6e736f6c6964617465",
				},
			},
		},
	}
	ignored = s.client.ignoreTx(&tx)
	c.Assert(ignored, Equals, true)

	// invalid tx missing vin
	tx = btcjson.TxRawResult{
		Vin: []btcjson.Vin{},
		Vout: []btcjson.Vout{
			btcjson.Vout{
				Value: 0.1234565,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Addresses: []string{"tb1qkq7weysjn6ljc2ywmjmwp8ttcckg8yyxjdz5k6"},
				},
			},
			btcjson.Vout{
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm: "OP_RETURN 74686f72636861696e3a636f6e736f6c6964617465",
				},
			},
		},
	}
	ignored = s.client.ignoreTx(&tx)
	c.Assert(ignored, Equals, true)

	// invalid tx multiple vout[0].Addresses
	tx = btcjson.TxRawResult{
		Vin: []btcjson.Vin{
			btcjson.Vin{
				Txid: "24ed2d26fd5d4e0e8fa86633e40faf1bdfc8d1903b1cd02855286312d48818a2",
				Vout: 0,
			},
		},
		Vout: []btcjson.Vout{
			btcjson.Vout{
				Value: 0.1234565,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Addresses: []string{
						"tb1qkq7weysjn6ljc2ywmjmwp8ttcckg8yyxjdz5k6",
						"bc1q0s4mg25tu6termrk8egltfyme4q7sg3h0e56p3",
					},
				},
			},
			btcjson.Vout{
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm: "OP_RETURN 74686f72636861696e3a636f6e736f6c6964617465",
				},
			},
		},
	}
	ignored = s.client.ignoreTx(&tx)
	c.Assert(ignored, Equals, true)

	// invalid tx > 2 vout with coins we only expect 2 max
	tx = btcjson.TxRawResult{
		Vin: []btcjson.Vin{
			btcjson.Vin{
				Txid: "24ed2d26fd5d4e0e8fa86633e40faf1bdfc8d1903b1cd02855286312d48818a2",
				Vout: 0,
			},
		},
		Vout: []btcjson.Vout{
			btcjson.Vout{
				Value: 0.1234565,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Addresses: []string{
						"bc1q0s4mg25tu6termrk8egltfyme4q7sg3h0e56p3",
					},
				},
			},
			btcjson.Vout{
				Value: 0.1234565,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Addresses: []string{
						"tb1qkq7weysjn6ljc2ywmjmwp8ttcckg8yyxjdz5k6",
					},
				},
			},
			btcjson.Vout{
				Value: 0.1234565,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Addresses: []string{
						"tb1qkq7weysjn6ljc2ywmjmwp8ttcckg8yyxjdz5k6",
					},
				},
			},
			btcjson.Vout{
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm: "OP_RETURN 74686f72636861696e3a636f6e736f6c6964617465",
				},
			},
		},
	}
	ignored = s.client.ignoreTx(&tx)
	c.Assert(ignored, Equals, true)

	// valid tx == 2 vout with coins, 1 to vault, 1 with change back to user
	tx = btcjson.TxRawResult{
		Vin: []btcjson.Vin{
			btcjson.Vin{
				Txid: "24ed2d26fd5d4e0e8fa86633e40faf1bdfc8d1903b1cd02855286312d48818a2",
				Vout: 0,
			},
		},
		Vout: []btcjson.Vout{
			btcjson.Vout{
				Value: 0.1234565,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Addresses: []string{
						"bc1q0s4mg25tu6termrk8egltfyme4q7sg3h0e56p3",
					},
				},
			},
			btcjson.Vout{
				Value: 0.1234565,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Addresses: []string{
						"tb1qkq7weysjn6ljc2ywmjmwp8ttcckg8yyxjdz5k6",
					},
				},
			},
			btcjson.Vout{
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm: "OP_RETURN 74686f72636861696e3a636f6e736f6c6964617465",
				},
			},
		},
	}
	ignored = s.client.ignoreTx(&tx)
	c.Assert(ignored, Equals, false)
}