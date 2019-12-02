package thorchain

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// Calculate pool rewards
func calcPoolRewards(totalPoolRewards, totalStakedRune sdk.Uint, pools []Pool) []sdk.Uint {
	var amts []sdk.Uint
	for _, pool := range pools {
		amt := common.GetShare(pool.BalanceRune, totalStakedRune, totalPoolRewards)
		amts = append(amts, amt)
	}
	return amts
}

// Calculate pool deficit based on the pool's accrued fees compared with total fees.
func calcPoolDeficit(stakerDeficit, totalFees sdk.Uint, poolFees sdk.Uint) sdk.Uint {
	return common.GetShare(poolFees, totalFees, stakerDeficit)
}

// Calculate the block rewards that bonders and stakers should receive
func calcBlockRewards(totalReserve sdk.Uint, totalLiquidityFees sdk.Uint) (sdk.Uint, sdk.Uint, sdk.Uint) {
	// Block Rewards will take the latest reserve, divide it by the emission curve factor, then divide by blocks per year
	blockReward := sdk.NewUint(uint64(math.Round(
		(float64(totalReserve.Uint64()) / float64(constants.EmissionCurve)) / float64(constants.BlocksPerYear),
	)))

	systemIncome := blockReward.Add(totalLiquidityFees)      // Get total system income for block
	stakerSplit := systemIncome.QuoUint64(3)                 // 1/3rd to Stakers
	bonderSplit := common.SafeSub(systemIncome, stakerSplit) // 2/3rd to Bonders

	stakerDeficit := sdk.ZeroUint()
	poolReward := sdk.ZeroUint()

	if stakerSplit.GTE(totalLiquidityFees) {
		// Stakers have not been paid enough already, pay more
		poolReward = common.SafeSub(stakerSplit, totalLiquidityFees) // Get how much to divert to add to staker split
	} else {
		// Stakers have been paid too much, calculate deficit
		stakerDeficit = common.SafeSub(totalLiquidityFees, stakerSplit) // Deduct existing income from split
	}

	bondReward := bonderSplit // Give bonders their split

	return bondReward, poolReward, stakerDeficit
}
