package common

import (
	"fmt"

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

func (c Coin) Valid() error {
	if c.Chain.IsEmpty() {
		return fmt.Errorf("Chain cannot be empty")
	}
	if c.Denom.IsEmpty() {
		return fmt.Errorf("Denom cannot be empty")
	}

	return nil
}
