package types

import (
	"fmt"
	"strings"
)

type TxID string

func NewTxID(hash string) (TxID, error) {
	if len(hash) != 64 {
		return TxID(""), fmt.Errorf("TxID Error: Must be 64 characters (got %d)", len(hash))
	}
	return TxID(strings.ToUpper(hash)), nil
}

func (tx TxID) Equals(tx2 TxID) bool {
	return strings.EqualFold(tx.String(), tx2.String())
}

func (tx TxID) IsEmpty() bool {
	return strings.TrimSpace(tx.String()) == ""
}

func (tx TxID) String() string {
	return string(tx)
}
