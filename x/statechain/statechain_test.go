package statechain

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/keys"
	cKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"gitlab.com/thorchain/bepswap/common"
	"gitlab.com/thorchain/bepswap/statechain/cmd"
	stypes "gitlab.com/thorchain/bepswap/statechain/x/swapservice/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/observe/config"
	"gitlab.com/thorchain/bepswap/observe/x/statechain/types"
)

func TestPackage(t *testing.T) { TestingT(t) }

type StatechainSuite struct{}

var _ = Suite(&StatechainSuite{})

func (*StatechainSuite) SetUpSuite(c *C) {
	cfg2 := sdk.GetConfig()
	cfg2.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	cfg2.Seal()
}

func setupStateChainForTest(c *C) (config.StateChainConfiguration, cKeys.Info, func()) {
	sscliDir := filepath.Join(os.TempDir(), ".sscli")
	cfg := config.StateChainConfiguration{
		ChainID:         "statechain",
		SignerName:      "bob",
		SignerPasswd:    "password",
		ChainHomeFolder: sscliDir,
	}
	kb, err := keys.NewKeyBaseFromDir(sscliDir)
	c.Assert(err, IsNil)
	info, _, err := kb.CreateMnemonic(cfg.SignerName, cKeys.English, cfg.SignerPasswd, cKeys.Secp256k1)
	c.Assert(err, IsNil)
	return cfg, info, func() {
		if err := os.RemoveAll(sscliDir); nil != err {
			c.Error(err)
		}
	}
}
func (s StatechainSuite) TestSign(c *C) {
	cfg, info, cleanup := setupStateChainForTest(c)
	defer cleanup()
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		c.Check(req.URL.String(), Equals, fmt.Sprintf("/auth/accounts/%s", info.GetAddress()))
		// Send response to be tested
		_, err := rw.Write([]byte(`{
			  "type": "cosmos-sdk/Account",
			  "value": {
				"address": "rune1v5n3r5j7hhvpdsdr4pkquqeq5x8plynnjgpc25",
				"coins": [
				  {
					"denom": "rune",
					"amount": "1000"
				  }
				],
				"public_key": {
				  "type": "tendermint/PubKeySecp256k1",
				  "value": "A8FfMkUK6aNsD6F6tFAfjMd8FrivIp+TXYZETyvPUbSh"
				},
				"account_number": "0",
				"sequence": "14"
			  }
			}`))
		c.Assert(err, IsNil)
	}))
	defer server.Close()
	u, err := url.Parse(server.URL)
	c.Assert(err, IsNil)
	cfg.ChainHost = u.Host
	tx := stypes.NewTxIn(
		common.TxID("20D150DF19DAB33405D375982E479F48F607D0C9E4EE95B146F6C35FA2A09269"),
		common.Coins{
			common.NewCoin(common.Ticker("BNB"), common.Amount("1.234")),
		},
		"This is my memo!",
		common.BnbAddress("bnb1ntqj0v0sv62ut0ehxt7jqh7lenfrd3hmfws0aq"),
	)
	bridge, err := NewStateChainBridge(cfg)
	c.Assert(err, IsNil)
	c.Assert(bridge, NotNil)
	signedMsg, err := bridge.Sign([]stypes.TxIn{tx})
	c.Assert(signedMsg, NotNil)
	c.Assert(err, IsNil)
}

func (s StatechainSuite) TestSend(c *C) {
	cfg, _, cleanup := setupStateChainForTest(c)
	defer cleanup()
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		c.Check(req.URL.String(), Equals, "/txs")
		// Send response to be tested
		_, err := rw.Write([]byte(`{"txhash":"E43FA2330C4317ECC084B0C6044DFE75AAE1FAB8F84A66107809E9739D02F80D", "height": "test_height"}`))
		c.Assert(err, IsNil)
	}))
	// Close the server when test finishes
	defer server.Close()

	u, err := url.Parse(server.URL)
	c.Assert(err, IsNil)
	cfg.ChainHost = u.Host
	bridge, err := NewStateChainBridge(cfg)
	c.Assert(err, IsNil)
	c.Assert(bridge, NotNil)
	stdTx := authtypes.StdTx{}
	mode := types.TxSync
	txID, err := bridge.Send(stdTx, mode)
	c.Assert(err, IsNil)
	c.Check(
		txID.String(),
		Equals,
		"E43FA2330C4317ECC084B0C6044DFE75AAE1FAB8F84A66107809E9739D02F80D",
	)
}
