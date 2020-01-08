package types

type BadCommit struct {
	Height string `json:"height"`
	TxHash string `json:"txhash"`
	Code   int    `json:"code"`
	Log    struct {
		CodeSpace string `json:"codespace"`
		Code      int    `json:"code"`
		Message   string `json:"message"`
	} `json:"raw_log"`
}

type log struct {
	Success bool   `json:"success"`
	Log     string `json:"log"`
}

type Commit struct {
	Height string `json:"height"`
	TxHash string `json:"txhash"`
	Logs   []log  `json:"logs"`
}
