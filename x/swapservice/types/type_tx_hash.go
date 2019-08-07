package types

// Meant to track if we have processed a specific binance tx
type TxHash struct {
	TxHash   string `json:"tx_hash"`  // binance chain tx hash
	Reverted bool   `json:"reverted"` // if we have reverted an event (ie refunded)
}

func NewTxHash(hash string) TxHash {
	return TxHash{
		TxHash: hash,
	}
}

func (tx TxHash) Empty() bool {
	return tx.TxHash == ""
}

func (tx TxHash) String() string {
	return tx.TxHash
}

// Generate db key for kvstore
func (tx TxHash) Key() string {
	return tx.TxHash
}

// Set reverted to true (there is no setting reverted to false)
func (tx *TxHash) SetReverted() {
	tx.Reverted = true
}
