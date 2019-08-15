package types

import (
	"fmt"
)

type status string
type TxInIndex []TxID

const (
	Incomplete status = "incomplete"
	Done       status = "done"
	Reverted   status = "reverted"
)

// Meant to track if we have processed a specific binance tx
type TxIn struct {
	Request TxID       `json:"request"` // binance chain request tx hash
	Status  status     `json:"status"`
	Done    TxID       `json:"txhash"` // completed binance chain tx hash
	Memo    string     `json:"memo"`   // memo
	Coins   Coins      `json:"coins"`  // coins sent in tx
	Sender  BnbAddress `json:"sender"`
}

func NewTxIn(hash TxID, coins Coins, memo string, sender BnbAddress) TxIn {
	return TxIn{
		Request: hash,
		Coins:   coins,
		Memo:    memo,
		Sender:  sender,
		Status:  Incomplete,
	}
}

func (tx TxIn) Valid() error {
	if tx.Request.Empty() {
		return fmt.Errorf("Request TxID cannot be empty")
	}
	if tx.Sender.Empty() {
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
	return tx.Request.Empty()
}

func (tx TxIn) String() string {
	return tx.Request.String()
}

// Generate db key for kvstore
func (tx TxIn) Key() TxID {
	return tx.Request
}

func (tx *TxIn) SetDone(hash TxID) {
	tx.Status = Done
	tx.Done = hash
}

func (tx *TxIn) SetReverted(hash TxID) {
	tx.Status = Reverted
	tx.Done = hash
}
