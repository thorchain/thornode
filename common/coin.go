package common

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Coin struct {
	Asset  Asset    `json:"asset"`
	Amount sdk.Uint `json:"amount"`
}

type Coins []Coin

// NewCoin return a new instance of Coin
func NewCoin(asset Asset, amount sdk.Uint) Coin {
	return Coin{
		Asset:  asset,
		Amount: amount,
	}
}

func (c Coin) Valid() error {
	if c.Asset.IsEmpty() {
		return fmt.Errorf("Denom cannot be empty")
	}

	return nil
}
