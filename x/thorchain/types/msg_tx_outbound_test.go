package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

type MsgOutboundTxSuite struct{}

var _ = Suite(&MsgOutboundTxSuite{})

func (MsgOutboundTxSuite) TestMsgOutboundTx(c *C) {
	txID := GetRandomTxHash()
	inTxID := GetRandomTxHash()
	bnb := GetRandomBNBAddress()
	acc1 := GetRandomBech32Addr()
	tx := common.NewTx(
		txID,
		bnb,
		GetRandomBNBAddress(),
		common.Coins{common.NewCoin(common.BNBAsset, sdk.OneUint())},
		"",
	)
	m := NewMsgOutboundTx(tx, 1, inTxID, acc1)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_tx_outbound")

	inputs := []struct {
		txID   common.TxID
		inTxID common.TxID
		sender common.Address
		signer sdk.AccAddress
	}{
		{
			txID:   common.TxID(""),
			inTxID: inTxID,
			sender: bnb,
			signer: acc1,
		},
		{
			txID:   txID,
			inTxID: common.TxID(""),
			sender: bnb,
			signer: acc1,
		},
		{
			txID:   txID,
			inTxID: inTxID,
			sender: common.NoAddress,
			signer: acc1,
		},
		{
			txID:   txID,
			inTxID: inTxID,
			sender: bnb,
			signer: sdk.AccAddress{},
		},
	}
	for i, item := range inputs {
		fmt.Println(i)
		tx := common.NewTx(
			item.txID,
			item.sender,
			GetRandomBNBAddress(),
			common.Coins{common.NewCoin(common.BNBAsset, sdk.OneUint())},
			"",
		)
		m := NewMsgOutboundTx(tx, 12, item.inTxID, item.signer)
		err := m.ValidateBasic()
		c.Assert(err, NotNil, Commentf("%s", err.Error()))
	}
}
