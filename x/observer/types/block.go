package types

type Block struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Result  struct {
		Txs []TxsItem `json:"txs"`
		TotalCount string `json:"total_count"`
	} `json:"result"`
}

type TxsItem struct {
	Hash     string `json:"hash"`
	Height   string `json:"height"`
	Index    int    `json:"index"`
	TxResult struct {
		Data string `json:"data"`
		Log  string `json:"log"`
		Tags []TxsItemTag `json:"tags"`
	} `json:"tx_result,omitempty"`
	Tx    string `json:"tx"`
	Proof struct {
		RootHash string `json:"RootHash"`
		Data     string `json:"Data"`
		Proof    struct {
			Total    string   `json:"total"`
			Index    string   `json:"index"`
			LeafHash string   `json:"leaf_hash"`
			Aunts    []string `json:"aunts"`
		} `json:"Proof"`
	} `json:"proof"`
}

type TxsItemTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
