package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type SlashingSuite struct{}

var _ = Suite(&SlashingSuite{})

func (s *SlashingSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

func (s *SlashingSuite) TestObservingSlashing(c *C) {
	var err error
	ctx, k := setupKeeperForTest(c)

	// add one
	na1 := GetRandomNodeAccount(NodeActive)
	k.SetNodeAccount(ctx, na1)

	// add two
	na2 := GetRandomNodeAccount(NodeActive)
	k.SetNodeAccount(ctx, na2)

	k.AddObservingAddresses(ctx, []sdk.AccAddress{na1.NodeAddress})

	// should slash na2 only
	slashForObservingAddresses(ctx, k)

	na1, err = k.GetNodeAccount(ctx, na1.NodeAddress)
	c.Assert(err, IsNil)
	na2, err = k.GetNodeAccount(ctx, na2.NodeAddress)
	c.Assert(err, IsNil)

	c.Assert(na1.SlashPoints, Equals, int64(0))
	c.Assert(na2.SlashPoints, Equals, int64(observingPenalty))

	// since we have cleared all node addresses in slashForObservingAddresses,
	// running it a second time should result in slashing nobody.
	slashForObservingAddresses(ctx, k)
	c.Assert(na1.SlashPoints, Equals, int64(0))
	c.Assert(na2.SlashPoints, Equals, int64(observingPenalty))
}
