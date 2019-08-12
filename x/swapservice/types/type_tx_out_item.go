package types

import (
	"strings"
)

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
