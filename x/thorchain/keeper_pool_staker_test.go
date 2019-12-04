package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperPoolStakerSuite struct{}

var _ = Suite(&KeeperPoolStakerSuite{})

func (s *KeeperPoolStakerSuite) TestPoolStaker(c *C) {
	ctx, k := setupKeeperForTest(c)

	asset := common.BNBAsset
	poolStaker := NewPoolStaker(asset, sdk.NewUint(12))

	k.SetPoolStaker(ctx, poolStaker)
	poolStaker, err := k.GetPoolStaker(ctx, asset)
	c.Assert(err, IsNil)
	c.Check(poolStaker.Asset.Equals(asset), Equals, true)
	c.Check(poolStaker.TotalUnits.Equal(sdk.NewUint(12)), Equals, true)
}
