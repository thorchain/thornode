package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type TxOutTestSuite struct{}

var _ = Suite(&TxOutTestSuite{})

func (TxOutTestSuite) TestTxOut(c *C) {
	bnbAddress, err := common.NewAddress("bnb1xlvns0n2mxh77mzaspn2hgav4rr4m8eerfju38")
	c.Assert(err, IsNil)
	txOut := NewTxOut(1)
	c.Assert(txOut, NotNil)
	c.Assert(txOut.TxArray, IsNil)
	c.Assert(txOut.IsEmpty(), Equals, true)
	c.Assert(txOut.Valid(), IsNil)
	txOutItem := &TxOutItem{
		PoolAddress: bnbAddress,
		ToAddress:   bnbAddress,
		Coins: common.Coins{
			common.NewCoin(common.BNBChain, common.BNBTicker, sdk.NewUint(100*common.One)),
		},
	}
	txOut.TxArray = append(txOut.TxArray, txOutItem)
	c.Assert(txOut.TxArray, NotNil)
	c.Check(len(txOut.TxArray), Equals, 1)
	c.Assert(txOut.IsEmpty(), Equals, false)
	c.Assert(txOut.Valid(), IsNil)
	strTxOutItem := txOutItem.String()
	c.Check(len(strTxOutItem) > 0, Equals, true)

	txOut1 := NewTxOut(2)
	txOut1.TxArray = append(txOut1.TxArray, txOutItem)
	txOut1.TxArray = append(txOut1.TxArray, &TxOutItem{
		ToAddress:   bnbAddress,
		PoolAddress: bnbAddress,
		Coins:       nil,
	})
	c.Assert(txOut1.Valid(), NotNil)

	txOut2 := NewTxOut(3)
	txOut2.TxArray = append(txOut2.TxArray, &TxOutItem{
		ToAddress:   "",
		PoolAddress: bnbAddress,
		Coins: common.Coins{
			common.NewCoin(common.BNBChain, common.BNBTicker, sdk.NewUint(100*common.One)),
		},
	})
	c.Assert(txOut2.Valid(), NotNil)
	txOut3 := NewTxOut(4)
	txOut3.TxArray = append(txOut3.TxArray, &TxOutItem{
		ToAddress:   bnbAddress,
		PoolAddress: "",
		Coins: common.Coins{
			common.NewCoin(common.BNBChain, common.BNBTicker, sdk.NewUint(100*common.One)),
		},
	})
	c.Assert(txOut3.Valid(), NotNil)
}
