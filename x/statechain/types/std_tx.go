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
