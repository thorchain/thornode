package types

/// AccountResp the response from thorclient
type AccountResp struct {
	Height string `json:"height"`
	Result struct {
		Value struct {
			AccountNumber string `json:"account_number"`
			Sequence      string `json:"sequence"`
		} `json:"value"`
	} `json:"result"`
}
