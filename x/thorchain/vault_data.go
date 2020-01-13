package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
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
func calcBlockRewards(totalStaked, totalBonded, totalReserve, totalLiquidityFees sdk.Uint, emissionCurve, blocksOerYear int64) (sdk.Uint, sdk.Uint, sdk.Uint) {
	// Block Rewards will take the latest reserve, divide it by the emission curve factor, then divide by blocks per year
	trD := sdk.NewDec(int64(totalReserve.Uint64()))
	ecD := sdk.NewDec(emissionCurve)
	bpyD := sdk.NewDec(blocksOerYear)
	blockRewardD := trD.Quo(ecD).Quo(bpyD)
	blockReward := sdk.NewUint(uint64((blockRewardD).RoundInt64()))

	systemIncome := blockReward.Add(totalLiquidityFees)                 // Get total system income for block
	stakerSplit := getPoolShare(totalStaked, totalBonded, systemIncome) // Get staker share
	bonderSplit := common.SafeSub(systemIncome, stakerSplit)            // Remainder to Bonders

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

func getPoolShare(totalStaked, totalBonded, totalRewards sdk.Uint) sdk.Uint {
	// Targets a linear change in rewards from 0% staked, 33% staked, 100% staked.
	// 0% staked: All rewards to stakers
	// 33% staked: 33% to stakers
	// 100% staked: All rewards to Bonders

	if totalStaked.GTE(totalBonded) { // Zero payments to stakers when staked == bonded
		return sdk.ZeroUint()
	}
	factor := totalBonded.Add(totalStaked).Quo(common.SafeSub(totalBonded, totalStaked)) // (y + x) / (y - x)
	return totalRewards.Quo(factor)
}
