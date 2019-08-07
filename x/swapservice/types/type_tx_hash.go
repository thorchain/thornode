package types

type status string

const (
	Incomplete status = "incomplete"
	Done       status = "done"
	Reverted   status = "reverted"
)

// Meant to track if we have processed a specific binance tx
type TxHash struct {
	Request string `json:"request"` // binance chain request tx hash
	Status  status `json:"status"`
	Hash    string `json:"txhash"` // completed binance chain tx hash
}

func NewTxHash(hash string) TxHash {
	return TxHash{
		Request: hash,
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
	tx.Hash = hash
}

func (tx *TxHash) SetReverted(hash string) {
	tx.Status = Reverted
	tx.Hash = hash
}
