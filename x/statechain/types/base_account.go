package types

type BaseAccount struct {
	Value struct {
		AccountNumber string `json:"account_number" yaml:"account_number"`
		Sequence      string `json:"sequence" yaml:"sequence"`
	} `json:"value"`
}
