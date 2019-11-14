package common

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Gas Coins

// Gas Fees
var BNBGasFeeSingleton = Gas{
	{Asset: BNBAsset, Amount: sdk.NewUint(37500)},
}

var BNBGasFeeMulti = Gas{
	{Asset: BNBAsset, Amount: sdk.NewUint(30000)},
}

// Calculates the amount of gas for x number of coins in a single tx.
func GetBNBGasFeeMulti(count uint64) Gas {
	gas := BNBGasFeeMulti
	gas[0].Amount = gas[0].Amount.MulUint64(count)
	return gas
}

func (g Gas) IsValid() error {
	for _, coin := range g {
		if err := coin.IsValid(); err != nil {
			return err
		}
	}

	return nil
}
