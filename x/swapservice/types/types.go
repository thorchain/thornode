package types

import (
	"fmt"
	"strings"
)

// PoolStruct is a struct that contains all the metadata of a pooldata
// This is the structure we will saved to the key value store
type PoolStruct struct {
	PoolID       string `json:"p_id"` // pool id
	BalanceRune  string `json:"r"`    // how many RUNE in the pool
	BalanceToken string `json:"t"`    // how many token in the pool
	Ticker       string `json:"ti"`   // what's the token's ticker
	TokenName    string `json:"tn"`   // what's the token's name
	PoolUnits    string `json:"pu"`   // total units of the pool
	PoolAddress  string `json:"addr"` // pool address on binance chain
	Status       string `json:"s"`    // status
}

// Returns a new PoolStruct
func NewPoolStruct() PoolStruct {
	return PoolStruct{
		BalanceRune:  "0",
		BalanceToken: "0",
		PoolUnits:    "0",
	}
}

// String implement fmt.Stringer
func (w PoolStruct) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintln("pool-id: " + w.PoolID))
	sb.WriteString(fmt.Sprintln("rune-balance: " + w.BalanceRune))
	sb.WriteString(fmt.Sprintln("token-balance: " + w.BalanceToken))
	sb.WriteString(fmt.Sprintln("ticker: " + w.Ticker))
	sb.WriteString(fmt.Sprintln("token-name: " + w.TokenName))
	sb.WriteString(fmt.Sprintln("pool-units: " + w.PoolUnits))
	sb.WriteString(fmt.Sprintln("pool-address" + w.PoolAddress))
	return sb.String()
}
