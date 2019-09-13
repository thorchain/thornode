package common

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	. "gopkg.in/check.v1"
)

type TypeConvertTestSuite struct{}

var _ = Suite(&TypeConvertTestSuite{})

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
