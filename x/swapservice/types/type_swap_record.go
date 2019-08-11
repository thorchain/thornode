package types

import "strings"

// SwapRecord is
type SwapRecord struct {
	RequestTxHash   TxID       `json:"request_tx_hash"`  // The TxHash on binance chain represent user send token to the pool
	SourceTicker    Ticker     `json:"source_ticker"`    // Source ticker
	TargetTicker    Ticker     `json:"target_ticker"`    // Target ticker
	Requester       BnbAddress `json:"requester"`        // Requester , should be the address on binance chain
	Destination     BnbAddress `json:"destination"`      // destination , used for swap and send , the destination address we send it to
	AmountRequested Amount     `json:"amount_requested"` // amount of source token in
	AmountPaidBack  Amount     `json:"amount_paid_back"` // amount of target token pay out to user
	PayTxHash       TxID       `json:"pay_tx_hash"`      // TxHash on binance chain represent our pay to user
}

func NewSwapRecord(reqTxHash TxID, source, target Ticker, from, to BnbAddress, amtReq, amtPaid Amount, payTxHash TxID) SwapRecord {
	return SwapRecord{
		RequestTxHash:   reqTxHash,
		SourceTicker:    source,
		TargetTicker:    target,
		Requester:       from,
		Destination:     to,
		AmountRequested: amtReq,
		AmountPaidBack:  amtPaid,
		PayTxHash:       payTxHash,
	}
}

// String implement stringer interface
func (sr SwapRecord) String() string {
	sb := strings.Builder{}
	sb.WriteString("request-txhash:" + sr.RequestTxHash.String())
	sb.WriteString("source-ticker:" + sr.SourceTicker.String())
	sb.WriteString("target-ticker:" + sr.TargetTicker.String())
	sb.WriteString("requester-address:" + sr.Requester.String())
	sb.WriteString("destination:" + sr.Destination.String())
	sb.WriteString("amount:" + sr.AmountRequested.String())
	sb.WriteString("amount-pay-to-user:" + sr.AmountPaidBack.String())
	return sb.String()
}
