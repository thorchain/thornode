package thorchain

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/cmd"
)

// calculate pool rewards
func calcPoolRewards(totalPoolRewards, totalStakedRune sdk.Uint, pools []Pool) []sdk.Uint {
	var amts []sdk.Uint
	for _, pool := range pools {
		amt := sdk.NewUint(uint64(math.Round(
			float64(totalPoolRewards.Uint64()) / (float64(totalStakedRune.Uint64()) / float64(pool.BalanceRune.Uint64())),
		)))
		amts = append(amts, amt)
	}

	return amts
}

// calculate the block rewards that bonders and stakers should receive
func calcBlockRewards(totalReserve sdk.Uint) (sdk.Uint, sdk.Uint) {
	blocksPerYear := 31536000 / cmd.SecondsPerBlock
	blockRewards := sdk.NewUint(uint64(
		(float64(totalReserve.Uint64()) / float64(5)) / float64(blocksPerYear),
	))
	poolReward := blockRewards.QuoUint64(3)
	bondReward := blockRewards.Sub(poolReward)

	return bondReward, poolReward
}
