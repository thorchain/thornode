package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type MsgStakeSuite struct{}

var _ = Suite(&MsgStakeSuite{})

func (MsgStakeSuite) TestMsgStake(c *C) {
	addr, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	c.Check(addr.Empty(), Equals, false)
	bnbAddress, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	txID, err := common.NewTxID("712882AC9587198FA46F8D79BDFF013E77A89B12882702F03FA60FD298C517A4")
	c.Assert(err, IsNil)
	c.Check(txID.IsEmpty(), Equals, false)
	m := NewMsgSetStakeData(common.BNBTicker, common.NewAmountFromFloat(1), common.NewAmountFromFloat(1), bnbAddress, txID, addr)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_stakedata")

	inputs := []struct {
		ticker        common.Ticker
		r             common.Amount
		token         common.Amount
		publicAddress common.BnbAddress
		txHash        common.TxID
		signer        sdk.AccAddress
	}{
		{
			ticker:        common.Ticker(""),
			r:             common.NewAmountFromFloat(1),
			token:         common.NewAmountFromFloat(1),
			publicAddress: bnbAddress,
			txHash:        txID,
			signer:        addr,
		},
		{
			ticker:        common.BNBTicker,
			r:             common.Amount(""),
			token:         common.NewAmountFromFloat(1),
			publicAddress: bnbAddress,
			txHash:        txID,
			signer:        addr,
		},
		{
			ticker:        common.BNBTicker,
			r:             common.NewAmountFromFloat(1),
			token:         common.Amount(""),
			publicAddress: bnbAddress,
			txHash:        txID,
			signer:        addr,
		},
		{
			ticker:        common.BNBTicker,
			r:             common.NewAmountFromFloat(1),
			token:         common.NewAmountFromFloat(1),
			publicAddress: common.NoBnbAddress,
			txHash:        txID,
			signer:        addr,
		},
		{
			ticker:        common.BNBTicker,
			r:             common.NewAmountFromFloat(1),
			token:         common.NewAmountFromFloat(1),
			publicAddress: bnbAddress,
			txHash:        common.TxID(""),
			signer:        addr,
		},
		{
			ticker:        common.BNBTicker,
			r:             common.NewAmountFromFloat(1),
			token:         common.NewAmountFromFloat(1),
			publicAddress: bnbAddress,
			txHash:        txID,
			signer:        sdk.AccAddress{},
		},
	}
	for _, item := range inputs {
		m := NewMsgSetStakeData(item.ticker, item.r, item.token, item.publicAddress, item.txHash, item.signer)
		c.Assert(m.ValidateBasic(), NotNil)
	}
}

func EnsureMsgBasicCorrect(m sdk.Msg, c *C) {
	signers := m.GetSigners()
	c.Check(signers, NotNil)
	c.Check(len(signers), Equals, 1)
	c.Check(m.ValidateBasic(), IsNil)
	c.Check(m.Route(), Equals, RouterKey)
	c.Check(m.GetSignBytes(), NotNil)
}
