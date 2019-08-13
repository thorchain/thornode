package types

type Msg struct {
	Type string `json:"type"`
	Value struct {
		TxHashes []TxHash `json:"tx_hashes"`
		Signer string `json:"signer"`
	} `json:"value"`
}

type TxHash struct {
	Request string `json:"request"`
	Status 	string `json:"status"`
	Txhash  string `json:"txhash"`
	Memo    string `json:"memo"`
	Coins   []Coins `json:"coins"`
	Sender string `json:"sender"`
}
