package thorclient

import (
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

// GetKeygen retrieves keygen request for the given block height from thorchain
func (b *ThorchainBridge) GetKeygenBlock(blockHeight int64, pk string) (*types.KeygenBlock, error) {
	url := fmt.Sprintf("%s/%d/%s", KeygenEndpoint, blockHeight, pk)
	body, status, err := b.get(url)
	if err != nil {
		if status == http.StatusNotFound {
			return nil, nil
		}
		b.errCounter.WithLabelValues("fail_get_keygen", strconv.FormatInt(blockHeight, 10)).Inc()
		return nil, fmt.Errorf("failed to get keygen for a block height: %w", err)
	}
	var keygen types.KeygenBlock
	if err := b.cdc.UnmarshalJSON(body, &keygen); err != nil {
		b.errCounter.WithLabelValues("fail_unmarshal_keygen", strconv.FormatInt(blockHeight, 10)).Inc()
		return nil, fmt.Errorf("failed to unmarshal Keygen: %w", err)
	}
	return &keygen, nil
}
