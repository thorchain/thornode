package types

import (
	"fmt"
	"strings"
)

// Shows if a tx hash has been processed
type QueryTxHash struct {
	Done     string `json:"done"`
	Refunded string `json:"refunded"`
}

func (n QueryTxHash) String() string {
	return fmt.Sprintf("TxDone: %s | TxRefunded: %s", n.Done, n.Refunded)
}

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
