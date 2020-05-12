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
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/client/keys"
	cKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"gitlab.com/thorchain/txscript"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/bifrost/tss"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain"
	types2 "gitlab.com/thorchain/thornode/x/thorchain/types"
)

type BitcoinSignerSuite struct {
	client  *Client
	server  *httptest.Server
	bridge  *thorclient.ThorchainBridge
	cfg     config.ChainConfiguration
	m       *metrics.Metrics
	cleanup func()
}

var _ = Suite(&BitcoinSignerSuite{})

func (s *BitcoinSignerSuite) SetUpTest(c *C) {
	s.m = GetMetricForTest(c)
	s.cfg = config.ChainConfiguration{
		ChainID:     "BTC",
		UserName:    "bob",
		Password:    "password",
		DisableTLS:  true,
		HTTPostMode: true,
		BlockScanner: config.BlockScannerConfiguration{
			StartBlockHeight: 1, // avoids querying thorchain for block height
		},
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
		if req.RequestURI == "/thorchain/vaults/thorpub1addwnpepqts24euwrgly2vtez3zdvusmk6u3cwf8leuzj8m4ynvmv5cst7us2vltqrh/signers" {
			_, err := rw.Write([]byte("[]"))
			c.Assert(err, IsNil)
		} else {
			r := struct {
				Method string `json:"method"`
			}{}
			json.NewDecoder(req.Body).Decode(&r)
			defer func() {
				c.Assert(req.Body.Close(), IsNil)
			}()
			switch r.Method {
			case "getbestblockhash":
				httpTestHandler(c, rw, "../../../../test/fixtures/btc/getbestblockhash.json")
			case "getblock":
				httpTestHandler(c, rw, "../../../../test/fixtures/btc/block.json")
			case "getrawtransaction":
				httpTestHandler(c, rw, "../../../../test/fixtures/btc/tx.json")
			case "getinfo":
				httpTestHandler(c, rw, "../../../../test/fixtures/btc/getinfo.json")
			case "sendrawtransaction":
				httpTestHandler(c, rw, "../../../../test/fixtures/btc/sendrawtransaction.json")
			}
		}
	}))

	s.cfg.RPCHost = s.server.Listener.Addr().String()
	cfg.ChainHost = s.server.Listener.Addr().String()
	s.bridge, err = thorclient.NewThorchainBridge(cfg, s.m)
	c.Assert(err, IsNil)
	s.client, err = NewClient(thorKeys, s.cfg, nil, s.bridge, s.m)
	storage := storage.NewMemStorage()
	db, err := leveldb.Open(storage, nil)
	c.Assert(err, IsNil)
	accessor, err := NewLevelDBBlockMetaAccessor(db)
	c.Assert(err, IsNil)
	s.client.blockMetaAccessor = accessor
	c.Assert(err, IsNil)
	c.Assert(s.client, NotNil)
}

func (s *BitcoinSignerSuite) TearDownTest(c *C) {
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
	defer os.Setenv("NET", "testnet")
	param := s.client.getChainCfg()
	c.Assert(param, Equals, &chaincfg.TestNet3Params)
	os.Setenv("NET", "mainnet")
	param = s.client.getChainCfg()
	c.Assert(param, Equals, &chaincfg.MainNetParams)
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

	// nothing to sign , because there is not enough UTXO
	result, err = s.client.SignTx(txOutItem, 4)
	c.Assert(err, NotNil)
	c.Assert(result, IsNil)

	blockMeta := NewBlockMeta("", 100, "")
	blockMeta.AddUTXO(GetRandomUTXO(0.5))
	c.Assert(s.client.blockMetaAccessor.SaveBlockMeta(100, blockMeta), IsNil)

	result, err = s.client.SignTx(txOutItem, 5)
	c.Assert(err, NotNil)
	c.Assert(result, IsNil)
}

