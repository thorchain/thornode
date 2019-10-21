package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type MsgSetTxInSuite struct{}

var _ = Suite(&MsgSetTxInSuite{})

func (MsgSetTxInSuite) TestMsgSetTxIn(c *C) {
	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	bnb, err := common.NewAddress("bnb1xlvns0n2mxh77mzaspn2hgav4rr4m8eerfju38")
	c.Assert(err, IsNil)
	acc1, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	observePoolAddr, err := common.NewAddress("bnb1dvpx8jru4mcyknavp2dtvw8uxg007pv8de35f9")
	c.Assert(err, IsNil)
	txIn := NewTxIn(common.Coins{
		common.NewCoin(common.BNBChain, common.RuneA1FTicker, sdk.NewUint(1)),
	}, "hello", bnb, sdk.NewUint(1), observePoolAddr)
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
			NewTxIn(nil, "hello", bnb, sdk.NewUint(1), observePoolAddr),
		}),
	}, acc1)
	c.Assert(m4.ValidateBasic(), NotNil)

}
