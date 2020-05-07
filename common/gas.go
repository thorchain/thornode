package common

import (
	"math/big"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Gas Coins

var (
	bnbSingleTxFee = sdk.NewUint(37500)
	bnbMultiTxFee  = sdk.NewUint(30000)
	ethTransferFee = sdk.NewUint(21000)
	ethGasPerByte  = sdk.NewUint(68)
)

// Gas Fees
var BNBGasFeeSingleton = Gas{
	{Asset: BNBAsset, Amount: bnbSingleTxFee},
}

var BNBGasFeeMulti = Gas{
	{Asset: BNBAsset, Amount: bnbMultiTxFee},
}

var ETHGasFeeTransfer = Gas{
	{Asset: ETHAsset, Amount: ethTransferFee},
}

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
			// defaults
			units = []sdk.Uint{sdk.NewUint(37500), sdk.NewUint(30000)}
		}
		gasCoin := tx.Gas.ToCoins().GetCoin(BNBAsset)
		lenCoins := uint64(len(tx.Coins))
		if lenCoins == 1 {
			units[0] = gasCoin.Amount
		} else if lenCoins > 1 {
			units[1] = gasCoin.Amount.QuoUint64(lenCoins)
		}
	case BTCAsset, ETHAsset:
		// BTC chain there is only one coin, which is bitcoin, gas is paid in bitcoin as well
		gasCoin := tx.Gas.ToCoins().GetCoin(asset)
		if nil == units {
			return []sdk.Uint{gasCoin.Amount}
		}
		units[0] = gasCoin.Amount

	}
	return units
}

// UpdateBNBGasFee
func UpdateBNBGasFee(gas Gas, numberCoins int) {
	if gas.IsEmpty() {
		return
	}
	if err := gas.IsValid(); err != nil {
		return
	}
	gasCoin := gas.ToCoins().GetCoin(BNBAsset)
	if gasCoin.Equals(NoCoin) {
		return
	}

	if numberCoins == 1 {
		if gasCoin.Amount.Equal(bnbSingleTxFee) {
			return
		}
		bnbSingleTxFee = gasCoin.Amount
		BNBGasFeeSingleton = Gas{
			{Asset: BNBAsset, Amount: bnbSingleTxFee},
		}
		return
	}
	multiGas := gasCoin.Amount.QuoUint64(uint64(numberCoins))
	if multiGas.Equal(bnbMultiTxFee) {
		return
	}
	bnbMultiTxFee = multiGas
	BNBGasFeeMulti = Gas{
		{Asset: BNBAsset, Amount: multiGas},
	}
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

func GetETHGasFee(gasPrice *big.Int, msgLen uint64) Gas {
	gasBytes := ethGasPerByte.MulUint64(msgLen)
	return Gas{
		{Asset: ETHAsset, Amount: ethTransferFee.Add(gasBytes).Mul(sdk.NewUintFromBigInt(gasPrice))},
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
