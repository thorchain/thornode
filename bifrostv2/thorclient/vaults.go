package thorclient

import (
	"encoding/json"

	"github.com/pkg/errors"

	"gitlab.com/thorchain/thornode/bifrostv2/thorclient/types"
)

func (c *Client) GetVaults() (types.Vaults, error) {
	buf, err := c.get(BaseEndpoint + VaultsEndpoint)
	if err != nil {
		return types.Vaults{}, errors.Wrap(err, "fail to get from thorchain")
	}

	var vaults types.Vaults
	if err := json.Unmarshal(buf, &vaults); err != nil {
		return types.Vaults{}, errors.Wrap(err, "failed to unmarshal vaults")
	}

	return vaults, nil
}
