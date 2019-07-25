package types

import (
	"fmt"
	"strings"
)

// Staker means staker obviosuly
type Staker struct {
	PoolID        string `json:"pool_id"`        // represent the pool that they stake in
	Ticker        string `json:"ticker"`         // the token symbol
	TokenAmount   string `json:"token_amount"`   // Total amount token they put in
	RuneAmount    string `json:"rune_amount"`    // Total amount of Rune they put in
	PoolUnits     string `json:"pool_units"`     // PoolUnits they own
	PublicAddress string `json:"public_address"` // their address on binance chain
}

func getPoolIDFromTicker(ticker string) string {
	return "pool-" + ticker
}

// NewStaker create a new instance of stake
func NewStaker(ticker, tokenAmount, runeAmount, poolUnits, publicAddress string) Staker {
	return Staker{
		PoolID:        getPoolIDFromTicker(ticker),
		Ticker:        ticker,
		TokenAmount:   tokenAmount,
		RuneAmount:    runeAmount,
		PoolUnits:     poolUnits,
		PublicAddress: publicAddress,
	}
}

// String implement fmt.Stringer , return a friend string
func (s Staker) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintln("pool-id: %s" + s.PoolID))
	sb.WriteString(fmt.Sprintln("ticker: " + s.Ticker))
	sb.WriteString(fmt.Sprintln("token-amount: " + s.TokenAmount))
	sb.WriteString(fmt.Sprintln("rune-amount: " + s.RuneAmount))
	sb.WriteString(fmt.Sprintln("pool-units: " + s.PoolUnits))
	sb.WriteString(fmt.Sprintln("public-address: " + s.PublicAddress))

	return sb.String()
}
