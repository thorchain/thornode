package types

type StatechainPools []struct {
	BalanceRune  int64 `json:"balance_rune,string"`
	BalanceAsset int64 `json:"balance_asset,string"`
	Asset        struct {
		Chain  string `json:"chain"`
		Symbol string `json:"symbol"`
		Ticker string `json:"string"`
	} `json:"asset"`
	PoolUnits   int64  `json:"pool_units,string"`
	PoolAddress string `json:"pool_address"`
	Status      string `json:"status"`
}

type Staker struct {
	StakerID     string `json:"staker_id"`
	PoolAndUnits []struct {
		Asset struct {
			Chain  string `json:"chain"`
			Symbol string `json:"symbol"`
			Ticker string `json:"string"`
		} `json:"asset"`
		Units int64 `json:"units,string"`
	} `json:"pool_and_units"`
}
