package common

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// One is useful type so THORNode doesn't need to manage 8 zeroes all the time
const One = 100000000

func GetShare(part sdk.Uint, total sdk.Uint, allocation sdk.Uint) sdk.Uint {
	if part.IsZero() || total.IsZero() {
		return sdk.ZeroUint()
	}
	// A / (Total / part) == A * (part/Total) but safer when part < Totals
	pD := sdk.NewDec(int64(part.Uint64()))
	tD := sdk.NewDec(int64(total.Uint64()))
	aD := sdk.NewDec(int64(allocation.Uint64()))
	return sdk.NewUint(uint64((aD.Quo(tD.Quo(pD))).RoundInt64()))
}

func SafeSub(input1, input2 sdk.Uint) sdk.Uint {
	if input2.GT(input1) {
		return sdk.ZeroUint()
	}
	return input1.Sub(input2)
}

// UintToUint64
func UintToUint64(input sdk.Uint) uint64 {
	return input.Uint64()
}

//
func IntToInt64(input sdk.Int) int64 {
	return input.Int64()
}

func IntToUint64(input sdk.Int) uint64 {
	return uint64(input.Int64())
}
