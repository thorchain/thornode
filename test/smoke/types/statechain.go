package types

type Pools []struct {
	BalanceRune  float64 `json:"balance_rune,string"`
	BalanceAsset float64 `json:"balance_asset,string"`
	Asset        struct {
		Chain  string `json:"chain"`
		Symbol string `json:"symbol"`
		Ticker string `json:"string"`
	} `json:"asset"`
	PoolUnits   float64 `json:"pool_units,string"`
	PoolAddress string  `json:"pool_address"`
	Status      string  `json:"status"`
}

type Staker struct {
	StakerID     string `json:"staker_id"`
	PoolAndUnits []struct {
		Asset struct {
			Chain  string `json:"chain"`
			Symbol string `json:"symbol"`
			Ticker string `json:"string"`
		} `json:"asset"`
		Units float64 `json:"units,string"`
	} `json:"pool_and_units"`
}
