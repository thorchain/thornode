package types

import "time"

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
			Msg []Msg `json:"msg"`
			Fee struct {
				Amount []interface{} `json:"amount"`
				Gas    string        `json:"gas"`
			} `json:"fee"`
			Signatures []Signature `json:"signatures"`
			Memo       string      `json:"memo"`
		} `json:"value"`
	} `json:"tx"`
	Timestamp time.Time `json:"timestamp"`
}
