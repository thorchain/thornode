package common

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type GSuite struct{}

var _ = Suite(&GSuite{})

func (s *GSuite) TestMultiGasCalc(c *C) {
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

func (s *GSuite) TestCombineGas(c *C) {
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
