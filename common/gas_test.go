package common

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type GasSuite struct{}

var _ = Suite(&GasSuite{})

func (s *GasSuite) TestETHGasFee(c *C) {
	gas := GetETHGasFee(big.NewInt(20), 4)
	amt := gas[0].Amount
	c.Check(
		amt.Equal(sdk.NewUint(425440)),
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

func (s *GasSuite) TestCalcGasPrice(c *C) {
	gasInfo := []sdk.Uint{sdk.NewUint(37500), sdk.NewUint(30000)}
	tx := Tx{
		Coins: Coins{
			NewCoin(BNBAsset, sdk.NewUint(80808080)),
		},
	}

	gas := CalcGasPrice(tx, BNBAsset, gasInfo)
	c.Check(gas.Equals(Gas{NewCoin(BNBAsset, sdk.NewUint(37500))}), Equals, true)

	tx = Tx{
		Coins: Coins{
			NewCoin(BNBAsset, sdk.NewUint(80808080)),
			NewCoin(BNBAsset, sdk.NewUint(80808080)),
		},
	}

	gas = CalcGasPrice(tx, BNBAsset, gasInfo)
	c.Check(gas.Equals(Gas{NewCoin(BNBAsset, sdk.NewUint(60000))}), Equals, true)
}

func (s *GasSuite) TestUpdateGasPrice(c *C) {
	gasInfo := UpdateGasPrice(Tx{}, BNBAsset, []sdk.Uint{sdk.NewUint(33)})
	c.Assert(gasInfo, HasLen, 1)
	c.Check(gasInfo[0].Equal(sdk.NewUint(33)), Equals, true)

	tx := Tx{
		Coins: Coins{
			NewCoin(BNBAsset, sdk.NewUint(80808080)),
		},
		Gas: Gas{
			NewCoin(BNBAsset, sdk.NewUint(222)),
		},
	}

	gasInfo = UpdateGasPrice(tx, BNBAsset, nil)
	c.Assert(gasInfo, HasLen, 2)
	c.Check(gasInfo[0].Equal(sdk.NewUint(222)), Equals, true)

	tx = Tx{
		Coins: Coins{
			NewCoin(BNBAsset, sdk.NewUint(80808080)),
			NewCoin(BTCAsset, sdk.NewUint(80808080)),
		},
		Gas: Gas{
			NewCoin(BNBAsset, sdk.NewUint(222)),
		},
	}

	gasInfo = UpdateGasPrice(tx, BNBAsset, gasInfo)
	c.Assert(gasInfo, HasLen, 2)
	c.Check(gasInfo[1].Equal(sdk.NewUint(111)), Equals, true)
}
