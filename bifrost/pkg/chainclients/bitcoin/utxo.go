package bitcoin

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// UnspentTransactionOutput
type UnspentTransactionOutput struct {
	TxID  chainhash.Hash `json:"tx_id"`
	N     uint32         `json:"n"`
	Value float64        `json:"value"`
}

// NewUnspentTransactionOutput create a new instance of UnspentTransactionOutput
func NewUnspentTransactionOutput(txID chainhash.Hash, n uint32, value float64) UnspentTransactionOutput {
	return UnspentTransactionOutput{
		TxID:  txID,
		N:     n,
		Value: value,
	}
}

// GetKey return a key
func (t UnspentTransactionOutput) GetKey() string {
	return fmt.Sprintf("%s:%d", t.TxID.String(), t.N)
}
