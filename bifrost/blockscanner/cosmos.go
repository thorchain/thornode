package blockscanner

// This implementation is design for cosmos based blockchains

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

type itemData struct {
	Txs []string `json:"txs"`
}

type itemHeader struct {
	Height string `json:"height"`
}

type itemBlock struct {
	Header itemHeader `json:"header"`
	Data   itemData   `json:"data"`
}

type itemResult struct {
	Block itemBlock `json:"block"`
}

type item struct {
	Jsonrpc string     `json:"jsonrpc"`
	ID      string     `json:"id"`
	Result  itemResult `json:"result"`
}

type CosmosSupplemental struct{}

func (cosmos CosmosSupplemental) BlockRequest(rpcHost string, height int64) (string, string) {
	u, _ := url.Parse(rpcHost)
	u.Path = "block"
	if height > 0 {
		u.RawQuery = fmt.Sprintf("height=%d", height)
	}
	return u.String(), ""
}

func (cosmos CosmosSupplemental) UnmarshalBlock(buf []byte) (int64, []string, error) {
	var block item
	err := json.Unmarshal(buf, &block)
	if err != nil {
		return 0, nil, errors.Wrap(err, "fail to unmarshal body to rpcBlock")
	}

	height, err := strconv.ParseInt(block.Result.Block.Header.Height, 10, 64)
	if err != nil {
		return 0, nil, errors.Wrap(err, "fail to convert block height to int")
	}
	return height, block.Result.Block.Data.Txs, nil
}
