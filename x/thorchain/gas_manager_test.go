package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type GasManagerTestSuite struct{}

var _ = Suite(&GasManagerTestSuite{})

func (GasManagerTestSuite) TestGasManager(c *C) {
	ctx, k := setupKeeperForTest(c)
	gasMgr := NewGasMgr()
	gasEvent := gasMgr.gasEvent
	c.Assert(gasMgr, NotNil)
	gasMgr.BeginBlock()
	c.Assert(gasEvent != gasMgr.gasEvent, Equals, true)

	gasMgr.AddGasAsset(common.Gas{
		common.NewCoin(common.BNBAsset, sdk.NewUint(37500)),
		common.NewCoin(common.BTCAsset, sdk.NewUint(1000)),
	})
	c.Assert(gasMgr.gasEvent.Pools, HasLen, 2)
	gasMgr.AddGasAsset(common.Gas{
		common.NewCoin(common.BNBAsset, sdk.NewUint(38500)),
		common.NewCoin(common.BTCAsset, sdk.NewUint(2000)),
	})
	c.Assert(gasMgr.gasEvent.Pools, HasLen, 2)
	gasMgr.AddGasAsset(common.Gas{
		common.NewCoin(common.RuneAsset(), sdk.NewUint(38500)),
	})
	c.Assert(gasMgr.gasEvent.Pools, HasLen, 3)
	gasMgr.AddRune(common.BTCAsset, sdk.NewUint(3000))
	c.Assert(gasMgr.gasEvent.Pools, HasLen, 3)
	gasMgr.EndBlock(ctx, k)
	eventID, err := k.GetCurrentEventID(ctx)
	c.Assert(err, IsNil)
	event, err := k.GetEvent(ctx, eventID-1)
	c.Assert(err, IsNil)
	c.Assert(event.Type, Equals, gasMgr.gasEvent.Type())
}
