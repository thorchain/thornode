package common

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	. "gopkg.in/check.v1"
)

type TypeConvertTestSuite struct{}

var _ = Suite(&TypeConvertTestSuite{})

func (TypeConvertTestSuite) TestSafeSub(c *C) {
	input1 := sdk.NewUint(1)
	input2 := sdk.NewUint(2)

	result1 := SafeSub(input2, input2)
	result2 := SafeSub(input1, input2)
	result3 := SafeSub(input2, input1)

	c.Check(result1.Equal(sdk.ZeroUint()), Equals, true, Commentf("%d", result1.Uint64()))
	c.Check(result2.Equal(sdk.ZeroUint()), Equals, true, Commentf("%d", result2.Uint64()))
	c.Check(result3.Equal(sdk.NewUint(1)), Equals, true, Commentf("%d", result3.Uint64()))
	c.Check(result3.Equal(input2.Sub(input1)), Equals, true, Commentf("%d", result3.Uint64()))
}

func (TypeConvertTestSuite) TestUintToFloat64(c *C) {
	input := sdk.NewUint(1)
	c.Check(UintToFloat64(input), Equals, 1.0)
	input100m := sdk.NewUint(One)
	c.Check(UintToFloat64(input100m), Equals, float64(One))
}

func (TypeConvertTestSuite) TestFloatToUint(c *C) {
	f := 1.99999
	c.Check(FloatToUint(f).Uint64(), Equals, uint64(2))
	f1 := 2.0000000001
	c.Check(FloatToUint(f1).Uint64(), Equals, uint64(2))
	c.Check(FloatToUint(1666666.7899999999).Uint64(), Equals, uint64(1666667))
}

func (TypeConvertTestSuite) TestFloatToUintAndMultipleOne(c *C) {
	c.Check(FloatToUintAndMultipleOne(1.0234560001).Uint64(), Equals, uint64(102345600))
	c.Check(FloatToUintAndMultipleOne(1234.567898765).Uint64(), Equals, uint64(123456789877))
}

func (TypeConvertTestSuite) TestAmountToUint(c *C) {
	amt := NewAmountFromFloat(1.23)
	c.Check(AmountToUint(amt).Uint64(), Equals, uint64(1))
	amt = NewAmountFromFloat(2.99999)
	c.Check(AmountToUint(amt).Uint64(), Equals, uint64(3))
}

func (TypeConvertTestSuite) TestUintToAmount(c *C) {
	u := sdk.NewUint(1)
	c.Check(UintToAmount(u).Float64(), Equals, 1.0)
	u = sdk.NewUint(123456789)
	c.Check(UintToAmount(u).Float64(), Equals, 123456789.0)
}

func (TypeConvertTestSuite) TestFloatToDec(c *C) {
	dec := FloatToDec(34.275)
	c.Check(dec.String(), Equals, "34.275000000000000000")
}
