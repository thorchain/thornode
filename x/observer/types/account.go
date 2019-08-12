package types

type Account struct {
	AccountNumber int    `json:"account_number"`
	Address       string `json:"address"`
	Balances      []struct {
		Free   string `json:"free"`
		Frozen string `json:"frozen"`
		Locked string `json:"locked"`
		Symbol string `json:"symbol"`
	}
	Flags     int   `json:"flags"`
	PublicKey []int `json:"public_key"`
	Sequence  int   `json:"sequence"`
}
