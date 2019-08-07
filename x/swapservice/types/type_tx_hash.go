package types

// Meant to track if we have processed a specific binance tx
type TxHash struct {
	Request  string `json:"request"`  // binance chain request tx hash
	Done     string `json:"done"`     // binacne chain completed tx hash
	Reverted string `json:"reverted"` // hash if we have reverted a request (ie refunded)
}

func NewTxHash(hash string) TxHash {
	return TxHash{
		Request: hash,
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
