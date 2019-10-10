package types

import (
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/common"
)

// TxOutItem represent an tx need to be sent to binance chain
type TxOutItem struct {
	ToAddress   common.BnbAddress `json:"to"`
	PoolAddress common.BnbAddress `json:"pool_address"`
	// TODO update common.Coins to use sdk.Coins
	Coins common.Coins `json:"coins"`
}

func (toi TxOutItem) Valid() error {
	if toi.ToAddress.IsEmpty() {
		return errors.New("To address cannot be empty")
	}

	return nil
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
	Height  uint64       `json:"height"`
	Hash    common.TxID  `json:"hash"`
	TxArray []*TxOutItem `json:"tx_array"`
}

// NewTxOut create a new item ot TxOut
func NewTxOut(height uint64) *TxOut {
	return &TxOut{
		Height:  height,
		TxArray: nil,
	}
}

func (out TxOut) Valid() error {

	for _, tx := range out.TxArray {
		if err := tx.Valid(); err != nil {
			return err
		}
	}

	return nil
}
