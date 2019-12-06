package common

import (
	"sort"

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

func GetBNBGasFee(count uint64) Gas {
	if count == 0 {
		return nil
	}
	if count == 1 {
		return BNBGasFeeSingleton
	}
	return GetBNBGasFeeMulti(count)
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

// Check if two lists of coins are equal to each other. Order does not matter
func (gas1 Gas) Equals(gas2 Gas) bool {
	if len(gas1) != len(gas2) {
		return false
	}

	// sort both lists
	sort.Slice(gas1[:], func(i, j int) bool {
		return gas1[i].Asset.String() < gas1[j].Asset.String()
	})
	sort.Slice(gas2[:], func(i, j int) bool {
		return gas2[i].Asset.String() < gas2[j].Asset.String()
	})

	for i := range gas1 {
		if !gas1[i].Equals(gas2[i]) {
			return false
		}
	}

	return true
}
