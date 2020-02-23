package thorchain

import (
	"github.com/blang/semver"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperNodeAccountSuite struct{}

var _ = Suite(&KeeperNodeAccountSuite{})

func (s *KeeperNodeAccountSuite) TestNodeAccount(c *C) {
	ctx, k := setupKeeperForTest(c)
	ctx = ctx.WithBlockHeight(10)

	na1 := GetRandomNodeAccount(NodeActive)
	na2 := GetRandomNodeAccount(NodeStandby)
	c.Assert(k.SetNodeAccount(ctx, na1), IsNil)
	c.Assert(k.SetNodeAccount(ctx, na2), IsNil)
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

	na, err = k.GetNodeAccountByPubKey(ctx, na1.PubKeySet.Secp256k1)
	c.Assert(err, IsNil)
	c.Check(na.Equals(na1), Equals, true)

	na, err = k.GetNodeAccountByBondAddress(ctx, na1.BondAddress)
	c.Assert(err, IsNil)
	c.Check(na.Equals(na1), Equals, true)

	valCon := "im unique!"
	pubkeys := GetRandomPubKeySet()
	err = k.EnsureNodeKeysUnique(ctx, na1.ValidatorConsPubKey, common.EmptyPubKeySet)
	c.Assert(err, NotNil)
	err = k.EnsureNodeKeysUnique(ctx, "", pubkeys)
	c.Assert(err, NotNil)
	err = k.EnsureNodeKeysUnique(ctx, na1.ValidatorConsPubKey, pubkeys)
	c.Assert(err, NotNil)
	err = k.EnsureNodeKeysUnique(ctx, valCon, na1.PubKeySet)
	c.Assert(err, NotNil)
	err = k.EnsureNodeKeysUnique(ctx, valCon, pubkeys)
	c.Assert(err, IsNil)
	addr := GetRandomBech32Addr()
	na, err = k.GetNodeAccount(ctx, addr)
	c.Assert(err, IsNil)
	c.Assert(na.Status, Equals, NodeUnknown)
	c.Assert(na.ValidatorConsPubKey, Equals, "")
}

func (s *KeeperNodeAccountSuite) TestGetMinJoinVersion(c *C) {
	type nodeInfo struct {
		status  NodeStatus
		version semver.Version
	}
	inputs := []struct {
		nodeInfoes      []nodeInfo
		expectedVersion semver.Version
	}{
		{
			nodeInfoes: []nodeInfo{
				{
					status:  NodeActive,
					version: semver.MustParse("0.2.0"),
				},
				{
					status:  NodeActive,
					version: semver.MustParse("0.3.0"),
				},
				{
					status:  NodeActive,
					version: semver.MustParse("0.3.0"),
				},
				{
					status:  NodeStandby,
					version: semver.MustParse("0.2.0"),
				},
				{
					status:  NodeStandby,
					version: semver.MustParse("0.2.0"),
				},
			},
			expectedVersion: semver.MustParse("0.3.0"),
		},
		{
			nodeInfoes: []nodeInfo{
				{
					status:  NodeActive,
					version: semver.MustParse("0.2.0"),
				},
				{
					status:  NodeActive,
					version: semver.MustParse("1.3.0"),
				},
				{
					status:  NodeActive,
					version: semver.MustParse("0.3.0"),
				},
				{
					status:  NodeStandby,
					version: semver.MustParse("0.2.0"),
				},
				{
					status:  NodeStandby,
					version: semver.MustParse("0.2.0"),
				},
			},
			expectedVersion: semver.MustParse("0.3.0"),
		},
		{
			nodeInfoes: []nodeInfo{
				{
					status:  NodeActive,
					version: semver.MustParse("0.2.0"),
				},
				{
					status:  NodeActive,
					version: semver.MustParse("1.3.0"),
				},
				{
					status:  NodeActive,
					version: semver.MustParse("0.3.0"),
				},
				{
					status:  NodeActive,
					version: semver.MustParse("0.2.0"),
				},
				{
					status:  NodeActive,
					version: semver.MustParse("0.2.0"),
				},
			},
			expectedVersion: semver.MustParse("0.2.0"),
		},
	}

	for _, item := range inputs {
		ctx, k := setupKeeperForTest(c)
		for _, ni := range item.nodeInfoes {
			na1 := GetRandomNodeAccount(ni.status)
			na1.Version = ni.version
			c.Assert(k.SetNodeAccount(ctx, na1), IsNil)
		}
		c.Check(k.GetMinJoinVersion(ctx).Equals(item.expectedVersion), Equals, true, Commentf("%+v", k.GetMinJoinVersion(ctx)))
	}
}
