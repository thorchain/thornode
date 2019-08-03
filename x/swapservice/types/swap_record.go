package types

import "strings"

// SwapRecord is
type SwapRecord struct {
	RequestTxHash   string `json:"request_tx_hash"`  // The TxHash on binance chain represent user send token to the pool
	SourceTicker    string `json:"source_ticker"`    // Source ticker
	TargetTicker    string `json:"target_ticker"`    // Target ticker
	Requester       string `json:"requester"`        // Requester , should be the address on binance chain
	Destination     string `json:"destination"`      // destination , used for swap and send , the destination address we send it to
	AmountRequested string `json:"amount_requested"` // amount of source token in
	AmountPaidBack  string `json:"amount_paid_back"` // amount of target token pay out to user
	PayTxHash       string `json:"pay_tx_hash"`      // TxHash on binance chain represent our pay to user
}

// String implement stringer interface
func (sr SwapRecord) String() string {
	sb := strings.Builder{}
	sb.WriteString("request-txhash:" + sr.RequestTxHash)
	sb.WriteString("source-ticker:" + sr.SourceTicker)
	sb.WriteString("target-ticker:" + sr.TargetTicker)
	sb.WriteString("requester-address:" + sr.Requester)
	sb.WriteString("destination:" + sr.Destination)
	sb.WriteString("amount:" + sr.AmountRequested)
	sb.WriteString("amount-pay-to-user:" + sr.AmountPaidBack)
	return sb.String()
}
