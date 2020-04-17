package thorclient

import (
	"fmt"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

// GetLastObservedInHeight returns the lastobservedin value for the chain past in
func (b *ThorchainBridge) GetLastObservedInHeight(chain common.Chain) (int64, error) {
	lastblock, err := b.getLastBlock(chain)
	if err != nil {
		return 0, fmt.Errorf("failed to GetLastObservedInHeight: %w", err)
	}
	return lastblock.LastChainHeight, nil
}

// GetLastSignedOutHeight returns the lastsignedout value for thorchain
func (b *ThorchainBridge) GetLastSignedOutHeight() (int64, error) {
	lastblock, err := b.getLastBlock("")
	if err != nil {
		return 0, fmt.Errorf("failed to GetLastSignedOutHeight: %w", err)
	}
	return lastblock.LastSignedHeight, nil
}

// GetBlockHeight returns the current height for thorchain blocks
func (b *ThorchainBridge) GetBlockHeight() (int64, error) {
	lastblock, err := b.getLastBlock("")
	if err != nil {
		return 0, fmt.Errorf("failed to GetStatechainHeight: %w", err)
	}
	return lastblock.Statechain, nil
}

// getLastBlock calls the /lastblock/{chain} endpoint and Unmarshal's into the QueryResHeights type
func (b *ThorchainBridge) getLastBlock(chain common.Chain) (types.QueryResHeights, error) {
	path := LastBlockEndpoint
	if chain.String() != "" {
		path = fmt.Sprintf("%s/%s", path, chain.String())
	}
	buf, _, err := b.getWithPath(path)
	if err != nil {
		return types.QueryResHeights{}, fmt.Errorf("failed to get lastblock: %w", err)
	}
	var lastBlock types.QueryResHeights
	if err := b.cdc.UnmarshalJSON(buf, &lastBlock); err != nil {
		b.errCounter.WithLabelValues("fail_unmarshal_lastblock", "").Inc()
		return types.QueryResHeights{}, fmt.Errorf("failed to unmarshal last block: %w", err)
	}
	return lastBlock, nil
}
