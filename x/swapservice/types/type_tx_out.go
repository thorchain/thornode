package types

import "strings"

// TxOutItem represent an tx need to be sent to binance chain
type TxOutItem struct {
	ToAddress BnbAddress `json:"to"`
	Coins     []Coin     `json:"coins"`
}

// String implement stringer interface
func (toi TxOutItem) String() string {
	sb := strings.Builder{}
	sb.WriteString("to address:" + toi.ToAddress.String())
	for _, c := range toi.Coins {
		sb.WriteString("denom:" + c.Denom.String())
		sb.WriteString("Amount:" + c.Amount.String())
	}
	return sb.String()
}

// TxOut is a structure represent all the tx we need to return to client
type TxOut struct {
	Height  int64        `json:"height"`
	Hash    TxID         `json:"hash"`
	TxArray []*TxOutItem `json:"tx_array"`
}

// NewTxOut create a new item ot TxOut
func NewTxOut(height int64) *TxOut {
	return &TxOut{
		Height:  height,
		TxArray: nil,
	}
}
