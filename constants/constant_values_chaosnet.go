// +build chaosnet

// For internal testing and mockneting
package constants

func init() {
	int64Overrides = map[ConstantName]int64{
		DesireValidatorSet:            12,
		NewPoolCycle:                  17280,           // daily
		RotatePerBlockHeight:          17280,           // daily
		BadValidatorRate:              17280,           // daily
		OldValidatorRate:              17280,           // daily
		MinimumBondInRune:             100000_00000000, // minimum bond 100K RUNE only for chaosnet
		ArtificialRagnarokBlockHeight: 500_000,         // after block height 500,000, start to rotate more nodes out until it reach the minimum BFT
		MaximumStakeRune:              600000_00000000, // on chaosnet , make sure the total staked RUNE per pool is less than 600K
	}
	boolOverrides = map[ConstantName]bool{
		StrictBondStakeRatio: false,
	}
	stringOverrides = map[ConstantName]string{
		DefaultPoolStatus: "Enabled",
	}
}
