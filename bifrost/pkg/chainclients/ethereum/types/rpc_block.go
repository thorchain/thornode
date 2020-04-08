package types

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/core/types"
)

type txExtraInfo struct {
	BlockNumber string `json:"blockNumber,omitempty"`
	BlockHash   string `json:"blockHash,omitempty"`
	From        string `json:"from,omitempty"`
}

type RPCTransaction struct {
	Transaction *types.Transaction
	txExtraInfo
}

func (tx *RPCTransaction) UnmarshalJSON(msg []byte) error {
	if err := json.Unmarshal(msg, &tx.Transaction); err != nil {
		return err
	}
	return json.Unmarshal(msg, &tx.txExtraInfo)
}

type RPCBlock struct {
	Hash         string           `json:"hash"`
	Transactions []RPCTransaction `json:"transactions"`
	UncleHashes  []string         `json:"uncles"`
}
