package types

type SocketTxfr struct {
	Stream string `json:"stream"`
	Data   struct {
		EventType   string `json:"e"`
		EventHeight int    `json:"E"`
		Hash        string `json:"H"`
		Memo        string `json:"M"`
		FromAddr    string `json:"f"`
		T           []struct {
			ToAddr string `json:"o"`
			Coins  []struct {
				Asset  string `json:"a"`
				Amount string `json:"A"`
			} `json:"c"`
		} `json:"t"`
	} `json:"data"`
}
