package types

// Coin our custom coin definition
// This one will be replaced by sdk.Coin later on
// TODO remove this struct, and replace it with sdk.Coin
type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}
