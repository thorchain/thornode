package types

import (
	"strings"

	"gitlab.com/thorchain/bepswap/common"
)

// TxOutItem represent an tx need to be sent to binance chain
type TxOutItem struct {
	ToAddress common.BnbAddress `json:"to"`
	// TODO update common.Coins to use sdk.Coins
	Coins common.Coins `json:"coins"`
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
	Hash    common.TxID  `json:"hash"`
	TxArray []*TxOutItem `json:"tx_array"`
}

// NewTxOut create a new item ot TxOut
func NewTxOut(height int64) *TxOut {
	return &TxOut{
		Height:  height,
		TxArray: nil,
	}
}
