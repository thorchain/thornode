package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type MsgRefundTxSuite struct{}

var _ = Suite(&MsgRefundTxSuite{})

func (MsgRefundTxSuite) TestMsgRefundTx(c *C) {
	txID := GetRandomTxHash()
	inTxID := GetRandomTxHash()
	bnb := GetRandomBNBAddress()
	acc1 := GetRandomBech32Addr()
	tx := NewObservedTx(common.NewTx(
		txID,
		bnb,
		GetRandomBNBAddress(),
		common.Coins{common.NewCoin(common.BNBAsset, sdk.OneUint())},
		common.BNBGasFeeSingleton,
		"",
	), sdk.NewUint(12), GetRandomPubKey())
	m := NewMsgRefundTx(tx, inTxID, acc1)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_tx_refund")

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

	for _, item := range inputs {
		tx := NewObservedTx(common.NewTx(
			item.txID,
			item.sender,
			GetRandomBNBAddress(),
			common.Coins{common.NewCoin(common.BNBAsset, sdk.OneUint())},
			common.BNBGasFeeSingleton,
			"",
		), sdk.NewUint(12), GetRandomPubKey())
		m := NewMsgRefundTx(tx, item.inTxID, item.signer)
		err := m.ValidateBasic()
		c.Assert(err, NotNil, Commentf("%s", err.Error()))
	}
}
