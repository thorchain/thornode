package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type MsgOutboundTxSuite struct{}

var _ = Suite(&MsgOutboundTxSuite{})

func (MsgOutboundTxSuite) TestMsgOutboundTx(c *C) {
	txID := GetRandomTxHash()
	bnb := GetRandomBNBAddress()
	acc1 := GetRandomBech32Addr()
	m := NewMsgOutboundTx(txID, 1, bnb, acc1)
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
		m := NewMsgOutboundTx(item.txID, item.height, item.sender, item.signer)
		c.Assert(m.ValidateBasic(), NotNil)
	}
}
