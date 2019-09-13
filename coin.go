package common

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Coin struct {
	Denom  Ticker   `json:"denom"`
	Amount sdk.Uint `json:"amount"`
}

type Coins []Coin

// NewCoin return a new instance of Coin
func NewCoin(denom Ticker, amount sdk.Uint) Coin {
	return Coin{
		Denom:  denom,
		Amount: amount,
	}
}
