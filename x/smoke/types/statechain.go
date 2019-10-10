package types

import "time"

type Pools []struct {
	BalanceRune  string    `json:"balance_rune"`
	BalanceToken string    `json:"balance_token"`
	Symbol       string    `json:"symbol"`
	PoolUnits    string    `json:"pool_units"`
	PoolAddress  string    `json:"pool_address"`
	Status       string    `json:"status"`
	ExpiryUtc    time.Time `json:"expiry_utc"`
}