func (s *BitcoinSignerSuite) TestSignTxHappyPathWithPrivateKey(c *C) {
	addr, err := types2.GetRandomPubKey().GetAddress(common.BTCChain)
	c.Assert(err, IsNil)
	txOutItem := stypes.TxOutItem{
		Chain:       common.BTCChain,
		ToAddress:   addr,
		VaultPubKey: "thorpub1addwnpepqw2k68efthm08f0f5akhjs6fk5j2pze4wkwt4fmnymf9yd463puru988m2y",
		SeqNo:       0,
		Coins: common.Coins{
			common.NewCoin(common.BTCAsset, sdk.NewUint(10)),
		},
		MaxGas: common.Gas{
			common.NewCoin(common.BTCAsset, sdk.NewUint(1)),
		},
		InHash:  "",
		OutHash: "",
	}
	txHash, err := chainhash.NewHashFromStr("256222fb25a9950479bb26049a2c00e75b89abbb7f0cf646c623b93e942c4c34")
	c.Assert(err, IsNil)
	utxo := NewUnspentTransactionOutput(*txHash, 0, 0.01049996, 100, txOutItem.VaultPubKey)
	blockMeta := NewBlockMeta("000000000000008a0da55afa8432af3b15c225cc7e04d32f0de912702dd9e2ae",
		100,
		"0000000000000068f0710c510e94bd29aa624745da43e32a1de887387306bfda")
	blockMeta.AddUTXO(utxo)
	c.Assert(s.client.blockMetaAccessor.SaveBlockMeta(blockMeta.Height, blockMeta), IsNil)
	priKeyBuf, err := hex.DecodeString("b404c5ec58116b5f0fe13464a92e46626fc5db130e418cbce98df86ffe9317c5")
	c.Assert(err, IsNil)
	pkey, _ := btcec.PrivKeyFromBytes(btcec.S256(), priKeyBuf)
	c.Assert(pkey, NotNil)
	ksw, err := NewKeySignWrapper(pkey, s.client.bridge, s.client.ksWrapper.tssKeyManager)
	c.Assert(err, IsNil)
	s.client.privateKey = pkey
	s.client.ksWrapper = ksw
	vaultPubKey, err := GetBech32AccountPubKey(pkey)
	c.Assert(err, IsNil)
	txOutItem.VaultPubKey = vaultPubKey
	buf, err := s.client.SignTx(txOutItem, 1)
	c.Assert(err, IsNil)
	c.Assert(buf, NotNil)
}

func (s *BitcoinSignerSuite) TestSignTxWithTSS(c *C) {
	pubkey, err := common.NewPubKey("thorpub1addwnpepqts24euwrgly2vtez3zdvusmk6u3cwf8leuzj8m4ynvmv5cst7us2vltqrh")
	c.Assert(err, IsNil)
	addr, err := pubkey.GetAddress(common.BTCChain)
	c.Assert(err, IsNil)
	txOutItem := stypes.TxOutItem{
		Chain:       common.BTCChain,
		ToAddress:   addr,
		VaultPubKey: "thorpub1addwnpepqts24euwrgly2vtez3zdvusmk6u3cwf8leuzj8m4ynvmv5cst7us2vltqrh",
		SeqNo:       0,
		Coins: common.Coins{
			common.NewCoin(common.BTCAsset, sdk.NewUint(10)),
		},
		MaxGas: common.Gas{
			common.NewCoin(common.BTCAsset, sdk.NewUint(1)),
		},
		InHash:  "",
		OutHash: "",
	}
	thorKeyManager := &tss.MockThorchainKeyManager{}
	s.client.ksWrapper, err = NewKeySignWrapper(s.client.privateKey, s.client.bridge, thorKeyManager)
	txHash, err := chainhash.NewHashFromStr("66d2d6b5eb564972c59e4797683a1225a02515a41119f0a8919381236b63e948")
	c.Assert(err, IsNil)
	utxo := NewUnspentTransactionOutput(*txHash, 0, 0.00018, 100, txOutItem.VaultPubKey)
	blockMeta := NewBlockMeta("000000000000008a0da55afa8432af3b15c225cc7e04d32f0de912702dd9e2ae",
		100,
		"0000000000000068f0710c510e94bd29aa624745da43e32a1de887387306bfda")
	blockMeta.AddUTXO(utxo)
	c.Assert(s.client.blockMetaAccessor.SaveBlockMeta(blockMeta.Height, blockMeta), IsNil)
	buf, err := s.client.SignTx(txOutItem, 1)
	c.Assert(err, IsNil)
	c.Assert(buf, NotNil)
}

