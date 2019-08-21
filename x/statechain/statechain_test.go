package statechain

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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
		ChainHost:       "localhost",
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

func (StatechainSuite) TestNewStateChainBridge(c *C) {
	var testFunc = func(cfg config.StateChainConfiguration, errChecker Checker, sbChecker Checker) {
		sb, err := NewStateChainBridge(cfg)
		c.Assert(err, errChecker)
		c.Assert(sb, sbChecker)
	}
	testFunc(config.StateChainConfiguration{
		ChainID:         "",
		ChainHost:       "localhost",
		ChainHomeFolder: "~/.sscli",
		SignerName:      "signer",
		SignerPasswd:    "signerpassword",
	}, NotNil, IsNil)
	testFunc(config.StateChainConfiguration{
		ChainID:         "chainid",
		ChainHost:       "",
		ChainHomeFolder: "~/.sscli",
		SignerName:      "signer",
		SignerPasswd:    "signerpassword",
	}, NotNil, IsNil)
	testFunc(config.StateChainConfiguration{
		ChainID:         "chainid",
		ChainHost:       "localhost",
		ChainHomeFolder: "~/.sscli",
		SignerName:      "",
		SignerPasswd:    "signerpassword",
	}, NotNil, IsNil)
	testFunc(config.StateChainConfiguration{
		ChainID:         "chainid",
		ChainHost:       "localhost",
		ChainHomeFolder: "~/.sscli",
		SignerName:      "signer",
		SignerPasswd:    "",
	}, NotNil, IsNil)
	cfg, _, cleanup := setupStateChainForTest(c)
	testFunc(cfg, IsNil, NotNil)
	defer cleanup()
}
func (StatechainSuite) TestGetAccountNumberAndSequenceNumber(c *C) {
	testfunc := func(handleFunc http.HandlerFunc, expectedAccNum int64, expectedSeq int64, errChecker Checker) {
		cfg, keyInfo, cleanup := setupStateChainForTest(c)
		defer cleanup()
		scb, err := NewStateChainBridge(cfg)
		c.Assert(err, IsNil)
		c.Assert(scb, NotNil)
		_ = keyInfo
		if nil != handleFunc {
			s := httptest.NewServer(handleFunc)
			defer s.Close()
			cfg.ChainHost = s.Listener.Addr().String()
		}
		requestUrl := scb.getAccountInfoUrl(cfg.ChainHost)
		if cfg.ChainHost == "localhost" {
			requestUrl = ""
		}
		accountNumber, seqNo, err := scb.getAccountNumberAndSequenceNumber(requestUrl)
		c.Assert(accountNumber, Equals, accountNumber)
		c.Assert(seqNo, Equals, seqNo)
		c.Assert(err, errChecker)
	}
	testfunc(nil, 0, 0, NotNil)
	testfunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	}, 0, 0, NotNil)
	testfunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	}, 0, 0, NotNil)
	testfunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusAccepted)
	}, 0, 0, NotNil)
	testfunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, err := writer.Write([]byte("whatever")); nil != err {
			c.Error(err)
		}
	}, 0, 0, NotNil)
	testfunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, err := writer.Write([]byte("")); nil != err {
			c.Error(err)
		}
	}, 0, 0, NotNil)
	testfunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, err := writer.Write([]byte(`{
"type": "cosmos-sdk/Account",
"value": {
"address": "",
"coins": [],
"public_key": null,
"account_number": "asdf",
"sequence": "0"
}
}`)); nil != err {
			c.Error(err)
		}
	}, 0, 0, NotNil)
	testfunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, err := writer.Write([]byte(`{
"type": "cosmos-sdk/Account",
"value": {
"address": "",
"coins": [],
"public_key": null,
"account_number": "0",
"sequence": "whatever"
}
}`)); nil != err {
			c.Error(err)
		}
	}, 0, 0, NotNil)
	testfunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, err := writer.Write([]byte(`{
"type": "cosmos-sdk/Account",
"value": {
"address": "",
"coins": [],
"public_key": null,
"account_number": "5",
"sequence": "6"
}
}`)); nil != err {
			c.Error(err)
		}
	}, 5, 6, IsNil)
}

