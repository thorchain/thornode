package common

import sdk "github.com/cosmos/cosmos-sdk/types"

type Fee struct {
	Coins      Coins    `json:"coins"`
	PoolDeduct sdk.Uint `json:"pool_deduct"`
}
