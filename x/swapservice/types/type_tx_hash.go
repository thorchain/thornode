package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type status string

const (
	Incomplete status = "incomplete"
	Done       status = "done"
	Reverted   status = "reverted"
)

// Meant to track if we have processed a specific binance tx
type TxHash struct {
	Request string    `json:"request"` // binance chain request tx hash
	Status  status    `json:"status"`
	Done    string    `json:"txhash"` // completed binance chain tx hash
	Memo    string    `json:"memo"`   // memo
	Coins   sdk.Coins `json:"coins"`  // coins sent in tx
	Sender  string    `json:"sender"`
}

func NewTxHash(hash string, coins sdk.Coins, memo, sender string) TxHash {
	return TxHash{
		Request: hash,
		Coins:   coins,
		Memo:    memo,
		Sender:  sender,
		Status:  Incomplete,
	}
}

func (tx TxHash) Empty() bool {
	return tx.Request == ""
}

func (tx TxHash) String() string {
	return tx.Request
}

// Generate db key for kvstore
func (tx TxHash) Key() string {
	return tx.Request
}

func (tx *TxHash) SetDone(hash string) {
	tx.Status = Done
	tx.Done = hash
}

func (tx *TxHash) SetReverted(hash string) {
	tx.Status = Reverted
	tx.Done = hash
}
