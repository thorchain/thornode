package common

import (
	"fmt"
	"strings"

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

func (c Coin) Equals(cc Coin) bool {
	if !c.Asset.Equals(cc.Asset) {
		return false
	}
	if !c.Amount.Equal(cc.Amount) {
		return false
	}
	return true
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

func (c Coin) String() string {
	return fmt.Sprintf("%s%s", c.Asset.String(), c.Amount.String())
}

func (cs Coins) IsValid() error {
	for _, coin := range cs {
		if err := coin.IsValid(); err != nil {
			return err
		}
	}

	return nil
}

func (cs Coins) String() string {
	coins := make([]string, len(cs))
	for i, c := range cs {
		coins[i] = c.String()
	}
	return strings.Join(coins, ", ")
}
