package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperVaultSuite struct{}

var _ = Suite(&KeeperVaultSuite{})

func (KeeperVaultSuite) TestVaultData(c *C) {
	ctx, k := setupKeeperForTest(c)
	vd := NewVaultData()
	err := k.SetVaultData(ctx, vd)
	c.Assert(err, IsNil)
	c.Assert(k.UpdateVaultData(ctx), IsNil)

	// add something in vault
	vd.TotalReserve = sdk.NewUint(common.One * 100)
	vd.Gas = common.GetBNBGasFeeMulti(1)
	err = k.SetVaultData(ctx, vd)
	c.Assert(err, IsNil)
	c.Assert(k.UpdateVaultData(ctx), IsNil)
}
