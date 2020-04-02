package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type GasManagerTestSuite struct{}

var _ = Suite(&GasManagerTestSuite{})

func (GasManagerTestSuite) TestEmitGasEvent(c *C) {
	ctx, k := setupKeeperForTest(c)
	blockGas, err := k.GetBlockGas(ctx)
	c.Assert(err, IsNil)
	c.Assert(blockGas.IsEmpty(), Equals, true)
	gas := common.Gas{
		common.NewCoin(common.BNBAsset, sdk.NewUint(37500)),
	}
	blockGas.AddGas(gas, GasTypeSpend)
	err = k.SaveBlockGas(ctx, blockGas)
	c.Assert(err, IsNil)
	err = EmitGasEvents(ctx, k)
	c.Assert(err, IsNil)
	eventID, err := k.GetCurrentEventID(ctx)
	c.Assert(err, IsNil)
	e, err := k.GetEvent(ctx, eventID-1)
	c.Assert(err, IsNil)
	c.Assert(e.Type, Equals, txGas.String())
}
