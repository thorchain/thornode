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

// This function combines two gas objects into one, adding amounts where needed
// or appending new coins.
func (g Gas) Add(g2 Gas) Gas {
	var newGasCoins Gas
	for _, gc2 := range g2 {
		matched := false
		for i, gc1 := range g {
			if gc1.Asset.Equals(gc2.Asset) {
				g[i].Amount = g[i].Amount.Add(gc2.Amount)
				matched = true
			}
		}
		if !matched {
			newGasCoins = append(newGasCoins, gc2)
		}
	}

	return append(g, newGasCoins...)
}
