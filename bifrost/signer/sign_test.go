package signer

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
	"sync"

	ctypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/cosmos/cosmos-sdk/client/keys"
	cKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/pkg/chainclients"
	"gitlab.com/thorchain/thornode/bifrost/pkg/chainclients/binance"
	pubkeymanager "gitlab.com/thorchain/thornode/bifrost/pubkeymanager"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	ttypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/x/thorchain"
	types2 "gitlab.com/thorchain/thornode/x/thorchain/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

func TestPackage(t *testing.T) { TestingT(t) }

var m *metrics.Metrics

func GetMetricForTest(c *C) *metrics.Metrics {
	if m == nil {
		var err error
		m, err = metrics.NewMetrics(config.MetricsConfiguration{
			Enabled:      false,
			ListenPort:   9000,
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
			Chains:       common.Chains{common.BNBChain},
		})
		c.Assert(m, NotNil)
		c.Assert(err, IsNil)
	}
	return m
}

type SignSuite struct{
	thordir  string
	thorKeys *thorclient.Keys
	bridge   *thorclient.ThorchainBridge
	m        *metrics.Metrics
	rpcHost  string
	storage  *SignerStore
}

var _ = Suite(&SignSuite{})

type MockCheckTransactionChain struct {
	chainclients.DummyChain
	validateMetaData bool
}

func (m *MockCheckTransactionChain) ValidateMetadata(_ interface{}) bool {
	return m.validateMetaData
}

func (s *SignSuite) SetUpSuite(c *C) {
	thorchain.SetupConfigForTest()
	s.m = GetMetricForTest(c)
	c.Assert(s.m, NotNil)
	ns := strconv.Itoa(time.Now().Nanosecond())
	types2.SetupConfigForTest()
	ctypes.Network = ctypes.TestNetwork
	c.Assert(os.Setenv("NET", "testnet"), IsNil)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		c.Logf("requestUri:%s", req.RequestURI)
		if strings.HasPrefix(req.RequestURI, "/txs") {
			_, err := rw.Write([]byte(`{ "jsonrpc": "2.0", "id": "", "result": { "height": "1", "txhash": "AAAA000000000000000000000000000000000000000000000000000000000000", "logs": [{"success": "true", "log": ""}] } }`))
			c.Assert(err, IsNil)
		} else if strings.HasPrefix(req.RequestURI, "/thorchain/lastblock/BNB") {
			_, err := rw.Write([]byte(`{ "jsonrpc": "2.0", "id": "", "result": { "chain": "BNB", "lastobservedin": "0", "lastsignedout": "0", "statechain": "0" } }`))
			c.Assert(err, IsNil)
		} else if strings.HasPrefix(req.RequestURI, "/thorchain/lastblock") {
			_, err := rw.Write([]byte(`{ "jsonrpc": "2.0", "id": "", "result": { "chain": "ThorChain", "lastobservedin": "0", "lastsignedout": "0", "statechain": "0" } }`))
			c.Assert(err, IsNil)
		} else if strings.HasPrefix(req.RequestURI, "/auth/accounts/") {
			_, err := rw.Write([]byte(`{ "jsonrpc": "2.0", "id": "", "result": { "height": "0", "result": { "value": { "account_number": "0", "sequence": "0" } } } |`))
			c.Assert(err, IsNil)
		} else if strings.HasPrefix(req.RequestURI, "/thorchain/vaults/pubkeys") {
			_, err := rw.Write([]byte(`{ "jsonrpc": "2.0", "id": "", "result": { "asgard": ["thorpub1addwnpepq2jgpsw2lalzuk7sgtmyakj7l6890f5cfpwjyfp8k4y4t7cw2vk8vcglsjy"], "yggdrasil": ["thorpub1addwnpepqdqvd4r84lq9m54m5kk9sf4k6kdgavvch723pcgadulxd6ey9u70kgjgrwl"] } }`))
			c.Assert(err, IsNil)
		} else if strings.HasPrefix(req.RequestURI, "/thorchain/keysign") {
			_, err := rw.Write([]byte(`{
			"chains": {
				"BNB": {
					"chain": "BNB",
					"hash": "",
					"height": "1",
					"tx_array": [
						{
							"chain": "BNB",
							"coin": {
								"amount": "10000000000",
								"asset": "BNB.BNB"
							},
							"in_hash": "ENULZOBGZHEKFOIBYRLLBELKFZVGXOBLTRQGTOWNDHMPZQMBLGJETOXJLHPVQIKY",
							"memo": "",
							"out_hash": "",
							"to": "tbnb145wcuncewfkuc4v6an0r9laswejygcul43c3wu",
							"vault_pubkey": "thorpub1addwnpepqfgfxharps79pqv8fv9ndqh90smw8c3slrtrssn58ryc5g3p9sx856x07yn"
						}
					]
				}
			}
		}
	`))
			c.Assert(err, IsNil)
		} else if strings.HasSuffix(req.RequestURI, "/signers") {
			_, err := rw.Write([]byte(`[
  "thorpub1addwnpepqflvfv08t6qt95lmttd6wpf3ss8wx63e9vf6fvyuj2yy6nnyna5763e2kck",
  "thorpub1addwnpepq2flfr96skc5lkwdv0n5xjsnhmuju20x3zndgu42zd8dtkrud9m2v0zl2qu",
  "thorpub1addwnpepqwhnus6xs4208d4ynm05lv493amz3fexfjfx4vptntedd7k0ajlcup0pzgk"
]`))
			c.Assert(err, IsNil)
		} else {
		}
	}))

	s.thordir = filepath.Join(os.TempDir(), ns, ".thorcli")
	splitted := strings.SplitAfter(server.URL, ":")
	s.rpcHost = splitted[len(splitted)-1]
	cfg := config.ClientConfiguration{
		ChainID:         "thorchain",
		ChainHost:       "localhost:" + s.rpcHost,
		SignerName:      "bob",
		SignerPasswd:    "password",
		ChainHomeFolder: s.thordir,
	}

	kb, err := keys.NewKeyBaseFromDir(s.thordir)
	c.Assert(err, IsNil)
	_, _, err = kb.CreateMnemonic(cfg.SignerName, cKeys.English, cfg.SignerPasswd, cKeys.Secp256k1)
	c.Assert(err, IsNil)
	s.thorKeys, err = thorclient.NewKeys(cfg.ChainHomeFolder, cfg.SignerName, cfg.SignerPasswd)
	c.Assert(err, IsNil)
	s.bridge, err = thorclient.NewThorchainBridge(cfg, s.m)
	c.Assert(err, IsNil)
	s.storage, err = NewSignerStore("signer_data", "")
	c.Assert(err, IsNil)
}

