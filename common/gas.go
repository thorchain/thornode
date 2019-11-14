package common

import sdk "github.com/cosmos/cosmos-sdk/types"

type Gas Coins

// Gas Fees
var BNBGasFeeSingleton = Gas{
	{Asset: BNBAsset, Amount: sdk.NewUint(37500)},
}

func (g Gas) IsValid() error {
	for _, coin := range g {
		if err := coin.IsValid(); err != nil {
			return err
		}
	}

	return nil
}
