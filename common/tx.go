package common

import (
	"fmt"
	"strings"
)

type TxID string
type TxIDs []TxID

func NewTxID(hash string) (TxID, error) {
	switch len(hash) {
	case 64:
		// do nothing
	case 66: // ETH check
		if !strings.HasPrefix(hash, "0x") {
			err := fmt.Errorf("TxID Error: Must be 66 characters (got %d)", len(hash))
			return TxID(""), err
		}
	default:
		err := fmt.Errorf("TxID Error: Must be 64 characters (got %d)", len(hash))
		return TxID(""), err
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

type Tx struct {
	ID          TxID
	FromAddress Address
	ToAddress   Address
	Coins       Coins
	Memo        string
}

func NewTx(txID TxID, from Address, to Address, coins Coins, memo string) Tx {
	return Tx{
		ID:          txID,
		FromAddress: from,
		ToAddress:   to,
		Coins:       coins,
		Memo:        memo,
	}
}

func (tx Tx) IsEmpty() bool {
	return tx.ID.IsEmpty()
}
