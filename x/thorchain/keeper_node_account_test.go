package thorchain

import (
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type KeeperNodeAccountSuite struct{}

var _ = Suite(&KeeperNodeAccountSuite{})

func (s *KeeperNodeAccountSuite) TestNodeAccount(c *C) {
	ctx, k := setupKeeperForTest(c)
	ctx = ctx.WithBlockHeight(10)

	na1 := GetRandomNodeAccount(NodeActive)
	na2 := GetRandomNodeAccount(NodeStandby)
	k.SetNodeAccount(ctx, na1)
	k.SetNodeAccount(ctx, na2)
	c.Check(na1.ActiveBlockHeight, Equals, int64(10))
	c.Check(na1.SlashPoints, Equals, int64(0))
	c.Check(na2.ActiveBlockHeight, Equals, int64(0))
	c.Check(na2.SlashPoints, Equals, int64(0))

	count, err := k.TotalActiveNodeAccount(ctx)
	c.Assert(err, IsNil)
	c.Check(count, Equals, 1)

	na, err := k.GetNodeAccount(ctx, na1.NodeAddress)
	c.Assert(err, IsNil)
	c.Check(na.Equals(na1), Equals, true)

	na, err = k.GetNodeAccountByPubKey(ctx, na1.NodePubKey.Secp256k1)
	c.Assert(err, IsNil)
	c.Check(na.Equals(na1), Equals, true)

	na, err = k.GetNodeAccountByBondAddress(ctx, na1.BondAddress)
	c.Assert(err, IsNil)
	c.Check(na.Equals(na1), Equals, true)

	valCon := "im unique!"
	pubkeys := GetRandomPubkeys()
	err = k.EnsureTrustAccountUnique(ctx, na1.ValidatorConsPubKey, common.EmptyPubKeys)
	c.Assert(err, NotNil)
	err = k.EnsureTrustAccountUnique(ctx, "", pubkeys)
	c.Assert(err, NotNil)
	err = k.EnsureTrustAccountUnique(ctx, na1.ValidatorConsPubKey, pubkeys)
	c.Assert(err, NotNil)
	err = k.EnsureTrustAccountUnique(ctx, valCon, na1.NodePubKey)
	c.Assert(err, NotNil)
	err = k.EnsureTrustAccountUnique(ctx, valCon, pubkeys)
	c.Assert(err, IsNil)
}
