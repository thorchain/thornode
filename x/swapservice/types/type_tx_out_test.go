package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type TxOutTestSuite struct{}

var _ = Suite(&TxOutTestSuite{})

func (TxOutTestSuite) TestTxOut(c *C) {
	txOut := NewTxOut(1)
	c.Assert(txOut, NotNil)
	c.Assert(txOut.TxArray, IsNil)
	bnbAddress, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)
	txOutItem := &TxOutItem{
		ToAddress: bnbAddress,
		Coins: common.Coins{
			common.NewCoin(common.BNBTicker, sdk.NewUint(100*common.One)),
		},
	}
	txOut.TxArray = append(txOut.TxArray, txOutItem)
	c.Assert(txOut.TxArray, NotNil)
	c.Check(len(txOut.TxArray), Equals, 1)
	strTxOutItem := txOutItem.String()
	c.Check(len(strTxOutItem) > 0, Equals, true)

}
