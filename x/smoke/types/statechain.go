package types

import "time"

const (
	StatechainURL = "https://testnet-chain.bepswap.net/swapservice/pools"
	PoolSymbol    = "BNB"
)

type StatechainPools []struct {
	BalanceRune  string    `json:"balance_rune"`
	BalanceToken string    `json:"balance_token"`
	Symbol       string    `json:"symbol"`
	PoolUnits    string    `json:"pool_units"`
	PoolAddress  string    `json:"pool_address"`
	Status       string    `json:"status"`
	ExpiryUtc    time.Time `json:"expiry_utc"`
}
