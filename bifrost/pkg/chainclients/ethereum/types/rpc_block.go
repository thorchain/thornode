package types

import (
	"encoding/json"
	"fmt"

	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

func BlockRequest(rpcHost string, height int64) (string, string) {
	return rpcHost, `{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x` + fmt.Sprintf("%x", height) + `", true],"id":1}`
}

func UnmarshalBlock(buf []byte) (string, []string, error) {
	var block etypes.Block
	err := json.Unmarshal(buf, &block)
	if err != nil {
		return "", nil, errors.Wrap(err, "fail to unmarshal body to RPCBlock")
	}
	txs := make([]string, 0)
	for _, tx := range block.Transactions() {
		bytes, err := tx.MarshalJSON()
		if err != nil {
			return "", nil, errors.Wrap(err, "fail to unmarshal tx from block")
		}
		txs = append(txs, string(bytes))
	}
	return block.Number().String(), txs, nil
}
