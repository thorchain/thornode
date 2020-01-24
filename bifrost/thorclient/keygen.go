package thorclient

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
)

// GetKeygens retrieves keygens from this block height from thorchain
func (b *ThorchainBridge) GetKeygens(blockHeight int64, pk string) (*types.Keygens, error) {
	url := fmt.Sprintf("%s/%d/%s", KeygenEndpoint, blockHeight, pk)
	body, err := b.get(url)

	if err != nil {
		b.errCounter.WithLabelValues("fail_get_keygens", strconv.FormatInt(blockHeight, 10)).Inc()
		return &types.Keygens{}, errors.Wrap(err, "failed to get keygens from a block height")
	}
	var keygens types.Keygens
	if err := b.cdc.UnmarshalJSON(body, &keygens); err != nil {
		b.errCounter.WithLabelValues("fail_unmarshal_keygens", strconv.FormatInt(blockHeight, 10)).Inc()
		return &types.Keygens{}, errors.Wrap(err, "failed to unmarshal Keygens")
	}
	return &keygens, nil
}
