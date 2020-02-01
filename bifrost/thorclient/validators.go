package thorclient

import (
	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

// GetValidators returns validators from thorchain
func (b *ThorchainBridge) GetValidators() (*types.ValidatorsResp, error) {
	body, err := b.get(ValidatorsEndpoint)
	if err != nil {
		b.errCounter.WithLabelValues("fail_get_validators", "").Inc()
		return &types.ValidatorsResp{}, errors.Wrap(err, "failed to get validators")
	}
	var vr types.ValidatorsResp
	if err := b.cdc.UnmarshalJSON(body, &vr); err != nil {
		b.errCounter.WithLabelValues("fail_unmarshal_validators", "").Inc()
		return &types.ValidatorsResp{}, errors.Wrap(err, "failed to unmarshal validators")
	}
	return &vr, nil
}
