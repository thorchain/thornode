package types

type QueryResult struct {
	Result struct {
		Response struct {
			Value string `json:"value"`
		} `json:"response"`
	} `json:"result"`
}

type itemData struct {
	Txs []string `json:"txs"`
}

type itemHeader struct {
	Height string `json:"height"`
}

type itemBlock struct {
	Header itemHeader `json:"header"`
	Data   itemData   `json:"data"`
}

type itemResult struct {
	Block itemBlock `json:"block"`
}

type BlockResult struct {
	Jsonrpc string     `json:"jsonrpc"`
	ID      string     `json:"id"`
	Result  itemResult `json:"result"`
}
