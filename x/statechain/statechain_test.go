package statechain

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"gitlab.com/thorchain/bepswap/common"
	config "gitlab.com/thorchain/bepswap/observe/config"
	"gitlab.com/thorchain/bepswap/observe/x/statechain/types"
	cmd "gitlab.com/thorchain/bepswap/statechain/cmd"
	stypes "gitlab.com/thorchain/bepswap/statechain/x/swapservice/types"
)

var testCase = "case1"

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	tc := "TEST_CASE=" + testCase
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", tc}
	return cmd
}
func getTestConfiguration() config.Configuration {
	return config.Configuration{
		PoolAddress:      "pooladdress",
		RuneAddress:      "runeaddress",
		DEXHost:          "dexhost",
		RPCHost:          "rpchost",
		PrivateKey:       "privatekey",
		ChainHost:        "chainhost",
		SignerName:       "johnny",
		SignerPasswd:     "johnnysupersecurepassword",
		ObserverDbPath:   "",
		SignerDbPath:     "",
		SocketPoing:      30,
		MessageProcessor: 10,
	}
}
func TestSign(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(cmd.Bech32PrefixValAddr, cmd.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(cmd.Bech32PrefixConsAddr, cmd.Bech32PrefixConsPub)
	config.Seal()

	addr, err := sdk.AccAddressFromBech32("rune1gnaghgzcpd73hcxeturml96maa0fajg9t8m0yj")
	assert.Equal(t, err, nil)
	// c.Assert(err, IsNil)

	tx := stypes.NewTxIn(
		common.TxID("20D150DF19DAB33405D375982E479F48F607D0C9E4EE95B146F6C35FA2A09269"),
		common.Coins{
			common.NewCoin(common.Ticker("BNB"), common.Amount("1.234")),
		},
		"This is my memo!",
		common.BnbAddress("bnb1ntqj0v0sv62ut0ehxt7jqh7lenfrd3hmfws0aq"),
	)

	// use fake execCommand
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }() // set back to real one.

	signed, err := Sign([]stypes.TxIn{tx}, addr, getTestConfiguration())
	assert.Equal(t, err, nil)
	assert.Equal(t, signed.Value.Signatures[0].Signature, "8fwtZUvIWz63P5oLFMKnmoQCWBOTv2A96SRM4ITXgR52YalMjK3eMTemm947N0wqL/0OhXtrmhAPTHSSl/Q0sQ==")
}

func TestSend(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		assert.Equal(t, req.URL.String(), "/txs")
		// Send response to be tested
		_, err := rw.Write([]byte(`{"txhash":"E43FA2330C4317ECC084B0C6044DFE75AAE1FAB8F84A66107809E9739D02F80D", "height": "test_height"}`))
		assert.Equal(t, err, nil)
	}))
	// Close the server when test finishes
	defer server.Close()

	u, err := url.Parse(server.URL)
	assert.Equal(t, err, nil)
	config.ChainHost = u.Host

	stdTx := types.StdTx{}
	mode := types.TxSync

	txID, err := Send(stdTx, mode)
	assert.Equal(t, err, nil)
	assert.Equal(t, txID.String(), "E43FA2330C4317ECC084B0C6044DFE75AAE1FAB8F84A66107809E9739D02F80D")
}

func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	switch os.Getenv("TEST_CASE") {
	case "case1":
		fmt.Fprintf(os.Stdout, `{ "type": "cosmos-sdk/StdTx", "value": { "msg": [ { "type": "swapservice/MsgSetTxIn", "value": { "tx_hashes": [ { "request": "20D150DF19DAB33405D375982E479F48F607D0C9E4EE95B146F6C35FA2A09269", "status": "incomplete", "txhash": "", "memo": "This is my memo!", "coins": [ { "denom": "BNB", "amount": "1.234" } ], "sender": "bnb1ntqj0v0sv62ut0ehxt7jqh7lenfrd3hmfws0aq" } ], "signer": "rune1gnaghgzcpd73hcxeturml96maa0fajg9t8m0yj" } } ], "fee": { "amount": [], "gas": "200000" }, "signatures": [ { "pub_key": { "type": "tendermint/PubKeySecp256k1", "value": "A8FfMkUK6aNsD6F6tFAfjMd8FrivIp+TXYZETyvPUbSh" }, "signature": "8fwtZUvIWz63P5oLFMKnmoQCWBOTv2A96SRM4ITXgR52YalMjK3eMTemm947N0wqL/0OhXtrmhAPTHSSl/Q0sQ==" } ], "memo": "" } }`)
	}
	os.Exit(0)
}
