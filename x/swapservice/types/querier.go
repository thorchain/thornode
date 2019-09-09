package types

import (
	"strings"
)

// Query Result Payload for a pools query
type QueryResPools []Pool

// implement fmt.Stringer
func (n QueryResPools) String() string {
	var tickers []string
	for _, record := range n {
		tickers = append(tickers, record.Ticker.String())
	}
	return strings.Join(tickers, "\n")
}
