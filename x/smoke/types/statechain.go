package types

type Pools []struct {
	BalanceRune  float64 `json:"balance_rune,string"`
	BalanceToken float64 `json:"balance_token,string"`
	Symbol       string  `json:"symbol"`
	PoolUnits    float64 `json:"pool_units,string"`
	PoolAddress  string  `json:"pool_address"`
	Status       string  `json:"status"`
}

type Staker struct {
	StakerID     string `json:"staker_id"`
	PoolAndUnits []struct {
		Symbol string  `json:"symbol"`
		Units  float64 `json:"units,string"`
	} `json:"pool_and_units"`
}
