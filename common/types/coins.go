package types

type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type Coins []Coin
