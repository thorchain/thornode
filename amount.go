package common

import (
	"strconv"
	"strings"
)

const floatPrecision = 8

type Amount string

var ZeroAmount Amount = Amount("0")

func NewAmount(amount string) (Amount, error) {
	_, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return ZeroAmount, err
	}
	return Amount(amount), nil
}

func NewAmountFromFloat(f float64) Amount {
	return Amount(strconv.FormatFloat(f, 'f', floatPrecision, 64))
}

func (a Amount) Equals(a2 Amount) bool {
	return a.Float64() == a2.Float64()
}

func (a Amount) Plus(a2 Amount) Amount {
	return NewAmountFromFloat(a.Float64() + a2.Float64())
}

func (a Amount) Minus(a2 Amount) Amount {
	return NewAmountFromFloat(a.Float64() - a2.Float64())
}

func (a Amount) IsEmpty() bool {
	return strings.TrimSpace(a.String()) == ""
}

func (a Amount) GreaterThen(f float64) bool {
	return a.Float64() > f
}

func (a Amount) LessThen(f float64) bool {
	return a.Float64() < f
}

func (a Amount) IsZero() bool {
	return a.Equals(ZeroAmount) || a.Float64() == ZeroAmount.Float64()
}

func (a Amount) Float64() float64 {
	amt, _ := strconv.ParseFloat(a.String(), 64)
	return amt
}
func (a Amount) IsNegative() bool {
	return a.Float64() < 0
}

func (a Amount) String() string {
	return string(a)
}
