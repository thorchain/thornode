package common

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	bnbSingleTxFee = sdk.NewUint(37500)
	bnbMultiTxFee  = sdk.NewUint(30000)
)

// Gas Fees
var BNBGasFeeSingleton = Gas{
	{Asset: BNBAsset, Amount: bnbSingleTxFee},
}

var BNBGasFeeMulti = Gas{
	{Asset: BNBAsset, Amount: bnbMultiTxFee},
}

type Gas Coins

func CalcGasPrice(tx Tx, asset Asset, units []sdk.Uint) Gas {
	lenCoins := uint64(len(tx.Coins))

	switch asset {
	case BNBAsset:
		if lenCoins == 0 {
			return nil
		} else if lenCoins == 1 {
			return Gas{NewCoin(BNBAsset, units[0])}
		} else if lenCoins > 1 {
			return Gas{NewCoin(BNBAsset, units[1].MulUint64(lenCoins))}
		}
	}
	return nil
}

func UpdateGasPrice(tx Tx, asset Asset, units []sdk.Uint) []sdk.Uint {
	if tx.Gas.IsEmpty() {
		// no change
		return units
	}

	switch asset {
	case BNBAsset:
		// first unit is single txn, second unit is multiple transactions
		if len(units) != 2 {
			units = make([]sdk.Uint, 2)
		}
		gasCoin := tx.Gas.ToCoins().GetCoin(BNBAsset)
		lenCoins := uint64(len(tx.Coins))
		if lenCoins == 1 {
			units[0] = gasCoin.Amount
		} else if lenCoins > 1 {
			units[1] = gasCoin.Amount.QuoUint64(lenCoins)
		}
	}
	return units
}

func (g Gas) IsValid() error {
	for _, coin := range g {
		if err := coin.IsValid(); err != nil {
			return err
		}
	}

	return nil
}

func (g Gas) IsEmpty() bool {
	for _, coin := range g {
		if !coin.IsEmpty() {
			return false
		}
	}
	return true
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

func (gas Gas) ToCoins() Coins {
	coins := make(Coins, len(gas))
	for i := range gas {
		coins[i] = NewCoin(gas[i].Asset, gas[i].Amount)
	}
	return coins
}
