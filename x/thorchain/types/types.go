package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultCodespace sdk.CodespaceType = ModuleName
	MajorityFactor   uint64            = 3
)

func HasMajority(signers, total int) bool {
	if signers > total {
		return false // will not have majority if THORNode have more signers than node accounts. This shouldn't be possible
	}
	if signers <= 0 {
		return false // edge case
	}
	mU := sdk.NewUint(MajorityFactor)

	// 10*4 / (6.67*2) <= 3
	// 4*4 / (3*2) <= 3
	// 3*4 / (2*2) <= 3
	// Is able to determine "majority" without needing floats or DECs
	tU := sdk.NewUint(uint64(total))
	sU := sdk.NewUint(uint64(signers))
	factor := tU.MulUint64(2).Quo(sU)
	return mU.GTE(factor)
}
