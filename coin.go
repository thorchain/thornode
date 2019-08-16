package common

type Coin struct {
	Denom  Ticker `json:"denom"`
	Amount Amount `json:"amount"`
}

type Coins []Coin

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
