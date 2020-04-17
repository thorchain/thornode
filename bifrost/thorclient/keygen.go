package thorclient

import (
	"fmt"
	"net/http"
	"strconv"

	btypes "gitlab.com/thorchain/thornode/bifrost/blockscanner/types"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

// GetKeygen retrieves keygen request for the given block height from thorchain
func (b *ThorchainBridge) GetKeygenBlock(blockHeight int64, pk string) (types.KeygenBlock, error) {
	path := fmt.Sprintf("%s/%d/%s", KeygenEndpoint, blockHeight, pk)
	body, status, err := b.getWithPath(path)
	if err != nil {
		if status == http.StatusNotFound {
			return types.KeygenBlock{}, btypes.UnavailableBlock
		}
		b.errCounter.WithLabelValues("fail_get_keygen", strconv.FormatInt(blockHeight, 10)).Inc()
		return types.KeygenBlock{}, fmt.Errorf("failed to get keygen for a block height: %w", err)
	}
	var keygen types.KeygenBlock
	if err := b.cdc.UnmarshalJSON(body, &keygen); err != nil {
		b.errCounter.WithLabelValues("fail_unmarshal_keygen", strconv.FormatInt(blockHeight, 10)).Inc()
		return types.KeygenBlock{}, fmt.Errorf("failed to unmarshal Keygen: %w", err)
	}
	return keygen, nil
}
