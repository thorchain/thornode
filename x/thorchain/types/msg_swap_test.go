package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"
)

type MsgSwapSuite struct{}

var _ = Suite(&MsgSwapSuite{})

func (MsgSwapSuite) TestMsgSwap(c *C) {
	addr := GetRandomBech32Addr()
	c.Check(addr.Empty(), Equals, false)
	bnbAddress := GetRandomBNBAddress()
	txID := GetRandomTxHash()
	c.Check(txID.IsEmpty(), Equals, false)

	tx := common.NewTx(
		txID,
		GetRandomBNBAddress(),
		GetRandomBNBAddress(),
		common.Coins{
			common.NewCoin(common.BTCAsset, sdk.NewUint(1)),
		},
		"SWAP:BNB.BNB",
	)

	m := NewMsgSwap(tx, common.BNBAsset, bnbAddress, sdk.NewUint(200000000), addr)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_swap")

	inputs := []struct {
		requestTxHash common.TxID
		source        common.Asset
		target        common.Asset
		amount        sdk.Uint
		requester     common.Address
		destination   common.Address
		targetPrice   sdk.Uint
		signer        sdk.AccAddress
	}{
		{
			requestTxHash: common.TxID(""),
			source:        common.RuneAsset(),
			target:        common.BNBAsset,
			amount:        sdk.NewUint(100000000),
			requester:     bnbAddress,
			destination:   bnbAddress,
			targetPrice:   sdk.NewUint(200000000),
			signer:        addr,
		},
		{
			requestTxHash: txID,
			source:        common.Asset{},
			target:        common.BNBAsset,
			amount:        sdk.NewUint(100000000),
			requester:     bnbAddress,
			destination:   bnbAddress,
			targetPrice:   sdk.NewUint(200000000),
			signer:        addr,
		},
		{
			requestTxHash: txID,
			source:        common.BNBAsset,
			target:        common.BNBAsset,
			amount:        sdk.NewUint(100000000),
			requester:     bnbAddress,
			destination:   bnbAddress,
			targetPrice:   sdk.NewUint(200000000),
			signer:        addr,
		},
		{
			requestTxHash: txID,
			source:        common.RuneAsset(),
			target:        common.Asset{},
			amount:        sdk.NewUint(100000000),
			requester:     bnbAddress,
			destination:   bnbAddress,
			targetPrice:   sdk.NewUint(200000000),
			signer:        addr,
		},
		{
			requestTxHash: txID,
			source:        common.RuneAsset(),
			target:        common.BNBAsset,
			amount:        sdk.ZeroUint(),
			requester:     bnbAddress,
			destination:   bnbAddress,
			targetPrice:   sdk.NewUint(200000000),
			signer:        addr,
		},
		{
			requestTxHash: txID,
			source:        common.RuneAsset(),
			target:        common.BNBAsset,
			amount:        sdk.NewUint(100000000),
			requester:     common.NoAddress,
			destination:   bnbAddress,
			targetPrice:   sdk.NewUint(200000000),
			signer:        addr,
		},
		{
			requestTxHash: txID,
			source:        common.RuneAsset(),
			target:        common.BNBAsset,
			amount:        sdk.NewUint(100000000),
			requester:     bnbAddress,
			destination:   common.NoAddress,
			targetPrice:   sdk.NewUint(200000000),
			signer:        addr,
		},
		{
			requestTxHash: txID,
			source:        common.RuneAsset(),
			target:        common.BNBAsset,
			amount:        sdk.NewUint(100000000),
			requester:     bnbAddress,
			destination:   bnbAddress,
			targetPrice:   sdk.NewUint(200000000),
			signer:        sdk.AccAddress{},
		},
	}
	for _, item := range inputs {
		tx := common.NewTx(
			item.requestTxHash,
			item.requester,
			item.destination,
			common.Coins{
				common.NewCoin(item.source, item.amount),
			},
			"SWAP:BNB.BNB",
		)

		m := NewMsgSwap(tx, item.target, item.destination, item.targetPrice, item.signer)
		c.Assert(m.ValidateBasic(), NotNil)
	}
}
