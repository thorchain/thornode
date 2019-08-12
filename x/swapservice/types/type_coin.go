package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

// Coin our custom coin definition
// This one will be replaced by sdk.Coin later on
type Coin struct {
	Denom  Ticker `json:"denom"`
	Amount Amount `json:"amount"`
}

// NewCoin return a new instance of Coin
func NewCoin(denom Ticker, amount Amount) Coin {
	if amount.IsNegative() {
		amount = ZeroAmount
	}
	return Coin{
		Denom:  denom,
		Amount: amount,
	}
}

// Coins
type Coins []Coin

// FromSdkCoins convert the sdk.Coins type to our own coin type
func FromSdkCoins(c sdk.Coins) (Coins, error) {
	var cs Coins
	for _, item := range c {
		t, err := NewTicker(item.Denom)
		if nil != err {
			return nil, errors.Wrapf(err, "fail to convert sdk.Coin to statechain Coin type,ticker:%s invalid", item.Denom)
		}
		a, err := NewAmount(item.Amount.String())
		if nil != err {
			return nil, errors.Wrapf(err, "fail to convert amount %s ", item.Amount)
		}
		cs = append(cs, NewCoin(t, a))
	}
	return cs, nil
}
