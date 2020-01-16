package thorclient

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

// GetKeygens retrieves keygens from this block height from thorchain
func (c *Client) GetKeygens(blockHeight int64, pk string) (*types.Keygens, error) {
	url := fmt.Sprintf("%s/%d/%s", KeygenEndpoint, blockHeight, pk)
	body, err := c.get(url)

	if err != nil {
		c.errCounter.WithLabelValues("fail_get_keygens", strconv.FormatInt(blockHeight, 10)).Inc()
		return &types.Keygens{}, errors.Wrap(err, "failed to get keygens from a block height")
	}
	var keygens types.Keygens
	if err := c.cdc.UnmarshalJSON(body, &keygens); err != nil {
		c.errCounter.WithLabelValues("fail_unmarshal_keygens", strconv.FormatInt(blockHeight, 10)).Inc()
		return &types.Keygens{}, errors.Wrap(err, "failed to unmarshal Keygens")
	}
	return &keygens, nil
}
