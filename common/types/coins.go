package types

type Coins struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type Coin Coins
