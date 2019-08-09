package types

import (
	"strings"
)

// TxOutItem represent an tx need to be sent to binance chain
type TxOutItem struct {
	ToAddress string `json:"to"`
	// TODO later on refractor this to use sdk.Coin
	Coins []Coin `json:"coins"`
}

// String implement stringer interface
func (toi TxOutItem) String() string {
	sb := strings.Builder{}
	sb.WriteString("to address:" + toi.ToAddress)
	for _, c := range toi.Coins {
		sb.WriteString("denom:" + c.Denom)
		sb.WriteString("Amount:" + c.Amount)
	}
	return sb.String()
}
