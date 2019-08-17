package types

type ApiTx struct {
	Hash   string `json:"hash"`
	Height string `json:"height"`
	Log    string `json:"log"`
	Ok     bool   `json:"ok"`
	Tx     struct {
		Type  string `json:"type"`
		Value struct {
			Memo string `json:"memo"`
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
		} `json:"value"`
	} `json:"tx"`
}
