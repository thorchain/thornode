package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type MsgSwapSuite struct{}

var _ = Suite(&MsgSwapSuite{})

func (MsgSwapSuite) TestMsgSwap(c *C) {
	addr, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	c.Check(addr.Empty(), Equals, false)
	bnbAddress, err := common.NewAddress("bnb1xlvns0n2mxh77mzaspn2hgav4rr4m8eerfju38")
	txID, err := common.NewTxID("712882AC9587198FA46F8D79BDFF013E77A89B12882702F03FA60FD298C517A4")
	c.Assert(err, IsNil)
	c.Check(txID.IsEmpty(), Equals, false)

	m := NewMsgSwap(txID, common.RuneA1FTicker, common.BNBTicker, sdk.NewUint(100000000), bnbAddress, bnbAddress, sdk.NewUint(200000000), addr)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_swap")

	inputs := []struct {
		requestTxHash common.TxID
		source        common.Ticker
		target        common.Ticker
		amount        sdk.Uint
		requester     common.Address
		destination   common.Address
		targetPrice   sdk.Uint
		signer        sdk.AccAddress
	}{
		{
			requestTxHash: common.TxID(""),
			source:        common.RuneA1FTicker,
			target:        common.BNBTicker,
			amount:        sdk.NewUint(100000000),
			requester:     bnbAddress,
			destination:   bnbAddress,
			targetPrice:   sdk.NewUint(200000000),
			signer:        addr,
		},
		{
			requestTxHash: txID,
			source:        common.Ticker(""),
			target:        common.BNBTicker,
			amount:        sdk.NewUint(100000000),
			requester:     bnbAddress,
			destination:   bnbAddress,
			targetPrice:   sdk.NewUint(200000000),
			signer:        addr,
		},
		{
			requestTxHash: txID,
			source:        common.BNBTicker,
			target:        common.BNBTicker,
			amount:        sdk.NewUint(100000000),
			requester:     bnbAddress,
			destination:   bnbAddress,
			targetPrice:   sdk.NewUint(200000000),
			signer:        addr,
		},
		{
			requestTxHash: txID,
			source:        common.RuneA1FTicker,
			target:        common.Ticker(""),
			amount:        sdk.NewUint(100000000),
			requester:     bnbAddress,
			destination:   bnbAddress,
			targetPrice:   sdk.NewUint(200000000),
			signer:        addr,
		},
		{
			requestTxHash: txID,
			source:        common.RuneA1FTicker,
			target:        common.BNBTicker,
			amount:        sdk.ZeroUint(),
			requester:     bnbAddress,
			destination:   bnbAddress,
			targetPrice:   sdk.NewUint(200000000),
			signer:        addr,
		},
		{
			requestTxHash: txID,
			source:        common.RuneA1FTicker,
			target:        common.BNBTicker,
			amount:        sdk.NewUint(100000000),
			requester:     common.NoAddress,
			destination:   bnbAddress,
			targetPrice:   sdk.NewUint(200000000),
			signer:        addr,
		},
		{
			requestTxHash: txID,
			source:        common.RuneA1FTicker,
			target:        common.BNBTicker,
			amount:        sdk.NewUint(100000000),
			requester:     bnbAddress,
			destination:   common.NoAddress,
			targetPrice:   sdk.NewUint(200000000),
			signer:        addr,
		},
		{
			requestTxHash: txID,
			source:        common.RuneA1FTicker,
			target:        common.BNBTicker,
			amount:        sdk.NewUint(100000000),
			requester:     bnbAddress,
			destination:   bnbAddress,
			targetPrice:   sdk.NewUint(200000000),
			signer:        sdk.AccAddress{},
		},
	}
	for _, item := range inputs {
		m := NewMsgSwap(item.requestTxHash, item.source, item.target, item.amount, item.requester, item.destination, item.targetPrice, item.signer)
		c.Assert(m.ValidateBasic(), NotNil)
	}
}
