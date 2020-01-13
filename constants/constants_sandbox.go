// +build sandbox

// For internal testing and sandboxing
package constants

func init() {
	int64Overrides = map[ConstantName]int64{
		DesireValidatorSet:   12,
		RotatePerBlockHeight: 720,         // hourly
		BadValidatorRate:     720,         // hourly
		OldValidatorRate:     720,         // hourly
		MinimumBondInRune:    100_000_000, // 1 rune
	}
}
