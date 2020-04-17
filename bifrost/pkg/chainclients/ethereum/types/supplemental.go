package types

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
)

// Ethereum supplemental with block scanner methods block request and unmarshal block
type EthereumSupplemental struct{}

func (eth EthereumSupplemental) BlockRequest(rpcHost string, height int64) (string, string) {
	return rpcHost, `{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x` + fmt.Sprintf("%x", height) + `", true],"id":1}`
}

func (eth EthereumSupplemental) UnmarshalBlock(buf []byte) ([]string, error) {
	var head *types.Header
	var body RPCBlock
	if err := json.Unmarshal(buf, &head); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(buf, &body); err != nil {
		return nil, err
	}
	txs := make([]string, 0)
	for _, tx := range body.Transactions {
		bytes, err := tx.Transaction.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("fail to unmarshal tx from block: %w", err)
		}
		txs = append(txs, string(bytes))
	}
	return txs, nil
}
