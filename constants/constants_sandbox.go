// +build sandbox

// For internal testing and sandboxing
package constants

func init() {
	int64Overrides = map[ConstantName]int64{
		DesireValidatorSet:   12,
		RotatePerBlockHeight: 60,          // hourly
		BadValidatorRate:     60,          // hourly
		OldValidatorRate:     60,          // hourly
		MinimumBondInRune:    100_000_000, // 1 rune
	}
}
