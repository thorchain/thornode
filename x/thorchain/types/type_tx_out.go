package types

import (
	"strings"

	"github.com/pkg/errors"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

// TxOutItem represent an tx need to be sent to chain
type TxOutItem struct {
	Chain       common.Chain   `json:"chain"`
	ToAddress   common.Address `json:"to"`
	PoolAddress common.PubKey  `json:"pool_address"`
	SeqNo       uint64         `json:"seq_no"`
	Coin        common.Coin    `json:"coin"`
	Memo        string         `json:"memo"`
}

func (toi TxOutItem) Valid() error {
	if toi.Chain.IsEmpty() {
		return errors.New("chain cannot be empty")
	}
	if toi.ToAddress.IsEmpty() {
		return errors.New("To address cannot be empty")
	}
	if toi.PoolAddress.IsEmpty() {
		return errors.New("pool address cannot be empty")
	}
	if err := toi.Coin.IsValid(); err != nil {
		return err
	}
	return nil
}

// String implement stringer interface
func (toi TxOutItem) String() string {
	sb := strings.Builder{}
	sb.WriteString("To Address:" + toi.ToAddress.String())
	sb.WriteString("Asset:" + toi.Coin.Asset.String())
	sb.WriteString("Amount:" + toi.Coin.Amount.String())
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
		Height: height,
	}
}

// IsEmpty to determinate whether there are txitm in this TxOut
func (out TxOut) IsEmpty() bool {
	return len(out.TxArray) == 0
}

// Valid check every item in it's internal txarray, return an error if it is not valid
func (out TxOut) Valid() error {
	for _, tx := range out.TxArray {
		if err := tx.Valid(); err != nil {
			return err
		}
	}
	return nil
}
