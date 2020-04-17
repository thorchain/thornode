package signer

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	. "gopkg.in/check.v1"

	ctypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/cosmos/cosmos-sdk/client/keys"
	cKeys "github.com/cosmos/cosmos-sdk/crypto/keys"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/pubkeymanager"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/x/thorchain"
	types2 "gitlab.com/thorchain/thornode/x/thorchain/types"
)

func Test(t *testing.T) { TestingT(t) }

type ThorchainBlockScanSuite struct {
	thordir  string
	thorKeys *thorclient.Keys
	bridge   *thorclient.ThorchainBridge
	m        *metrics.Metrics
	storage  *SignerStore
	rpcHost  string
}

var _ = Suite(&ThorchainBlockScanSuite{})

func (s *ThorchainBlockScanSuite) SetUpSuite(c *C) {
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
			_, err := rw.Write([]byte(`{ "jsonrpc": "2.0", "id": "", "result": { "height": "1", "txhash": "ENULZOBGZHEKFOIBYRLLBELKFZVGXOBLTRQGTOWNDHMPZQMBLGJETOXJLHPVQIKY", "logs": [{"success": "true", "log": ""}] } }`))
			c.Assert(err, IsNil)
		} else if strings.HasPrefix(req.RequestURI, "/thorchain/lastblock/BNB") {
			_, err := rw.Write([]byte(`{ "jsonrpc": "2.0", "id": "", "result": { "chain": "BNB", "lastobservedin": "1", "lastsignedout": "1", "statechain": "1" } }`))
			c.Assert(err, IsNil)
		} else if strings.HasPrefix(req.RequestURI, "/thorchain/lastblock") {
			_, err := rw.Write([]byte(`{ "jsonrpc": "2.0", "id": "", "result": { "chain": "ThorChain", "lastobservedin": "1", "lastsignedout": "1", "statechain": "1" } }`))
			c.Assert(err, IsNil)
		} else if strings.HasPrefix(req.RequestURI, "/auth/accounts/") {
			_, err := rw.Write([]byte(`{ "jsonrpc": "2.0", "id": "", "result": { "height": "1", "result": { "value": { "account_number": "0", "sequence": "0" } } } |`))
			c.Assert(err, IsNil)
		} else if strings.HasPrefix(req.RequestURI, "/thorchain/vaults/pubkeys") {
			_, err := rw.Write([]byte(`{ "jsonrpc": "2.0", "id": "", "result": { "asgard": ["thorpub1addwnpepqfgfxharps79pqv8fv9ndqh90smw8c3slrtrssn58ryc5g3p9sx856x07yn"], "yggdrasil": ["thorpub1addwnpepqdqvd4r84lq9m54m5kk9sf4k6kdgavvch723pcgadulxd6ey9u70kgjgrwl"] } }`))
			c.Assert(err, IsNil)
		} else if req.RequestURI == "/block" {
			_, err := rw.Write([]byte(`{ "jsonrpc": "2.0", "id": "", "result": { "block": { "header": { "height": "1" } } } }`))
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
		} else if strings.HasPrefix(req.RequestURI, "/thorchain/keygen") {
			_, err := rw.Write([]byte(`{
		"height": "1",
		"keygens": [
		{
			"id": "AAAA000000000000000000000000000000000000000000000000000000000000",
			"type": "asgard",
			"members": [
				"thorpub1addwnpepqfgfxharps79pqv8fv9ndqh90smw8c3slrtrssn58ryc5g3p9sx856x07yn"
			]
		}]}`))
			c.Assert(err, IsNil)
		} else if strings.HasSuffix(req.RequestURI, "/signers") {
			_, err := rw.Write([]byte(`[
				"thorpub1addwnpepqfgfxharps79pqv8fv9ndqh90smw8c3slrtrssn58ryc5g3p9sx856x07yn"
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

func (s *ThorchainBlockScanSuite) TearDownSuite(c *C) {
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

func (s *ThorchainBlockScanSuite) TestProcess(c *C) {
	cfg := config.BlockScannerConfiguration{
		RPCHost:                    "127.0.0.1:" + s.rpcHost,
		ChainID:                    "ThorChain",
		StartBlockHeight:           1,
		EnforceBlockHeight:         true,
		BlockScanProcessors:        1,
		BlockHeightDiscoverBackoff: time.Second,
		BlockRetryInterval:         10 * time.Second,
	}
	blockScan, err := NewThorchainBlockScan(cfg, s.storage, s.bridge, s.m, pubkeymanager.NewMockPoolAddressValidator())
	c.Assert(blockScan, NotNil)
	c.Assert(err, IsNil)
}