func (s *SignSuite) TearDownSuite(c *C) {
	c.Assert(os.Unsetenv("NET"), IsNil)

	if err := os.RemoveAll(s.thordir); err != nil {
		c.Error(err)
	}

	if err := os.RemoveAll("signer_data"); err != nil {
		c.Error(err)
	}
	tempPath := filepath.Join(os.TempDir(), "/var/data/bifrost/signer")
	if err := os.RemoveAll(tempPath); err != nil {
		c.Error(err)
	}

	if err := os.RemoveAll("signer/var"); err != nil {
		c.Error(err)
	}
}

func (s *SignSuite) TestCheckTxn(c *C) {
	storage, err := NewSignerStore("", "")
	c.Assert(err, IsNil)

	mockChain := &MockCheckTransactionChain{
		validateMetaData: true,
	}
	chain, err := common.NewChain("MOCK")
	c.Assert(err, IsNil)

	chains := make(map[common.Chain]chainclients.ChainClient)
	chains[chain] = mockChain

	signer := &Signer{
		chains:  chains,
		storage: storage,
	}

	status, err := signer.CheckTransaction("", "bad chain", nil)
	c.Assert(err, NotNil)
	c.Check(status, Equals, TxUnknown)

	status, err = signer.CheckTransaction("", chain, nil)
	c.Assert(err, IsNil)
	c.Check(status, Equals, TxUnavailable)

	tx := NewTxOutStoreItem(12, ttypes.TxOutItem{Memo: "foo"})
	c.Assert(storage.Set(tx), IsNil)

	status, err = signer.CheckTransaction(tx.Key(), chain, nil)
	c.Assert(err, IsNil)
	c.Check(status, Equals, TxAvailable)

	spent := NewTxOutStoreItem(100, ttypes.TxOutItem{Memo: "spent"})
	spent.Status = TxSpent
	c.Assert(storage.Set(spent), IsNil)

	status, err = signer.CheckTransaction(spent.Key(), chain, nil)
	c.Assert(err, IsNil)
	c.Check(status, Equals, TxSpent)
}

type MockChainClient struct {
	account common.Account
}

func (b *MockChainClient) SignTx(tai stypes.TxOutItem, height int64) ([]byte, error) {
	return nil, nil
}

func (b *MockChainClient) GetHeight() (int64, error) {
	return 0, nil
}

func (b *MockChainClient) GetGasFee(count uint64) common.Gas {
	return common.GetBNBGasFee(count)
}

func (b *MockChainClient) CheckIsTestNet() (string, bool) {
	return "", true
}

func (b *MockChainClient) GetChain() common.Chain {
	return common.BNBChain
}

func (b *MockChainClient) ValidateMetadata(inter interface{}) bool {
	return true
}

func (b *MockChainClient) BroadcastTx(tx []byte) error {
	return nil
}

func (b *MockChainClient) GetAddress(poolPubKey common.PubKey) string {
	return "0dd3d0a4a6eacc98cc4894791702e46c270bde76"
}

