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
	"github.com/tendermint/tendermint/crypto/secp256k1"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
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
		switch r.Method {
		case "getrawtransaction":
			httpTestHandler(c, rw, "../../../../test/fixtures/btc/tx.json")
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
