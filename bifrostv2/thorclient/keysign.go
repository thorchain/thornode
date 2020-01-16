package thorclient

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

// GetKeysign retrieves txout from this block height from thorchain
func (c *Client) GetKeysign(blockHeight int64, pk string) (*types.TxOut, error) {
	url := fmt.Sprintf("%s/%d/%s", KeysignEndpoint, blockHeight, pk)
	body, err := c.get(url)

	if err != nil {
		c.errCounter.WithLabelValues("fail_get_tx_out", strconv.FormatInt(blockHeight, 10)).Inc()
		return &types.TxOut{}, errors.Wrap(err, "failed to get tx from a block height")
	}
	var txOut types.TxOut
	if err := c.cdc.UnmarshalJSON(body, &txOut); err != nil {
		c.errCounter.WithLabelValues("fail_unmarshal_tx_out", strconv.FormatInt(blockHeight, 10)).Inc()
		return &types.TxOut{}, errors.Wrap(err, "failed to unmarshal TxOut")
	}
	return &txOut, nil
}
