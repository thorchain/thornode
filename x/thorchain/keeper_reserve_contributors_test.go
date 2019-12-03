package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperReserveContributorsSuite struct{}

var _ = Suite(&KeeperReserveContributorsSuite{})

func (KeeperReserveContributorsSuite) TestReserveContributors(c *C) {
	ctx, k := setupKeeperForTest(c)
	c.Assert(k.AddFeeToReserve(ctx, sdk.NewUint(common.One*100)), IsNil)
	contributor := NewReserveContributor(GetRandomBNBAddress(), sdk.NewUint(common.One*1000))
	contributors := ReserveContributors{
		contributor,
	}
	c.Assert(k.SetReserveContributors(ctx, contributors), IsNil)
	r, err := k.GetReservesContributors(ctx)
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)
}
