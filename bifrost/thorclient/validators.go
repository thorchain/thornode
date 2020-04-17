package thorclient

import (
	"fmt"

	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

// GetValidators returns validators from thorchain
func (b *ThorchainBridge) GetValidators() (*types.ValidatorsResp, error) {
	body, _, err := b.getWithPath(ValidatorsEndpoint)
	if err != nil {
		b.errCounter.WithLabelValues("fail_get_validators", "").Inc()
		return &types.ValidatorsResp{}, fmt.Errorf("failed to get validators: %w", err)
	}
	var vr types.ValidatorsResp
	if err := b.cdc.UnmarshalJSON(body, &vr); err != nil {
		b.errCounter.WithLabelValues("fail_unmarshal_validators", "").Inc()
		return &types.ValidatorsResp{}, fmt.Errorf("failed to unmarshal validators: %w", err)
	}
	return &vr, nil
}
