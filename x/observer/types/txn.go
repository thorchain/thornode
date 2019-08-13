package types

import "time"

type InTx struct {	
	BlockHeight int `json:"blockHeight"`
	Count				int `json:"count"`
	TxArray     []TxItem `json:"txArray"`
}

type TxItem struct {
	Tx     string `json:"tx"`
	Memo   string `json:"MEMO"`
	Sender string `json:"sender"`
	Coins  struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	} `json:"coins"`
}

type Coins struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type Txfr struct {
	Stream string `json:"stream"`
	Data   struct {
		EventType 	string 	`json:"e"`
		EventHeight int   	`json:"E"`
		Hash 				string 	`json:"H"`
		Memo 				string	`json:"M"`
		FromAddr 		string 	`json:"f"`
		T []struct {
			ToAddr string `json:"o"`
			Coins []struct {
				Asset string `json:"a"`
				Amount string `json:"A"`
			} `json:"c"`
		} `json:"t"`
	} `json:"data"`
}

type Txns struct {
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
	}
	Total int `json:"total"`
}

type Tx struct {
	Code   int    `json:"code"`
	Hash   string `json:"hash"`
	Height string `json:"height"`
	Log    string `json:"log"`
	Ok     bool   `json:"ok"`
	Tx     struct {
		Type  string `json:"type"`
		Value struct {
			Data interface{} `json:"data"`
			Memo string      `json:"memo"`
			Msg  []struct {
				Type  string `json:"type"`
				Value struct {
					Inputs []struct {
						Address string `json:"address"`
						Coins   []struct {
							Amount string `json:"amount"`
							Denom  string `json:"denom"`
						} `json:"coins"`
					} `json:"inputs"`
					Outputs []struct {
						Address string `json:"address"`
						Coins   []struct {
							Amount string `json:"amount"`
							Denom  string `json:"denom"`
						} `json:"coins"`
					} `json:"outputs"`
				} `json:"value"`
			} `json:"msg"`
			Signatures []struct {
				AccountNumber string `json:"account_number"`
				PubKey        struct {
					Type  string `json:"type"`
					Value string `json:"value"`
				} `json:"pub_key"`
				Sequence  string `json:"sequence"`
				Signature string `json:"signature"`
			} `json:"signatures"`
			Source string `json:"source"`
		} `json:"value"`
	} `json:"tx"`
}
