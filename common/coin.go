package common

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Coin struct {
	Asset  Asset    `json:"asset"`
	Amount sdk.Uint `json:"amount"`
}

var NoCoin = Coin{}

type Coins []Coin

// NewCoin return a new instance of Coin
func NewCoin(asset Asset, amount sdk.Uint) Coin {
	return Coin{
		Asset:  asset,
		Amount: amount,
	}
}

func (c Coin) IsEmpty() bool {
	if c.Asset.IsEmpty() {
		return true
	}
	if c.Amount.IsZero() {
		return true
	}
	return false
}

func (c Coin) IsValid() error {
	if c.Asset.IsEmpty() {
		return fmt.Errorf("Denom cannot be empty")
	}
	if c.Amount.IsZero() {
		return fmt.Errorf("Amount cannot be zero")
	}

	return nil
}
