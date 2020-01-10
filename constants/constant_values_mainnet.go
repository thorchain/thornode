// +build mainnet

// For imainnet
package constants

func init() {
	boolOverrides = map[ConstantName]bool{
		StrictBondStakeRatio: true,
	}
}
