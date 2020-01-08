package types

type Log struct {
	MsgIndex string `json:"msg_index"`
	Success  bool   `json:"success"`
	Log      string `json:"log"`
}

type Commit struct {
	Height string `json:"height"`
	TxHash string `json:"txhash"`
	Logs   []Log  `json:"logs"`
}
