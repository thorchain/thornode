package common

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type GasSuite struct{}

var _ = Suite(&GasSuite{})

func (s *GasSuite) TestMultiGasCalc(c *C) {
	gas := GetBNBGasFeeMulti(1)
	amt := gas[0].Amount
	c.Check(
		amt.Equal(sdk.NewUint(30000)),
		Equals,
		true,
		Commentf("%d", amt.Uint64()),
	)

	gas = GetBNBGasFeeMulti(3)
	amt = gas[0].Amount
	c.Check(
		amt.Equal(sdk.NewUint(90000)),
		Equals,
		true,
		Commentf("%d", amt.Uint64()),
	)
}

func (s *GasSuite) TestIsEmpty(c *C) {
	gas1 := Gas{
		{Asset: BNBAsset, Amount: sdk.NewUint(11 * One)},
	}
	c.Check(gas1.IsEmpty(), Equals, false)
	c.Check(Gas{}.IsEmpty(), Equals, true)
}

func (s *GasSuite) TestCombineGas(c *C) {
	gas1 := Gas{
		{Asset: BNBAsset, Amount: sdk.NewUint(11 * One)},
	}
	gas2 := Gas{
		{Asset: BNBAsset, Amount: sdk.NewUint(14 * One)},
		{Asset: BTCAsset, Amount: sdk.NewUint(20 * One)},
	}

	gas := gas1.Add(gas2)
	c.Assert(gas, HasLen, 2)
	c.Check(gas[0].Asset.Equals(BNBAsset), Equals, true)
	c.Check(gas[0].Amount.Equal(sdk.NewUint(25*One)), Equals, true, Commentf("%d", gas[0].Amount.Uint64()))
	c.Check(gas[1].Asset.Equals(BTCAsset), Equals, true)
	c.Check(gas[1].Amount.Equal(sdk.NewUint(20*One)), Equals, true)
}

func (s *GasSuite) TestUpdateBNBGasFee(c *C) {
	noGas := Gas{}
	single := BNBGasFeeSingleton
	multiple := GetBNBGasFeeMulti(2)

	UpdateBNBGasFee(noGas, 1)
	// nothing change
	c.Assert(BNBGasFeeSingleton.Equals(single), Equals, true)
	// invalid gas
	UpdateBNBGasFee(Gas{
		{Asset: EmptyAsset, Amount: sdk.ZeroUint()},
	}, 1)
	c.Assert(BNBGasFeeSingleton.Equals(single), Equals, true)

	UpdateBNBGasFee(Gas{
		{Asset: BTCAsset, Amount: sdk.NewUint(1000)},
	}, 1)
	c.Assert(BNBGasFeeSingleton.Equals(single), Equals, true)
	newGas := Gas{
		{
			Asset: BNBAsset, Amount: sdk.NewUint(1000),
		},
	}
	UpdateBNBGasFee(newGas, 1)
	c.Assert(BNBGasFeeSingleton.Equals(newGas), Equals, true)
	UpdateBNBGasFee(single, 1)
	UpdateBNBGasFee(newGas, 2)
	c.Assert(BNBGasFeeMulti.Equals(Gas{
		{
			Asset: BNBAsset, Amount: sdk.NewUint(500),
		},
	}), Equals, true)
	UpdateBNBGasFee(multiple, 2)
}
