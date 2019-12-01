package common

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// One is useful type so we don't need to type 8 zeros all the time
const One = 100000000

func GetShare(part, total, allocation sdk.Uint) sdk.Uint {
	if part.IsZero() || total.IsZero() {
		return sdk.ZeroUint()
	}
	// A / (Total / part) == A * (part/Total) but safer when part < Total
	return sdk.NewUint(uint64(math.Round((float64(allocation.Uint64()) / (float64(total.Uint64()) / float64(part.Uint64()))))))
}

func SafeSub(input1, input2 sdk.Uint) sdk.Uint {
	if input2.GT(input1) {
		return sdk.ZeroUint()
	}
	return input1.Sub(input2)
}

// UintToFloat64
func UintToFloat64(input sdk.Uint) float64 {
	return float64(input.Uint64())
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

func FloatToUintAndMultipleOne(input float64) sdk.Uint {
	return sdk.NewUint(uint64(math.Round(input * One)))
}
func FloatToUint(input float64) sdk.Uint {
	return sdk.NewUint(uint64(math.Round(input)))
}

func AmountToUint(amount Amount) sdk.Uint {
	return FloatToUint(amount.Float64())
}

func UintToAmount(input sdk.Uint) Amount {
	return NewAmountFromFloat(float64(input.Uint64()))
}

func FloatToDec(input float64) sdk.Dec {
	i := int64(input * One)
	return sdk.NewDecWithPrec(i, 8)
}
