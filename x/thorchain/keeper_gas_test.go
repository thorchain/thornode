package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type KeeperGasSuite struct{}

var _ = Suite(&KeeperGasSuite{})

func (s *KeeperGasSuite) TestGas(c *C) {
	ctx, k := setupKeeperForTest(c)

	bnbGas := []sdk.Uint{
		sdk.NewUint(37500),
		sdk.NewUint(30000),
	}

	k.SetGas(ctx, common.BNBAsset, bnbGas)

	gas, err := k.GetGas(ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Assert(gas, HasLen, 2)
	c.Assert(gas[0].Equal(sdk.NewUint(37500)), Equals, true)
	c.Assert(gas[1].Equal(sdk.NewUint(30000)), Equals, true)
}
