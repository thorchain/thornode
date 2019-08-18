package statechain

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/user"
	"path/filepath"
	"testing"

	. "gopkg.in/check.v1"

	"github.com/cosmos/cosmos-sdk/client/keys"
	cKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"gitlab.com/thorchain/bepswap/common"
	"gitlab.com/thorchain/bepswap/statechain/cmd"
	stypes "gitlab.com/thorchain/bepswap/statechain/x/swapservice/types"

	"gitlab.com/thorchain/bepswap/observe/config"
	"gitlab.com/thorchain/bepswap/observe/x/statechain/types"
)

func TestPackage(t *testing.T) { TestingT(t) }

type StatechainSuite struct{}

var _ = Suite(&StatechainSuite{})

func (s StatechainSuite) TestSign(c *C) {
	// create a user in our keybase
	usr, err := user.Current()
	c.Assert(err, IsNil)
	sscliDir := filepath.Join(usr.HomeDir, ".sscli")
	kb, err := keys.NewKeyBaseFromDir(sscliDir)
	c.Assert(err, IsNil)

	config.SignerName = "bob"
	config.SignerPasswd = "password"
	info, _, err := kb.CreateMnemonic(config.SignerName, cKeys.English, config.SignerPasswd, cKeys.Secp256k1)
	c.Assert(err, IsNil)
	i, err := kb.Get("bob")
	c.Assert(err, IsNil, Commentf("Info: %+v", i))

	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	cfg.Seal()

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
	config.ChainHost = u.Host

	tx := stypes.NewTxIn(
		common.TxID("20D150DF19DAB33405D375982E479F48F607D0C9E4EE95B146F6C35FA2A09269"),
		common.Coins{
			common.NewCoin(common.Ticker("BNB"), common.Amount("1.234")),
		},
		"This is my memo!",
		common.BnbAddress("bnb1ntqj0v0sv62ut0ehxt7jqh7lenfrd3hmfws0aq"),
	)

	_, err = Sign([]stypes.TxIn{tx}, info.GetAddress())
	// bz, _ := json.Marshal(signed)
	c.Assert(err, IsNil)
	/*
		// This is commented out because each time this runs in CI, it creates a
		// new user with a different resulting signature. We can figure out a way
		// later to verify signature is correct.
		c.Check(
			b64.StdEncoding.EncodeToString(signed.Signatures[0].Signature),
			Equals,
			"8fwtZUvIWz63P5oLFMKnmoQCWBOTv2A96SRM4ITXgR52YalMjK3eMTemm947N0wqL/0OhXtrmhAPTHSSl/Q0sQ==",
		)
	*/
}

func (s StatechainSuite) TestSend(c *C) {
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
	config.ChainHost = u.Host

	stdTx := authtypes.StdTx{}
	mode := types.TxSync

	txID, err := Send(stdTx, mode)
	c.Assert(err, IsNil)
	c.Check(
		txID.String(),
		Equals,
		"E43FA2330C4317ECC084B0C6044DFE75AAE1FAB8F84A66107809E9739D02F80D",
	)
}
