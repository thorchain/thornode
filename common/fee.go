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

// Asset retun asset name of fee coins
func (fee *Fee) Asset() Asset {
	for _, coin := range fee.Coins {
		if !coin.Asset.IsRune() {
			return coin.Asset
		}
	}
	return Asset{}
}