func (b *MockChainClient) GetAccount(addr string) (common.Account, error) {
	return b.account, nil
}

func (b *MockChainClient) GetPubKey() crypto.PubKey {
	return nil
}

func (b *MockChainClient) Start(globalTxsQueue chan stypes.TxIn, pubkeyMgr pubkeymanager.PubKeyValidator, m *metrics.Metrics) error {
	return nil
}

func (b *MockChainClient) Stop() error {
	return nil
}

func (s *SignSuite) TestHandleYggReturn_Success_FeeSingleton(c *C) {
	sign := &Signer{
		chains: map[common.Chain]chainclients.ChainClient{
			common.BNBChain: &MockChainClient{
				account: common.Account{
					Coins: common.AccountCoins{
						common.AccountCoin{Denom: common.BNBChain.String(), Amount: 1000000},
					},
				},
			},
		},
	}
	input := `{ "chain": "BNB", "memo": "yggdrasil-", "to": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj", "coins": [] }`
	var item stypes.TxOutItem
	err := json.Unmarshal([]byte(input), &item)
	c.Check(err, IsNil)

	newItem, err := sign.handleYggReturn(item)
	c.Assert(err, IsNil)
	c.Check(newItem.Coins[0].Amount.Uint64(), Equals, uint64(962500))
}

func (s *SignSuite) TestHandleYggReturn_Success_FeeMulti(c *C) {
	sign := &Signer{
		chains: map[common.Chain]chainclients.ChainClient{
			common.BNBChain: &MockChainClient{
				account: common.Account{
					Coins: common.AccountCoins{
						common.AccountCoin{Denom: common.BNBChain.String(), Amount: 1000000},
						common.AccountCoin{Denom: "RUNE", Amount: 1000000},
					},
				},
			},
		},
	}
	input := `{ "chain": "BNB", "memo": "yggdrasil-", "to": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj", "coins": [] }`
	var item stypes.TxOutItem
	err := json.Unmarshal([]byte(input), &item)
	c.Check(err, IsNil)

	newItem, err := sign.handleYggReturn(item)
	c.Assert(err, IsNil)
	c.Check(newItem.Coins[0].Amount.Uint64(), Equals, uint64(940000))
}

func (s *SignSuite) TestHandleYggReturn_Success_NotEnough(c *C) {
	sign := &Signer{
		chains: map[common.Chain]chainclients.ChainClient{
			common.BNBChain: &MockChainClient{
				account: common.Account{
					Coins: common.AccountCoins{
						common.AccountCoin{Denom: common.BNBChain.String(), Amount: 0},
					},
				},
			},
		},
	}
	input := `{ "chain": "BNB", "memo": "yggdrasil-", "to": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj", "coins": [] }`
	var item stypes.TxOutItem
	err := json.Unmarshal([]byte(input), &item)
	c.Check(err, IsNil)

	newItem, err := sign.handleYggReturn(item)
	c.Assert(err, IsNil)
	c.Check(newItem.Coins[0].Amount.Uint64(), Equals, uint64(0))
}

func (s *SignSuite) TestProcess(c *C) {
	cfg := config.SignerConfiguration{
		SignerDbPath: filepath.Join(os.TempDir(), "/var/data/bifrost/signer"),
		BlockScanner: config.BlockScannerConfiguration{
			RPCHost: "127.0.0.1:" + s.rpcHost,
			ChainID: "ThorChain",
			StartBlockHeight: 1,
			EnforceBlockHeight: true,
			BlockScanProcessors: 1,
			BlockHeightDiscoverBackoff: time.Second,
			BlockRetryInterval: 10 * time.Second,
		},
		RetryInterval: 2 * time.Second,
	}

	chains := map[common.Chain]chainclients.ChainClient{
		common.BNBChain: &MockChainClient{
			account: common.Account{
				Coins: common.AccountCoins{
					common.AccountCoin{Denom: common.BNBChain.String(), Amount: 1000000},
					common.AccountCoin{Denom: "RUNE", Amount: 1000000},
				},
			},
		},
	}
	blockScan, err := NewThorchainBlockScan(cfg.BlockScanner, s.storage, s.bridge, s.m, binance.NewMockPoolAddressValidator())
	c.Assert(err, IsNil)
	sign := &Signer{
		logger:                log.With().Str("module", "signer").Logger(),
		cfg:                   cfg,
		wg:                    &sync.WaitGroup{},
		stopChan:              make(chan struct{}),
		thorchainBlockScanner: blockScan,
		chains:                chains,
		m:                     s.m,
		storage:               s.storage,
		errCounter:            s.m.GetCounterVec(metrics.SignerError),
		pubkeyMgr:             binance.NewMockPoolAddressValidator(),
		thorchainBridge:       s.bridge,
	}
	c.Assert(sign, NotNil)
	err = sign.Start()
	c.Assert(err, IsNil)
	time.Sleep(time.Second * 2)
	go sign.Stop()
}
