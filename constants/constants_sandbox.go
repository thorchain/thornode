// +build sandbox

// For internal testing and sandboxing
package constants

func init() {
	int64Overrides = map[ConstantName]int64{
		DesireValidatorSet:   12,
		RotatePerBlockHeight: 60,          // 5 min
		BadValidatorRate:     60,          // 5 min
		OldValidatorRate:     60,          // 5 min
		MinimumBondInRune:    100_000_000, // 1 rune
	}
}
