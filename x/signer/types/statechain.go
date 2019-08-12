package types

import "time"

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

type Txs struct {
	Height string `json:"height"`
	Txhash string `json:"txhash"`
	RawLog string `json:"raw_log"`
	Logs   []struct {
		MsgIndex int    `json:"msg_index"`
		Success  bool   `json:"success"`
		Log      string `json:"log"`
	} `json:"logs"`
	GasWanted string `json:"gas_wanted"`
	GasUsed   string `json:"gas_used"`
	Events    []struct {
		Type       string `json:"type"`
		Attributes []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"attributes"`
	} `json:"events"`
	Tx struct {
		Type  string `json:"type"`
		Value struct {
			Msg []struct {
				Type  string `json:"type"`
				Value struct {
					TxHashes []struct {
						Request string `json:"request"`
						Status  string `json:"status"`
						Txhash  string `json:"txhash"`
						Memo    string `json:"memo"`
						Coins   []struct {
							Denom  string `json:"denom"`
							Amount string `json:"amount"`
						} `json:"coins"`
						Sender string `json:"sender"`
					} `json:"tx_hashes"`
					Signer string `json:"signer"`
				} `json:"value"`
			} `json:"msg"`
			Fee struct {
				Amount []interface{} `json:"amount"`
				Gas    string        `json:"gas"`
			} `json:"fee"`
			Signatures []struct {
				PubKey struct {
					Type  string `json:"type"`
					Value string `json:"value"`
				} `json:"pub_key"`
				Signature string `json:"signature"`
			} `json:"signatures"`
			Memo string `json:"memo"`
		} `json:"value"`
	} `json:"tx"`
	Timestamp time.Time `json:"timestamp"`
}
