package common

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// One is useful type so THORNode doesn't need to manage 8 zeroes all the time
const One = 100000000

// GetShare this method will panic if any of the input parameter can't be convert to sdk.Dec
// which shouldn't happen
func GetShare(part sdk.Uint, total sdk.Uint, allocation sdk.Uint) sdk.Uint {
	if part.IsZero() || total.IsZero() {
		return sdk.ZeroUint()
	}

	// use string to convert sdk.Uint to sdk.Dec is the only way I can find out without being constrain to uint64
	// sdk.Uint can hold values way larger than uint64 , because it is using big.Int internally
	aD, err := sdk.NewDecFromStr(allocation.String())
	if err != nil {
		panic(fmt.Errorf("fail to convert %s to sdk.Dec: %w", allocation.String(), err))
	}

	pD, err := sdk.NewDecFromStr(part.String())
	if err != nil {
		panic(fmt.Errorf("fatil to convert %s to sdk.Dec: %w", part.String(), err))
	}
	tD, err := sdk.NewDecFromStr(total.String())
	if err != nil {
		panic(fmt.Errorf("fail to convert%s to sdk.Dec: %w", total.String(), err))
	}
	// A / (Total / part) == A * (part/Total) but safer when part < Totals
	result := aD.Quo(tD.Quo(pD))
	return sdk.NewUintFromBigInt(result.RoundInt().BigInt())
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