func (StatechainSuite) TestSignEx(c *C) {
	testFunc := func(in []stypes.TxIn, handleFunc http.HandlerFunc, resultChecker Checker, errChecker Checker) {
		cfg, _, cleanup := setupStateChainForTest(c)
		defer cleanup()
		if nil != handleFunc {
			s := httptest.NewServer(handleFunc)
			defer s.Close()
			cfg.ChainHost = s.Listener.Addr().String()
		}
		scb, err := NewStateChainBridge(cfg)
		c.Assert(err, IsNil)
		c.Assert(scb, NotNil)
		stx, err := scb.Sign(in)
		c.Assert(stx, resultChecker)
		c.Assert(err, errChecker)
	}
	testFunc(nil, nil, IsNil, NotNil)
	testBNBAddress, err := common.NewBnbAddress("tbnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqx")
	if nil != err {
		c.Error(err)
	}
	testFunc([]stypes.TxIn{
		{
			Request: "EBB78FA6FDFBB19EBD188316B5FF9E60799C3149214A263274D31F4F605B8FDE",
			Status:  stypes.Incomplete,
			Done:    common.TxID(""),
			Memo:    "",
			Coins:   nil,
			Sender:  testBNBAddress,
		},
	}, func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	}, IsNil, NotNil)
}

func (StatechainSuite) TestSendEx(c *C) {
	testFunc := func(in []stypes.TxIn, mode types.TxMode, handleFunc http.HandlerFunc, resultChecker Checker, errChecker Checker) {
		cfg, _, cleanup := setupStateChainForTest(c)
		defer cleanup()
		if nil != handleFunc {
			s := httptest.NewServer(handleFunc)
			defer s.Close()
			cfg.ChainHost = s.Listener.Addr().String()
		}
		scb, err := NewStateChainBridge(cfg)
		c.Assert(err, IsNil)
		c.Assert(scb, NotNil)
		stx, err := scb.Sign(in)
		c.Assert(stx, NotNil)
		c.Assert(err, IsNil)
		_, err = scb.Send(*stx, mode)
		c.Assert(err, errChecker)

	}
	testBNBAddress, err := common.NewBnbAddress("tbnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqx")
	if nil != err {
		c.Error(err)
	}
	txIns := []stypes.TxIn{
		{
			Request: "EBB78FA6FDFBB19EBD188316B5FF9E60799C3149214A263274D31F4F605B8FDE",
			Status:  stypes.Incomplete,
			Done:    common.TxID(""),
			Memo:    "",
			Coins:   nil,
			Sender:  testBNBAddress,
		},
	}
	testFunc(txIns, types.TxUnknown, func(writer http.ResponseWriter, request *http.Request) {
		if _, err := writer.Write([]byte(`{
"type": "cosmos-sdk/Account",
"value": {
"address": "",
"coins": [],
"public_key": null,
"account_number": "5",
"sequence": "6"
}
}`)); nil != err {
			c.Error(err)
		}
	}, IsNil, NotNil)
	testFunc(txIns, types.TxSync, func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.RequestURI, "/auth/accounts") {
			if _, err := writer.Write([]byte(`{
"type": "cosmos-sdk/Account",
"value": {
"address": "",
"coins": [],
"public_key": null,
"account_number": "5",
"sequence": "6"
}
}`)); nil != err {
				c.Error(err)
			}
			return
		}
		writer.WriteHeader(http.StatusInternalServerError)
	}, IsNil, NotNil)
	testFunc(txIns, types.TxSync, func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.RequestURI, "/auth/accounts") {
			if _, err := writer.Write([]byte(`{
"type": "cosmos-sdk/Account",
"value": {
"address": "",
"coins": [],
"public_key": null,
"account_number": "5",
"sequence": "6"
}
}`)); nil != err {
				c.Error(err)
			}
			return
		}

		if _, err := writer.Write([]byte(`
whatever`)); nil != err {
			c.Error(err)
		}

	}, IsNil, NotNil)

}
