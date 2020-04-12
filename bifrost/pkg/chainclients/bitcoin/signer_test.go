package bitcoin

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"time"

	ctypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/cosmos/cosmos-sdk/client/keys"
	cKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/common"
	types2 "gitlab.com/thorchain/thornode/x/thorchain/types"
)

type BitcoinSignerSuite struct {
	client  *Client
	server  *httptest.Server
	cfg     config.ChainConfiguration
	cleanup func()
}

var _ = Suite(&BitcoinSignerSuite{})

func (s *BitcoinSignerSuite) SetUpSuite(c *C) {
	s.cfg = config.ChainConfiguration{
		ChainID:     "BTC",
		UserName:    "bob",
		Password:    "password",
		DisableTLS:  true,
		HTTPostMode: true,
	}
	ns := strconv.Itoa(time.Now().Nanosecond())
	types2.SetupConfigForTest()
	ctypes.Network = ctypes.TestNetwork
	c.Assert(os.Setenv("NET", "testnet"), IsNil)

	thordir := filepath.Join(os.TempDir(), ns, ".thorcli")
	cfg := config.ClientConfiguration{
		ChainID:         "thorchain",
		ChainHost:       "localhost",
		SignerName:      "bob",
		SignerPasswd:    "password",
		ChainHomeFolder: thordir,
	}

	kb, err := keys.NewKeyBaseFromDir(thordir)
	c.Assert(err, IsNil)
	_, _, err = kb.CreateMnemonic(cfg.SignerName, cKeys.English, cfg.SignerPasswd, cKeys.Secp256k1)
	c.Assert(err, IsNil)
	thorKeys, err := thorclient.NewKeys(cfg.ChainHomeFolder, cfg.SignerName, cfg.SignerPasswd)
	c.Assert(err, IsNil)
	s.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r := struct {
			Method string `json:"method"`
		}{}
		json.NewDecoder(req.Body).Decode(&r)
		defer func() {
			c.Assert(req.Body.Close(), IsNil)
		}()
		switch r.Method {
		case "getrawtransaction":
			httpTestHandler(c, rw, "../../../../test/fixtures/btc/tx.json")
		case "getinfo":
			httpTestHandler(c, rw, "../../../../test/fixtures/btc/getinfo.json")
		case "sendrawtransaction":
			httpTestHandler(c, rw, "../../../../test/fixtures/btc/sendrawtransaction.json")
		}
	}))

	s.cfg.ChainHost = s.server.Listener.Addr().String()
	s.client, err = NewClient(thorKeys, s.cfg, nil)
	c.Assert(err, IsNil)
	c.Assert(s.client, NotNil)
}

func (s *BitcoinSignerSuite) TearDownSuite(c *C) {
	s.server.Close()
}

func (s *BitcoinSignerSuite) TestGetBTCPrivateKey(c *C) {
	input := "YjQwNGM1ZWM1ODExNmI1ZjBmZTEzNDY0YTkyZTQ2NjI2ZmM1ZGIxMzBlNDE4Y2JjZTk4ZGY4NmZmZTkzMTdjNQ=="
	buf, err := base64.StdEncoding.DecodeString(input)
	c.Assert(err, IsNil)
	c.Assert(buf, NotNil)
	prikeyByte, err := hex.DecodeString(string(buf))
	c.Assert(err, IsNil)
	pk := secp256k1.GenPrivKeySecp256k1(prikeyByte)
	btcPrivateKey, err := getBTCPrivateKey(pk)
	c.Assert(err, IsNil)
	c.Assert(btcPrivateKey, NotNil)
}

func (s *BitcoinSignerSuite) TestGetChainCfg(c *C) {
	os.Setenv("NET", "testnet")
	defer os.Remove("NET")
	param := s.client.getChainCfg()
	c.Assert(param, Equals, &chaincfg.TestNet3Params)
	os.Setenv("NET", "mainnet")
	param = s.client.getChainCfg()
	c.Assert(param, Equals, &chaincfg.MainNetParams)
}

func (s *BitcoinSignerSuite) TestGetLastOutput(c *C) {
	vOut, err := s.client.getLastOutput("xxxx", "xxxx")
	c.Assert(err, NotNil)
	c.Assert(vOut.Value, Equals, float64(0))
	vOut, err = s.client.getLastOutput("31f8699ce9028e9cd37f8a6d58a79e614a96e3fdd0f58be5fc36d2d95484716f", "tb1qdxxlx4r4jk63cve3rjpj428m26xcukjn5yegff")
	c.Assert(err, IsNil)
	c.Assert(vOut.Value, Equals, 0.19590108)
}

