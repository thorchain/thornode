package types

import (
	"fmt"

	common "gitlab.com/thorchain/bepswap/common"
)

type status string
type TxInIndex []common.TxID

const (
	Incomplete status = "incomplete"
	Done       status = "done"
	Reverted   status = "reverted"
)

// Meant to track if we have processed a specific binance tx
type TxIn struct {
	Request common.TxID       `json:"request"` // binance chain request tx hash
	Status  status            `json:"status"`
	Done    common.TxID       `json:"txhash"` // completed binance chain tx hash
	Memo    string            `json:"memo"`   // memo
	Coins   common.Coins      `json:"coins"`  // coins sent in tx
	Sender  common.BnbAddress `json:"sender"`
}

func NewTxIn(hash common.TxID, coins common.Coins, memo string, sender common.BnbAddress) TxIn {
	return TxIn{
		Request: hash,
		Coins:   coins,
		Memo:    memo,
		Sender:  sender,
		Status:  Incomplete,
	}
}

func (tx TxIn) Valid() error {
	if tx.Request.IsEmpty() {
		return fmt.Errorf("Request TxID cannot be empty")
	}
	if tx.Sender.IsEmpty() {
		return fmt.Errorf("Sender cannot be empty")
	}
	if len(tx.Coins) == 0 {
		return fmt.Errorf("Coins cannot be empty")
	}
	if len(tx.Memo) == 0 {
		return fmt.Errorf("Memo cannot be empty")
	}

	return nil
}

func (tx TxIn) Empty() bool {
	return tx.Request.IsEmpty()
}

func (tx TxIn) String() string {
	return tx.Request.String()
}

// Generate db key for kvstore
func (tx TxIn) Key() common.TxID {
	return tx.Request
}

func (tx *TxIn) SetDone(hash common.TxID) {
	tx.Status = Done
	tx.Done = hash
}

func (tx *TxIn) SetReverted(hash common.TxID) {
	tx.Status = Reverted
	tx.Done = hash
}
