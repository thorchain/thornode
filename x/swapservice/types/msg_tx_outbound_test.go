package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type MsgOutboundTxSuite struct{}

var _ = Suite(&MsgOutboundTxSuite{})

func (MsgOutboundTxSuite) TestMsgOutboundTx(c *C) {
	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	bnb, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)
	acc1, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	m := NewMsgOutboundTx(txID, 1, bnb, acc1)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_tx_outbound")

	inputs := []struct {
		txID   common.TxID
		height uint64
		sender common.BnbAddress
		signer sdk.AccAddress
	}{
		{
			txID:   common.TxID(""),
			height: 1,
			sender: bnb,
			signer: acc1,
		},
		{
			txID:   txID,
			height: 0,
			sender: bnb,
			signer: acc1,
		},
		{
			txID:   txID,
			height: 1,
			sender: common.NoBnbAddress,
			signer: acc1,
		},
		{
			txID:   txID,
			height: 1,
			sender: bnb,
			signer: sdk.AccAddress{},
		},
	}
	for _, item := range inputs {
		m := NewMsgOutboundTx(item.txID, item.height, item.sender, item.signer)
		c.Assert(m.ValidateBasic(), NotNil)
	}
}
