package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/statechain/cmd"
)

type MsgAddSuite struct{}

var _ = Suite(&MsgAddSuite{})

func (mas *MsgAddSuite) SetUpSuite(c *C) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(cmd.Bech32PrefixValAddr, cmd.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(cmd.Bech32PrefixConsAddr, cmd.Bech32PrefixConsPub)
	config.Seal()
}
func (mas *MsgAddSuite) TestMsgAdd(c *C) {

	txId, err := common.NewTxID("712882AC9587198FA46F8D79BDFF013E77A89B12882702F03FA60FD298C517A4")
	c.Assert(err, IsNil)
	c.Check(txId.IsEmpty(), Equals, false)
	addr, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	c.Check(addr.Empty(), Equals, false)
	ma := NewMsgAdd(common.BNBTicker, common.NewAmountFromFloat(1), common.NewAmountFromFloat(1), txId, addr)
	c.Check(ma.Route(), Equals, RouterKey)
	c.Check(ma.Type(), Equals, "set_add")
	err = ma.ValidateBasic()
	c.Assert(err, IsNil)
	buf := ma.GetSignBytes()
	c.Assert(buf, NotNil)
	c.Check(len(buf) > 0, Equals, true)
	signer := ma.GetSigners()
	c.Assert(signer, NotNil)
	c.Check(len(signer) > 0, Equals, true)

	inputs := []struct {
		ticker common.Ticker
		rune   common.Amount
		token  common.Amount
		txHash common.TxID
		signer sdk.AccAddress
	}{
		{
			ticker: common.Ticker(""),
			rune:   common.NewAmountFromFloat(1),
			token:  common.NewAmountFromFloat(1),
			txHash: txId,
			signer: addr,
		},
		{
			ticker: common.BNBTicker,
			rune:   common.Amount(""),
			token:  common.NewAmountFromFloat(1),
			txHash: txId,
			signer: addr,
		},
		{
			ticker: common.BNBTicker,
			rune:   common.NewAmountFromFloat(1),
			token:  common.Amount(""),
			txHash: txId,
			signer: addr,
		},
		{
			ticker: common.BNBTicker,
			rune:   common.NewAmountFromFloat(1),
			token:  common.NewAmountFromFloat(1),
			txHash: common.TxID(""),
			signer: addr,
		},
		{
			ticker: common.BNBTicker,
			rune:   common.NewAmountFromFloat(1),
			token:  common.NewAmountFromFloat(1),
			txHash: txId,
			signer: sdk.AccAddress{},
		},
	}
	for _, item := range inputs {
		msgAdd := NewMsgAdd(item.ticker, item.rune, item.token, item.txHash, item.signer)
		err := msgAdd.ValidateBasic()
		c.Assert(err, NotNil)
	}
}
