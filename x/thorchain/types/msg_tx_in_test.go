package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"
)

type MsgSetTxInSuite struct{}

var _ = Suite(&MsgSetTxInSuite{})

func (MsgSetTxInSuite) TestMsgSetTxIn(c *C) {
	txID := GetRandomTxHash()
	bnb := GetRandomBNBAddress()
	acc1 := GetRandomBech32Addr()
	observePoolAddr := GetRandomPubKey()
	txIn := NewTxIn(
		common.Coins{
			common.NewCoin(common.RuneAsset(), sdk.NewUint(1)),
		},
		"hello",
		bnb,
		GetRandomBNBAddress(),
		common.BNBGasFeeSingleton,
		sdk.NewUint(1),
		observePoolAddr,
	)
	txs := []TxInVoter{
		NewTxInVoter(txID, []TxIn{txIn}),
	}
	m := NewMsgSetTxIn(txs, acc1)
	EnsureMsgBasicCorrect(m, c)
	c.Check(m.Type(), Equals, "set_tx_hashes")

	m1 := NewMsgSetTxIn(nil, acc1)
	c.Assert(m1.ValidateBasic(), NotNil)
	m2 := NewMsgSetTxIn(txs, sdk.AccAddress{})
	c.Assert(m2.ValidateBasic(), NotNil)

	m3 := NewMsgSetTxIn([]TxInVoter{
		NewTxInVoter(common.TxID(""), []TxIn{}),
	}, acc1)
	c.Assert(m3.ValidateBasic(), NotNil)

	m4 := NewMsgSetTxIn([]TxInVoter{
		NewTxInVoter(txID, []TxIn{
			NewTxIn(nil, "hello", bnb, GetRandomBNBAddress(), common.BNBGasFeeSingleton, sdk.NewUint(1), observePoolAddr),
		}),
	}, acc1)
	c.Assert(m4.ValidateBasic(), NotNil)

}
