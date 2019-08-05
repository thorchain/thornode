package types

import "strings"

// Query Result Payload for a pools query
type QueryResPoolStructs []PoolStruct

// implement fmt.Stringer
func (n QueryResPoolStructs) String() string {
	var tickers []string
	for _, record := range n {
		tickers = append(tickers, record.Ticker)
	}
	return strings.Join(tickers[:], "\n")
}
