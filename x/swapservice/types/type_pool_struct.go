package types

import (
	"fmt"
	"strings"
)

// PoolDataKeyPrefix used as the key's prefix
const PoolDataKeyPrefix = "pool-"

// PoolStatus is an indication of what the pool state is
type PoolStatus int

const (
	// Active means the pool is active
	Active PoolStatus = iota
	// Suspended means we take down the pool temporary
	Suspended
)

var poolStatusStr = map[string]PoolStatus{
	"Active":    Active,
	"Suspended": Suspended,
}

// String implement stringer
func (ps PoolStatus) String() string {
	for key, item := range poolStatusStr {
		if item == ps {
			return key
		}
	}
	return ""
}

// GetPoolStatus from string
func GetPoolStatus(ps string) PoolStatus {
	for key, item := range poolStatusStr {
		if strings.EqualFold(key, ps) {
			return item
		}
	}
	return Active
}

// PoolStruct is a struct that contains all the metadata of a pooldata
// This is the structure we will saved to the key value store
type PoolStruct struct {
	BalanceRune  string `json:"balance_rune"`  // how many RUNE in the pool
	BalanceToken string `json:"balance_token"` // how many token in the pool
	Ticker       string `json:"ticker"`        // what's the token's ticker
	PoolUnits    string `json:"pool_units"`    // total units of the pool
	PoolAddress  string `json:"pool_address"`  // pool address on binance chain
	Status       string `json:"status"`        // status
}

// NewPoolStruct Returns a new PoolStruct
func NewPoolStruct() PoolStruct {
	return PoolStruct{
		BalanceRune:  "0",
		BalanceToken: "0",
		PoolUnits:    "0",
	}
}

func (w PoolStruct) Empty() bool {
	return w.Ticker == ""
}

// String implement fmt.Stringer
func (w PoolStruct) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintln("rune-balance: " + w.BalanceRune))
	sb.WriteString(fmt.Sprintln("token-balance: " + w.BalanceToken))
	sb.WriteString(fmt.Sprintln("ticker: " + w.Ticker))
	sb.WriteString(fmt.Sprintln("pool-units: " + w.PoolUnits))
	sb.WriteString(fmt.Sprintln("pool-address: " + w.PoolAddress))
	sb.WriteString(fmt.Sprintln("status: " + w.Status))
	return sb.String()
}
