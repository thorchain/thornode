package types

type OutTx struct {
	Height  string `json:"height"`
	Hash    string `json:"hash"`
	TxArray []OutTxItem `json:"tx_array"`
}

type OutTxItem struct {
	To    string `json:"to"`
	Coins []Coin `json:"coins"`
}
