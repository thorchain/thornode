// +build testnet

// For Public TestNet
package constants

func init() {
	int64Overrides = map[ConstantName]int64{
		DesireValidatorSet:   12,
		RotatePerBlockHeight: 17280,
		BadValidatorRate:     17280,
		OldValidatorRate:     17280,
		MinimumBondInRune:    100_000_000, // 1 rune
		StakeLockUpBlocks:    0,
	}
	boolOverrides = map[ConstantName]bool{
		StrictBondStakeRatio: false,
	}
	stringOverrides = map[ConstantName]string{
		DefaultPoolStatus: "Enabled",
	}
}
