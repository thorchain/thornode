package thorclient

import (
	"fmt"

	"github.com/pkg/errors"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

// GetLastObservedInHeight returns the lastobservedin value for the chain past in
func (b *ThorchainBridge) GetLastObservedInHeight(chain common.Chain) (int64, error) {
	lastblock, err := b.getLastBlock(chain)
	if err != nil {
		return 0, errors.Wrap(err, "failed to GetLastObservedInHeight")
	}
	return lastblock.LastChainHeight, nil
}

// GetLastSignedOutHeight returns the lastsignedout value for thorchain
func (b *ThorchainBridge) GetLastSignedOutHeight() (int64, error) {
	lastblock, err := b.getLastBlock("")
	if err != nil {
		return 0, errors.Wrap(err, "failed to GetLastSignedOutHeight")
	}
	return lastblock.LastSignedHeight, nil
}

// GetBlockHeight returns the current height for thorchain blocks
func (b *ThorchainBridge) GetBlockHeight() (int64, error) {
	lastblock, err := b.getLastBlock("")
	if err != nil {
		return 0, errors.Wrap(err, "failed to GetStatechainHeight")
	}
	return lastblock.Statechain, nil
}

// getLastBlock calls the /lastblock/{chain} endpoint and Unmarshal's into the QueryResHeights type
func (b *ThorchainBridge) getLastBlock(chain common.Chain) (types.QueryResHeights, error) {
	url := LastBlockEndpoint
	if chain.String() != "" {
		url = fmt.Sprintf("%s/%s", url, chain.String())
	}
	buf, err := b.get(url)
	if err != nil {
		return types.QueryResHeights{}, errors.Wrap(err, "failed to get lastblock")
	}
	var lastBlock types.QueryResHeights
	if err := b.cdc.UnmarshalJSON(buf, &lastBlock); nil != err {
		b.errCounter.WithLabelValues("fail_unmarshal_lastblock", "").Inc()
		return types.QueryResHeights{}, errors.Wrap(err, "failed to unmarshal last block")
	}
	return lastBlock, nil
}
