package types

type InTx struct {	
	BlockHeight int `json:"blockHeight"`
	Count				int `json:"count"`
	TxArray     []InTxItem `json:"txArray"`
}

type InTxItem struct {
	Tx     string `json:"tx"`
	Memo   string `json:"MEMO"`
	Sender string `json:"sender"`
	Coins  []Coins	`json:"coins"`
}
