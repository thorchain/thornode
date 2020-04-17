package types

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type RPCBlock struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Result  struct {
		Block struct {
			Header struct {
				Height string `json:"height"`
			} `json:"header"`
			Data struct {
				Txs []string `json:"txs"`
			} `json:"data"`
		} `json:"block"`
	} `json:"result"`
}

func BlockRequest(rpcHost string, height int64) (string, string) {
	u, _ := url.Parse(rpcHost)
	u.Path = "block"
	if height > 0 {
		u.RawQuery = fmt.Sprintf("height=%d", height)
	}
	return u.String(), ""
}

func UnmarshalBlock(buf []byte) (string, []string, error) {
	var block RPCBlock
	err := json.Unmarshal(buf, &block)
	if err != nil {
		return "", nil, fmt.Errorf("fail to unmarshal body to RPCBlock: %w", err)
	}
	return block.Result.Block.Header.Height, block.Result.Block.Data.Txs, nil
}
