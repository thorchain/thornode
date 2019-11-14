package common

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Gas Coins

var bnbSingleTxFee = sdk.NewUint(37500)
var bnbMultiTxFee = sdk.NewUint(30000)

// Gas Fees
var BNBGasFeeSingleton = Gas{
	{Asset: BNBAsset, Amount: bnbSingleTxFee},
}

var BNBGasFeeMulti = Gas{
	{Asset: BNBAsset, Amount: bnbMultiTxFee},
}

// Calculates the amount of gas for x number of coins in a single tx.
func GetBNBGasFeeMulti(count uint64) Gas {
	return Gas{
		{Asset: BNBAsset, Amount: bnbMultiTxFee.MulUint64(count)},
	}
}

func (g Gas) IsValid() error {
	for _, coin := range g {
		if err := coin.IsValid(); err != nil {
			return err
		}
	}

	return nil
}
