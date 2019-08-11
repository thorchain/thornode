package types

import "strings"

// UnstakeRecord is a record about unstake activity
type UnstakeRecord struct {
	RequestTxHash  TxID   `json:"request_tx_hash"`
	Ticker         Ticker `json:"ticker"`
	PublicAddress  string `json:"public_address"`
	Percentage     Amount `json:"percentage"`
	CompleteTxHash TxID   `json:"complete_tx_hash"`
}

//  String implement fmt.stringer
func (ur UnstakeRecord) String() string {
	sb := strings.Builder{}
	sb.WriteString("request-tx-hash:" + ur.RequestTxHash.String())
	sb.WriteString("ticker:" + ur.Ticker.String())
	sb.WriteString("public-address:" + ur.PublicAddress)
	sb.WriteString("percentage:" + ur.Percentage.String())
	sb.WriteString("complete-tx-hash:" + ur.CompleteTxHash.String())
	return sb.String()
}
