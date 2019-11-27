package thorchain

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/constants"
)

// calculate node account bond units
func calculateNodeAccountBondUints(height, activeBlock, slashPts int64) sdk.Uint {
	if height < 0 || activeBlock < 0 || slashPts < 0 {
		return sdk.ZeroUint()
	}
	blockCount := height - activeBlock
	// Minus slash pointss
	bCount := blockCount
	if bCount < slashPts {
		bCount = slashPts
	}

	return sdk.NewUint(uint64(bCount - slashPts))
}

// calculate node rewards
func calcNodeRewards(nodeUnits, totalUnits, totalRuneReward sdk.Uint) sdk.Uint {
	reward := sdk.NewUint(uint64(
		float64(totalRuneReward.Uint64()) / (float64(totalUnits.Uint64()) / float64(nodeUnits.Uint64())),
	))

	return reward
}

// Calculate pool rewards
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

// Calculate pool deficit based on the pool's accrued fees compared with total fees.
func calcPoolDeficit(stakerDeficit, totalFees sdk.Uint, poolFees sdk.Uint) sdk.Uint {
	//var amt sdk.Uint
	amt := sdk.NewUint(uint64(math.Round(
		float64(stakerDeficit.Uint64()) / (float64(totalFees.Uint64()) / float64(poolFees.Uint64())),
	)))

	return amt
}

// Calculate the block rewards that bonders and stakers should receive
func calcBlockRewards(totalReserve sdk.Uint, totalLiquidityFees sdk.Uint) (sdk.Uint, sdk.Uint, sdk.Uint) {
	// Block Rewards will take the latest reserve, divide it by the emission curve factor, then divide by blocks per year
	blockReward := sdk.NewUint(uint64(
		(float64(totalReserve.Uint64()) / float64(constants.EmissionCurve)) / float64(constants.BlocksPerYear),
	))

	systemIncome := blockReward.Add(totalLiquidityFees) // Get total system income for block
	stakerSplit := systemIncome.QuoUint64(3)            // 1/3rd to Stakers
	bonderSplit := systemIncome.Sub(stakerSplit)        // 2/3rd to Bonders

	stakerDeficit := sdk.ZeroUint()
	poolReward := sdk.ZeroUint()

	if stakerSplit.GTE(totalLiquidityFees) {
		// Stakers have not been paid enough already, pay more
		poolReward = stakerSplit.Sub(totalLiquidityFees) // Get how much to divert to add to staker split
	} else {
		// Stakers have been paid too much, calculate deficit
		stakerDeficit = totalLiquidityFees.Sub(stakerSplit) // Deduct existing income from split
	}

	bondReward := bonderSplit // Give bonders their split

	return bondReward, poolReward, stakerDeficit
}
