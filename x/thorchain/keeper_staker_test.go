package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperStakerSuite struct{}

var _ = Suite(&KeeperStakerSuite{})

func (s *KeeperStakerSuite) TestStaker(c *C) {
	ctx, k := setupKeeperForTest(c)
	asset := common.BNBAsset

	staker, err := k.GetStaker(ctx, asset, GetRandomBNBAddress())
	c.Assert(err, IsNil)
	c.Check(staker.PendingRune, NotNil)
	c.Check(staker.Units, NotNil)

	staker = Staker{
		Asset:        asset,
		Units:        sdk.NewUint(12),
		RuneAddress:  GetRandomBNBAddress(),
		AssetAddress: GetRandomBTCAddress(),
	}

	k.SetStaker(ctx, staker)
	staker, err = k.GetStaker(ctx, asset, staker.RuneAddress)
	c.Assert(err, IsNil)
	c.Check(staker.Asset.Equals(asset), Equals, true)
	c.Check(staker.Units.Equal(sdk.NewUint(12)), Equals, true)
}
