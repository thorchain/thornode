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
	m := NewMsgOutboundTx(txID, 1, bnb, common.BNBChain, acc1)
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
		m := NewMsgOutboundTx(item.txID, item.height, item.sender, common.BNBChain, item.signer)
		c.Assert(m.ValidateBasic(), NotNil)
	}
}