func GetRandomUTXO(amount float64) UnspentTransactionOutput {
	tx := wire.NewMsgTx(wire.TxVersion)
	pk := types2.GetRandomPubKey()
	addr, _ := pk.GetAddress(common.BTCChain)
	btcAddr, _ := btcutil.DecodeAddress(addr.String(), &chaincfg.TestNet3Params)
	script, _ := txscript.PayToAddrScript(btcAddr)
	btcAmt, _ := btcutil.NewAmount(amount)
	tx.AddTxOut(wire.NewTxOut(int64(btcAmt), script))
	return NewUnspentTransactionOutput(tx.TxHash(), 0, amount, 10, pk)
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

func (s *BitcoinSignerSuite) TestGetAllUTXOs(c *C) {
	vaultPubKey := thorchain.GetRandomPubKey()
	for i := 0; i < 150; i++ {
		previousHash := thorchain.GetRandomTxHash().String()
		blockHash := thorchain.GetRandomTxHash().String()
		blockMeta := NewBlockMeta(previousHash, int64(i), blockHash)
		utxo := GetRandomUTXO(1.0)
		utxo.VaultPubKey = vaultPubKey
		utxo.BlockHeight = int64(i)
		blockMeta.AddUTXO(utxo)
		c.Assert(s.client.blockMetaAccessor.SaveBlockMeta(blockMeta.Height, blockMeta), IsNil)
	}
	utxoes, err := s.client.getAllUtxos(150, vaultPubKey, 10)
	c.Assert(err, IsNil)

	// include block height 0 ~ 51
	c.Assert(utxoes, HasLen, 52)

	// mark them as spent
	for _, utxo := range utxoes {
		blockMeta, err := s.client.blockMetaAccessor.GetBlockMeta(utxo.BlockHeight)
		c.Assert(err, IsNil)
		blockMeta.SpendUTXO(utxo.GetKey())
		c.Assert(s.client.blockMetaAccessor.SaveBlockMeta(blockMeta.Height, blockMeta), IsNil)
	}

	// check prune is not returning them when spent
	c.Assert(s.client.blockMetaAccessor.PruneBlockMeta(150-BlockCacheSize), IsNil)
	allmetas, err := s.client.blockMetaAccessor.GetBlockMetas()
	c.Assert(err, IsNil)
	c.Assert(allmetas, HasLen, 100)

	// make sure block will not be Pruned when there are unspend UTXO in it
	for i := 150; i < 200; i++ {
		previousHash := thorchain.GetRandomTxHash().String()
		blockHash := thorchain.GetRandomTxHash().String()
		blockMeta := NewBlockMeta(previousHash, int64(i), blockHash)
		utxo := GetRandomUTXO(1.0)
		utxo.VaultPubKey = vaultPubKey
		utxo.BlockHeight = int64(i)
		blockMeta.AddUTXO(utxo)
		c.Assert(s.client.blockMetaAccessor.SaveBlockMeta(blockMeta.Height, blockMeta), IsNil)
	}

	c.Assert(s.client.blockMetaAccessor.PruneBlockMeta(200-BlockCacheSize), IsNil)
	allmetas, err = s.client.blockMetaAccessor.GetBlockMetas()
	c.Assert(err, IsNil)
	c.Assert(allmetas, HasLen, 148)
}
