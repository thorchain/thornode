package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type MsgUnstakeSuite struct{}

var _ = Suite(&MsgUnstakeSuite{})

func (MsgUnstakeSuite) TestMsgUnstake(c *C) {
	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	bnb, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)
	acc1, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	m := NewMsgSetUnStake(bnb, sdk.NewUint(10000), common.BNBTicker, txID, acc1)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_unstake")

	inputs := []struct {
		publicAddress       common.BnbAddress
		withdrawBasisPoints sdk.Uint
		ticker              common.Ticker
		requestTxHash       common.TxID
		signer              sdk.AccAddress
	}{
		{
			publicAddress:       common.NoBnbAddress,
			withdrawBasisPoints: sdk.NewUint(10000),
			ticker:              common.BNBTicker,
			requestTxHash:       txID,
			signer:              acc1,
		},
		{
			publicAddress:       bnb,
			withdrawBasisPoints: sdk.NewUint(12000),
			ticker:              common.BNBTicker,
			requestTxHash:       txID,
			signer:              acc1,
		},
		{
			publicAddress:       bnb,
			withdrawBasisPoints: sdk.ZeroUint(),
			ticker:              common.BNBTicker,
			requestTxHash:       txID,
			signer:              acc1,
		},
		{
			publicAddress:       bnb,
			withdrawBasisPoints: sdk.NewUint(10000),
			ticker:              common.Ticker(""),
			requestTxHash:       txID,
			signer:              acc1,
		},
		{
			publicAddress:       bnb,
			withdrawBasisPoints: sdk.NewUint(10000),
			ticker:              common.BNBTicker,
			requestTxHash:       common.TxID(""),
			signer:              acc1,
		},
		{
			publicAddress:       bnb,
			withdrawBasisPoints: sdk.NewUint(10000),
			ticker:              common.BNBTicker,
			requestTxHash:       txID,
			signer:              sdk.AccAddress{},
		},
	}
	for _, item := range inputs {
		m := NewMsgSetUnStake(item.publicAddress, item.withdrawBasisPoints, item.ticker, item.requestTxHash, item.signer)
		c.Assert(m.ValidateBasic(), NotNil)
	}
}
