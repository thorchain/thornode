package thorclient

import (
	"fmt"

	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

// GetNodeAccount retrieves node account for this address from thorchain
func (b *ThorchainBridge) GetNodeAccount(thorAddr string) (*types.NodeAccount, error) {
	path := fmt.Sprintf("%s/%s", NodeAccountEndpoint, thorAddr)
	body, _, err := b.getWithPath(path)
	if err != nil {
		b.errCounter.WithLabelValues("fail_get_node_account", thorAddr).Inc()
		return &types.NodeAccount{}, fmt.Errorf("failed to get node account: %w", err)
	}
	var na types.NodeAccount
	if err := b.cdc.UnmarshalJSON(body, &na); err != nil {
		b.errCounter.WithLabelValues("fail_unmarshal_node_account", thorAddr).Inc()
		return &types.NodeAccount{}, fmt.Errorf("failed to unmarshal node account: %w", err)
	}
	return &na, nil
}
