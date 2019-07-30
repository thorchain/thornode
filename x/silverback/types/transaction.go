package types

import "time"

type Txn struct {
	Tx []struct {
		TxHash        string      `json:"txHash"`
		BlockHeight   int         `json:"blockHeight"`
		TxType        string      `json:"txType"`
		TimeStamp     time.Time   `json:"timeStamp"`
		FromAddr      string      `json:"fromAddr"`
		ToAddr        string      `json:"toAddr"`
		Value         string      `json:"value"`
		TxAsset       string      `json:"txAsset"`
		TxFee         string      `json:"txFee"`
		TxAge         int         `json:"txAge"`
		OrderID       interface{} `json:"orderId"`
		Code          int         `json:"code"`
		Data          interface{} `json:"data"`
		ConfirmBlocks int         `json:"confirmBlocks"`
		Memo          string      `json:"memo"`
	} `json:"tx"`
	Total int `json:"total"`
}
