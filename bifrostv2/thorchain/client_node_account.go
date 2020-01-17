package thorchain

import (
	"fmt"

	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

// GetNodeAccount retrieves node account for this address from thorchain
func (c *Client) GetNodeAccount(thorAddr string) (*types.NodeAccount, error) {
	url := fmt.Sprintf("%s/%s", NodeAccountEndpoint, thorAddr)
	body, err := c.get(url)

	if err != nil {
		c.errCounter.WithLabelValues("fail_get_node_account", thorAddr).Inc()
		return &types.NodeAccount{}, errors.Wrap(err, "failed to get node account")
	}
	var na types.NodeAccount
	if err := c.cdc.UnmarshalJSON(body, &na); err != nil {
		c.errCounter.WithLabelValues("fail_unmarshal_node_account", thorAddr).Inc()
		return &types.NodeAccount{}, errors.Wrap(err, "failed to unmarshal node account")
	}
	return &na, nil
}
