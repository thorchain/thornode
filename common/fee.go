package common

import sdk "github.com/cosmos/cosmos-sdk/types"

type Fee struct {
	Coins      Coins    `json:"coins"`
	PoolDeduct sdk.Uint `json:"pool_deduct"`
}

// NewFee return a new instance of Fee
func NewFee(coins Coins, poolDeduct sdk.Uint) Fee {
	return Fee{
		Coins:      coins,
		PoolDeduct: poolDeduct,
	}
}

// EmptyFee return a empty instance of Fee
func EmptyFee() Fee {
	return Fee{
		PoolDeduct: sdk.ZeroUint(),
	}
}
