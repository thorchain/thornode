package types

import "strings"

// UnstakeRecord is a record about unstake activity
type UnstakeRecord struct {
	RequestTxHash  string `json:"request_tx_hash"`
	Ticker         string `json:"ticker"`
	PublicAddress  string `json:"public_address"`
	Percentage     string `json:"percentage"`
	CompleteTxHash string `json:"complete_tx_hash"`
}

//  String implement fmt.stringer
func (ur UnstakeRecord) String() string {
	sb := strings.Builder{}
	sb.WriteString("request-tx-hash:" + ur.RequestTxHash)
	sb.WriteString("ticker:" + ur.Ticker)
	sb.WriteString("public-address:" + ur.PublicAddress)
	sb.WriteString("percentage:" + ur.Percentage)
	sb.WriteString("complete-tx-hash:" + ur.CompleteTxHash)
	return sb.String()
}
