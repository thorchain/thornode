package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

type TxOutTestSuite struct{}

var _ = Suite(&TxOutTestSuite{})

func (TxOutTestSuite) TestTxOut(c *C) {
	bnbAddress := GetRandomPubKey()
	toAddr := GetRandomBNBAddress()
	txOut := NewTxOut(1)
	c.Assert(txOut, NotNil)
	c.Assert(txOut.TxArray, IsNil)
	c.Assert(txOut.IsEmpty(), Equals, true)
	c.Assert(txOut.Valid(), IsNil)
	txOutItem := &TxOutItem{
		Chain:       common.BNBChain,
		PoolAddress: bnbAddress,
		ToAddress:   toAddr,
		Coins: common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One)),
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
		Chain:       common.BNBChain,
		ToAddress:   toAddr,
		PoolAddress: bnbAddress,
		Coins:       nil,
	})
	c.Assert(txOut1.Valid(), NotNil)

	txOut2 := NewTxOut(3)
	txOut2.TxArray = append(txOut2.TxArray, &TxOutItem{
		Chain:       common.BNBChain,
		ToAddress:   "",
		PoolAddress: bnbAddress,
		Coins: common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One)),
		},
	})
	c.Assert(txOut2.Valid(), NotNil)
	txOut3 := NewTxOut(4)
	txOut3.TxArray = append(txOut3.TxArray, &TxOutItem{
		Chain:       common.BNBChain,
		ToAddress:   toAddr,
		PoolAddress: common.EmptyPubKey,
		Coins: common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One)),
		},
	})
	c.Assert(txOut3.Valid(), NotNil)
}
