package thorclient

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
)

// GetKeysign retrieves txout from this block height from thorchain
func (b *ThorchainBridge) GetKeysign(blockHeight int64, pk string) (*types.ChainsTxOut, error) {
	url := fmt.Sprintf("%s/%d/%s", KeysignEndpoint, blockHeight, pk)
	body, err := b.get(url)
	if err != nil {
		b.errCounter.WithLabelValues("fail_get_tx_out", strconv.FormatInt(blockHeight, 10)).Inc()
		return &types.ChainsTxOut{}, errors.Wrap(err, "failed to get tx from a block height")
	}
	var txOut types.ChainsTxOut
	if err := json.Unmarshal(body, &txOut); err != nil {
		b.errCounter.WithLabelValues("fail_unmarshal_tx_out", strconv.FormatInt(blockHeight, 10)).Inc()
		return &types.ChainsTxOut{}, errors.Wrap(err, "failed to unmarshal TxOut")
	}
	return &txOut, nil
}
