package statechain

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/common"
	cmd "gitlab.com/thorchain/bepswap/statechain/cmd"
	stypes "gitlab.com/thorchain/bepswap/statechain/x/swapservice/types"
)

func TestPackage(t *testing.T) { TestingT(t) }

type StatechainSuite struct{}

var _ = Suite(&StatechainSuite{})

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

	signed, err := Sign([]stypes.TxIn{tx}, addr)
	c.Assert(err, IsNil)
	c.Check(signed.Value.Signatures[0].Signature, Equals, "8fwtZUvIWz63P5oLFMKnmoQCWBOTv2A96SRM4ITXgR52YalMjK3eMTemm947N0wqL/0OhXtrmhAPTHSSl/Q0sQ==")
}
