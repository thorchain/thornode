package types

// type TxArr struct {
// 	RequestTxHash string
// 	From          string
// 	To            string
// 	Token         string
// 	Amount        string
// }

// type OutTx struct {
// 	TxOutID string
// 	Hash    string
// 	TxArray []TxArr
// }

type OutTx struct {
	Height  string `json:"height"`
	Hash    string `json:"hash"`
	TxArray []TxItem `json:"tx_array"`
}

type TxItem struct {
	To    string `json:"to"`
	Coins []Coin `json:"coins"`
}

type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}
