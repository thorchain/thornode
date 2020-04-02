package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type BlockGasTestSuite struct{}

var _ = Suite(&BlockGasTestSuite{})

func (BlockGasTestSuite) TestBlockGas(c *C) {
	b := NewBlockGas(1)
	gas := common.Gas{
		common.NewCoin(common.BNBAsset, sdk.NewUint(100)),
		common.NewCoin(common.BTCAsset, sdk.NewUint(200)),
	}
	b.AddGas(gas, GasTypeSpend)
	c.Assert(b.GasSpend.Equals(gas), Equals, true)
	b.AddGas(gas, GasTypeReimburse)
	c.Assert(b.GasReimburse.Equals(gas), Equals, true)
	b.AddGas(gas, GasTypeTopup)
	c.Assert(b.GasTopup.Equals(gas), Equals, true)
}
