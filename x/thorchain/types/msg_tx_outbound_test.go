package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

type MsgOutboundTxSuite struct{}

var _ = Suite(&MsgOutboundTxSuite{})

func (MsgOutboundTxSuite) TestMsgOutboundTx(c *C) {
	txID := GetRandomTxHash()
	bnb := GetRandomBNBAddress()
	acc1 := GetRandomBech32Addr()
	tx := common.NewTx(
		txID,
		bnb,
		GetRandomBNBAddress(),
		common.Coins{common.NewCoin(common.BNBAsset, sdk.OneUint())},
		"",
	)
	m := NewMsgOutboundTx(tx, 1, acc1)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_tx_outbound")

	inputs := []struct {
		txID   common.TxID
		height uint64
		sender common.Address
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
			sender: common.NoAddress,
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
		tx := common.NewTx(
			item.txID,
			item.sender,
			GetRandomBNBAddress(),
			common.Coins{common.NewCoin(common.BNBAsset, sdk.OneUint())},
			"",
		)
		m := NewMsgOutboundTx(tx, item.height, item.signer)
		c.Assert(m.ValidateBasic(), NotNil)
	}
}
