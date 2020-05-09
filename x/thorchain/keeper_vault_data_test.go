package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type KeeperVaultDataSuite struct{}

var _ = Suite(&KeeperVaultDataSuite{})

func (KeeperVaultDataSuite) TestVaultData(c *C) {
	ctx, k := setupKeeperForTest(c)
	ver := constants.SWVersion
	constAccessor := constants.GetConstantValues(ver)
	vd := NewVaultData()
	err := k.SetVaultData(ctx, vd)
	c.Assert(err, IsNil)
	gasManager := NewDummyGasManager()

	c.Assert(k.UpdateVaultData(ctx, constAccessor, gasManager, NewEventMgr()), IsNil)

	// add something in vault
	vd.TotalReserve = sdk.NewUint(common.One * 100)
	err = k.SetVaultData(ctx, vd)
	c.Assert(err, IsNil)
	c.Assert(k.UpdateVaultData(ctx, constAccessor, gasManager, NewEventMgr()), IsNil)
}

func (KeeperVaultDataSuite) TestGetTotalActiveNodeWithBound(c *C) {
	ctx, k := setupKeeperForTest(c)

	node1 := GetRandomNodeAccount(NodeActive)
	c.Assert(k.SetNodeAccount(ctx, node1), IsNil)
	node2 := GetRandomNodeAccount(NodeActive)
	node2.Bond = sdk.ZeroUint()
	c.Assert(k.SetNodeAccount(ctx, node2), IsNil)
	n, err := getTotalActiveNodeWithBond(ctx, k)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, int64(1))
}
