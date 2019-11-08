package types

type Asset struct {
	Chain  string `json:"chain"`
	Symbol string `json:"symbol"`
	Ticker string `json:"string"`
}

type ThorchainPools []struct {
	BalanceRune  int64  `json:"balance_rune,string"`
	BalanceAsset int64  `json:"balance_asset,string"`
	Asset        Asset  `json:"asset"`
	PoolUnits    int64  `json:"pool_units,string"`
	PoolAddress  string `json:"pool_address"`
	Status       string `json:"status"`
}

type PoolAndUnits []PoolAndUnit
type PoolAndUnit struct {
	Asset Asset `json:"asset"`
	Units int64 `json:"units,string"`
}

type Staker struct {
	StakerID     string       `json:"staker_id"`
	PoolAndUnits PoolAndUnits `json:"pool_and_units"`
}

type ThorchainResults struct {
	Tx             int          `json:"TX"`
	ThorchainPools ThorchainPools `json:"thorchain_pools"`
}
