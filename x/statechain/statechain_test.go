package statechain

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/common"
	cmd "gitlab.com/thorchain/bepswap/statechain/cmd"
	stypes "gitlab.com/thorchain/bepswap/statechain/x/swapservice/types"
)

var mockedExitStatus = 0
var mockedStdout string

func TestPackage(t *testing.T) { TestingT(t) }

type StatechainSuite struct{}

var _ = Suite(&StatechainSuite{})

func fakeExecCommand(command string, args ...string) (cmd *exec.Cmd) {
	cs := []string{"-test.run=TestHelperCloneProcess", "--", command}
	cs = append(cs, args...)
	cmd = exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestHelperProcess(*testing.T) {
	fmt.Fprintf(os.Stdout, `{ "type": "cosmos-sdk/StdTx", "value": { "msg": [ { "type": "swapservice/MsgSetTxIn", "value": { "tx_hashes": [ { "request": "20D150DF19DAB33405D375982E479F48F607D0C9E4EE95B146F6C35FA2A09269", "status": "incomplete", "txhash": "", "memo": "This is my memo!", "coins": [ { "denom": "BNB", "amount": "1.234" } ], "sender": "bnb1ntqj0v0sv62ut0ehxt7jqh7lenfrd3hmfws0aq" } ], "signer": "rune1gnaghgzcpd73hcxeturml96maa0fajg9t8m0yj" } } ], "fee": { "amount": [], "gas": "200000" }, "signatures": [ { "pub_key": { "type": "tendermint/PubKeySecp256k1", "value": "A8FfMkUK6aNsD6F6tFAfjMd8FrivIp+TXYZETyvPUbSh" }, "signature": "8fwtZUvIWz63P5oLFMKnmoQCWBOTv2A96SRM4ITXgR52YalMjK3eMTemm947N0wqL/0OhXtrmhAPTHSSl/Q0sQ==" } ], "memo": "" } }`)
	os.Exit(0)
}

func (s StatechainSuite) TestSign(c *C) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(cmd.Bech32PrefixValAddr, cmd.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(cmd.Bech32PrefixConsAddr, cmd.Bech32PrefixConsPub)
	config.Seal()

	addr, err := sdk.AccAddressFromBech32("rune1gnaghgzcpd73hcxeturml96maa0fajg9t8m0yj")
	c.Assert(err, IsNil)

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

	signed, err := Sign([]stypes.TxIn{tx}, addr)
	c.Assert(err, IsNil)
	c.Check(signed.Value.Signatures[0].Signature, Equals, "8fwtZUvIWz63P5oLFMKnmoQCWBOTv2A96SRM4ITXgR52YalMjK3eMTemm947N0wqL/0OhXtrmhAPTHSSl/Q0sQ==")
}