func (s *BitcoinSignerSuite) TestSignTx(c *C) {
	txOutItem := stypes.TxOutItem{
		Chain:       common.BNBChain,
		ToAddress:   types2.GetRandomBNBAddress(),
		VaultPubKey: types2.GetRandomPubKey(),
		SeqNo:       0,
		Coins: common.Coins{
			common.NewCoin(common.BTCAsset, sdk.NewUint(10)),
		},
		Memo: "whatever",
		MaxGas: common.Gas{
			common.NewCoin(common.BTCAsset, sdk.NewUint(1)),
		},
		InHash:  "",
		OutHash: "",
	}
	// incorrect chain should return an error
	result, err := s.client.SignTx(txOutItem, 1)
	c.Assert(err, NotNil)
	c.Assert(result, IsNil)

	// invalid pubkey should return an error
	txOutItem.Chain = common.BTCChain
	txOutItem.VaultPubKey = common.PubKey("helloworld")
	result, err = s.client.SignTx(txOutItem, 2)
	c.Assert(err, NotNil)
	c.Assert(result, IsNil)

	// invalid to address should return an error
	txOutItem.VaultPubKey = types2.GetRandomPubKey()
	result, err = s.client.SignTx(txOutItem, 3)
	c.Assert(err, NotNil)
	c.Assert(result, IsNil)

	addr, err := types2.GetRandomPubKey().GetAddress(common.BTCChain)
	c.Assert(err, IsNil)
	txOutItem.ToAddress = addr

	result, err = s.client.SignTx(txOutItem, 4)
	c.Assert(err, NotNil)
	c.Assert(result, IsNil)
}

func (s *BitcoinSignerSuite) TestBroadcastTx(c *C) {
	txOutItem := stypes.TxOutItem{
		Chain:       common.BNBChain,
		ToAddress:   types2.GetRandomBNBAddress(),
		VaultPubKey: types2.GetRandomPubKey(),
		SeqNo:       0,
		Coins: common.Coins{
			common.NewCoin(common.BTCAsset, sdk.NewUint(10)),
		},
		Memo: "whatever",
		MaxGas: common.Gas{
			common.NewCoin(common.BTCAsset, sdk.NewUint(1)),
		},
		InHash:  "",
		OutHash: "",
	}
	input := []byte("hello world")
	c.Assert(s.client.BroadcastTx(txOutItem, input), NotNil)
	input1, err := hex.DecodeString("01000000000103c7d45551ff54354be6711396560348ebbf273b989b542be36645568ed1dbecf10000000000ffffffff951ed70edc0bf2a4b3e1cbfe55d191a72850c5595c381309f69fc084c9af0b540100000000ffffffffc5db14c562b96bfd95f97d74a558a3e3b91841a96e1b09546208c9fb67424f420000000000ffffffff02231710000000000016001417acb08a31369e7666d94664d7e64f0e048220900000000000000000176a1574686f72636861696e3a636f6e736f6c6964617465024730440220756d15a363b78b070b583dfc1a6aba0dd605550407d5d3d92f5e785ef7e42aca02200db19dab144033c9c353481be30469da42c0c0a7580a513f49282bea77d7a29301210223da2ff73fa9b2258d335a4e63a4e7ef88211b8e800588280ed8b51e285ec0ff02483045022100a695f0fece36de02212b10bf6aa2f06dc6ef84ba30cae0c78749deddba1574530220315b490111c830c27e6cb810559c2a37cd00b123de82df79e061df26c8deb14301210223da2ff73fa9b2258d335a4e63a4e7ef88211b8e800588280ed8b51e285ec0ff0247304402207e586439b04985a90a53cf9fc511a53d86acece57b3e5571118562449d4f27ac02206d84f0fba1a68cf55efc8a1c2ec768924479b97ceaf2029ed6941176f004bf8101210223da2ff73fa9b2258d335a4e63a4e7ef88211b8e800588280ed8b51e285ec0ff00000000")
	c.Assert(err, IsNil)
	c.Assert(s.client.BroadcastTx(txOutItem, input1), IsNil)
}
