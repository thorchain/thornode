package thorchain

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/bepswap/thornode/cmd"
)

// The block reward emission curve targets a ~2% emission after 10 years (similar to Bitcoin).
// Since RUNE is a strictly-scarce asset, emissions need to be carefully considered for greatest network prosperity.
// Day 0 Emission is ~25%, Year 1 Emission is ~20%.

const emissionCurve = 6         // An arbitrary factor to target desired curve
const secondsPerYear = 31556952 // 365.2425 * 86400

// calculate node account bond units
func calculateNodeAccountBondUints(height, activeBlock, slashPts int64) sdk.Uint {
	if height < 0 || activeBlock < 0 || slashPts < 0 {
		return sdk.ZeroUint()
	}
	blockCount := height - activeBlock
	// Minus slash points
	bCount := blockCount
	if bCount < slashPts {
		bCount = slashPts
	}

	return sdk.NewUint(uint64(bCount - slashPts))
}

// calculate node rewards
func calcNodeRewards(naBlocks, totalUnits, totalRuneReward sdk.Uint) sdk.Uint {
	if totalUnits.Equal(sdk.ZeroUint()) || naBlocks.Equal(sdk.ZeroUint()) {
		return sdk.ZeroUint()
	}
	reward := sdk.NewUint(uint64(
		float64(totalRuneReward.Uint64()) / (float64(totalUnits.Uint64()) / float64(naBlocks.Uint64())),
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
	var amt sdk.Uint
	amt := sdk.NewUint(uint64(math.Round(
		float64(stakerDeficit.Uint64()) / (float64(totalFees.Uint64()) / float64(poolFees.Uint64())),
	)))
	return amt
}

// Calculate the block rewards that bonders and stakers should receive
func calcBlockRewards(totalReserve sdk.Uint, totalLiquidityFees sdk.Uint) (sdk.Uint, sdk.Uint, sdk.Uint) {
	blocksPerYear := secondsPerYear / cmd.SecondsPerBlock
	blockRewards := sdk.NewUint(uint64(
		(float64(totalReserve.Uint64()) / float64(emissionCurve)) / float64(blocksPerYear),
	))

	systemIncome = blockRewards.Add(totalLiquidityFees) // Get total system income for block
	stakerSplit := systemIncome.QuoUint64(3)            // 1/3rd to Stakers
	bonderSplit := systemIncome.Sub(stakerSplit)        // 2/3rd to Bonders

	stakerDeficit := sdk.ZeroUint()
	if stakerSplit >= totalLiquidityFees {
		// Stakers have not been paid enough already, pay more
		poolReward := stakerSplit.Sub(totalLiquidityFees) // Get how much to divert to add to staker split
	} else {
		// Stakers have been paid too much, calculate deficit
		stakerDeficit := totalLiquidityFees.Sub(stakerSplit) // Deduct existing income from split
		poolReward := sdk.ZeroUint()                         // Nothing to pay stakers now
	}

	bondReward := bonderSplit // Give bonders their cut

	return bondReward, poolReward, stakerDeficit
}
