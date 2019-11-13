package thorchain

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/cmd"
)

// With feedback from the community, the block reward emission curve has been
// modified to target 2% emission after 10 years (similiar to Bitcoin).
// This reduces the return in the first year, but allows it to be spread
// forward into the future. Since RUNE is a strictly-scarce fixed supply asset,
// emissions need to be carefully considered for greatest network prosperity.
// Day 0 Emission is now 25%, Year 1 Emission is 20%.
const emissionCurve = 5

// calculate node rewards
func calcNodeRewards(naBlocks, totalUnits, totalRuneReward sdk.Uint) sdk.Uint {
	reward := sdk.NewUint(uint64(
		float64(totalRuneReward.Uint64()) / (float64(totalUnits.Uint64()) / float64(naBlocks.Uint64())),
	))
	return reward
}

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
		(float64(totalReserve.Uint64()) / float64(emissionCurve)) / float64(blocksPerYear),
	))
	poolReward := blockRewards.QuoUint64(3)
	bondReward := blockRewards.Sub(poolReward)

	return bondReward, poolReward
}
