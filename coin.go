package common

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Coin struct {
	Chain  Chain    `json:"chain"`
	Denom  Ticker   `json:"denom"`
	Amount sdk.Uint `json:"amount"`
}

type Coins []Coin

// NewCoin return a new instance of Coin
func NewCoin(chain Chain, denom Ticker, amount sdk.Uint) Coin {
	return Coin{
		Chain:  chain,
		Denom:  denom,
		Amount: amount,
	}
}
