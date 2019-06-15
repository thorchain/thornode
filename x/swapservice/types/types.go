package types

import (
	"fmt"
	"strings"
)

// PoolStruct is a struct that contains all the metadata of a pooldata
type PoolStruct struct {
	BalanceAtom  string `json:"balance_atom"`
	BalanceToken string `json:"balance_token"`
	Ticker       string `json:"ticker"`
	TokenName    string `json:"token_name"`
}

// Returns a new PoolStruct
func NewPoolStruct() PoolStruct {
	return PoolStruct{}
}

// implement fmt.Stringer
func (w PoolStruct) String() string {
	return strings.TrimSpace(fmt.Sprintf(`Token: %s (%s) 
Atom Balance: %s
Token Balance: %s`, w.TokenName, w.Ticker, w.BalanceAtom, w.BalanceToken))
}
