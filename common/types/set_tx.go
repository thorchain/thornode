package types

type SetTx struct {
	Mode string `json:"mode"`
	Tx struct {
		Msg []Msg `json:"msg"`
		Fee struct {
			Amount []struct{} `json:"amount"`
			Gas string `json:"gas"`
		} `json:"fee"`
		Signatures []Signature `json:"signatures"`
		Memo string `json:"memo"`
	} `json:"tx"`
}
