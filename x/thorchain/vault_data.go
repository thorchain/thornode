package thorchain

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/constants"
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
func calcNodeRewards(naBlocks, totalUnits, totalRuneReward sdk.Uint) sdk.Uint {
	if totalUnits.Equal(sdk.ZeroUint()) || naBlocks.Equal(sdk.ZeroUint()) {
		return sdk.ZeroUint()
	}
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
	blockRewards := sdk.NewUint(uint64(
		(float64(totalReserve.Uint64()) / float64(constants.EmissionCurve)) / float64(constants.BlocksPerYear),
	))
	poolReward := blockRewards.QuoUint64(3)
	bondReward := blockRewards.Sub(poolReward)

	return bondReward, poolReward
}
