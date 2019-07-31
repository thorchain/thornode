package types

type Tokens []struct {
	Symbol string `json:"symbol"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
	Frozen string `json:"frozen"`
}
