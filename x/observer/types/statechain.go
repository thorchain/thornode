package types

type StdTx struct {
	Type string `json:"type"`
	Value struct {
		Msg []Msg `json:"msg"`
		Fee struct {
			Amount []struct{} `json:"amount"`
			Gas string `json:"gas"`
		} `json:"fee"`
		Signatures []Signature `json:"signatures"`
		Memo string `json:"memo"`
	} `json:"value"`
}

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

type Signature struct {
	PubKey struct {
		Type string `json:"type"`
		Value string `json:"value"`
	} `json:"pub_key"`
	Signature string `json:"signature"`
}

type SetTx struct {
	Mode string `json:"mode"`
	Tx struct {
		Msg []Msg `json:"msg"`
		Fee struct {
			Amount []struct{} `json:"amount"`
			Gas string `json:"gas"`
		} `json:"fee"`
		Signatures []Signature `json:"signatures"`
		Memo string `json:"memo"`
	} `json:"tx"`
}

type Commit struct {
	Height string `json:"height"`
	TxHash string `json:"txhash"`
}
