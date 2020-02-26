package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/constants"
)

type ValidatorMgrV1TestSuite struct{}

var _ = Suite(&ValidatorMgrV1TestSuite{})

func (vts *ValidatorMgrV1TestSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

func (vts *ValidatorMgrV1TestSuite) TestBadActors(c *C) {
	ctx, k := setupKeeperForTest(c)
	ctx = ctx.WithBlockHeight(1000)

	versionedTxOutStoreDummy := NewVersionedTxOutStoreDummy()
	versionedVaultMgrDummy := NewVersionedVaultMgrDummy(versionedTxOutStoreDummy)
	vMgr := newValidatorMgrV1(k, versionedTxOutStoreDummy, versionedVaultMgrDummy)
	c.Assert(vMgr, NotNil)

	// no bad actors with active node accounts
	nas, err := vMgr.findBadActors(ctx)
	c.Assert(err, IsNil)
	c.Assert(nas, HasLen, 0)

	activeNode := GetRandomNodeAccount(NodeActive)
	activeNode.SlashPoints = 0
	c.Assert(k.SetNodeAccount(ctx, activeNode), IsNil)

	// no bad actors with active node accounts with no slash points
	nas, err = vMgr.findBadActors(ctx)
	c.Assert(err, IsNil)
	c.Assert(nas, HasLen, 0)

	activeNode = GetRandomNodeAccount(NodeActive)
	activeNode.SlashPoints = 25
	c.Assert(k.SetNodeAccount(ctx, activeNode), IsNil)
	activeNode = GetRandomNodeAccount(NodeActive)
	activeNode.SlashPoints = 50
	c.Assert(k.SetNodeAccount(ctx, activeNode), IsNil)

	// finds the worse actor
	nas, err = vMgr.findBadActors(ctx)
	c.Assert(err, IsNil)
	c.Assert(nas, HasLen, 1)
	c.Check(nas[0].NodeAddress.Equals(activeNode.NodeAddress), Equals, true, Commentf("%+v\n", nas[0].SlashPoints))

	// create really bad actors (crossing the redline)
	bad1 := GetRandomNodeAccount(NodeActive)
	bad1.SlashPoints = 1000
	c.Assert(k.SetNodeAccount(ctx, bad1), IsNil)
	bad2 := GetRandomNodeAccount(NodeActive)
	bad2.SlashPoints = 10000
	c.Assert(k.SetNodeAccount(ctx, bad2), IsNil)

	nas, err = vMgr.findBadActors(ctx)
	c.Assert(err, IsNil)
	c.Assert(nas, HasLen, 2, Commentf("%d", len(nas)))
	c.Check(nas[0].NodeAddress.Equals(bad2.NodeAddress), Equals, true, Commentf("%+v\n", nas[0].SlashPoints))
	c.Check(nas[1].NodeAddress.Equals(bad1.NodeAddress), Equals, true, Commentf("%+v\n", nas[1].SlashPoints))
}

func (vts *ValidatorMgrV1TestSuite) TestRagnarokBond(c *C) {
	ctx, k := setupKeeperForTest(c)
	ctx = ctx.WithBlockHeight(1)
	ver := semver.MustParse("0.1.0")
	versionedTxOutStoreDummy := NewVersionedTxOutStoreDummy()
	txOutStore, err := versionedTxOutStoreDummy.GetTxOutStore(k, ver)
	c.Assert(err, IsNil)

	versionedVaultMgrDummy := NewVersionedVaultMgrDummy(versionedTxOutStoreDummy)
	vMgr := newValidatorMgrV1(k, versionedTxOutStoreDummy, versionedVaultMgrDummy)
	c.Assert(vMgr, NotNil)
	constAccessor := constants.GetConstantValues(ver)
	err = vMgr.setupValidatorNodes(ctx, 0, constAccessor)
	c.Assert(err, IsNil)

	activeNode := GetRandomNodeAccount(NodeActive)
	activeNode.Bond = sdk.NewUint(100)
	c.Assert(k.SetNodeAccount(ctx, activeNode), IsNil)

	disabledNode := GetRandomNodeAccount(NodeDisabled)
	disabledNode.Bond = sdk.ZeroUint()
	c.Assert(k.SetNodeAccount(ctx, disabledNode), IsNil)

	c.Assert(vMgr.ragnarokBond(ctx, 1), IsNil)
	activeNode, err = k.GetNodeAccount(ctx, activeNode.NodeAddress)
	c.Assert(err, IsNil)
	c.Check(activeNode.Bond.Equal(sdk.NewUint(90)), Equals, true)
	items, err := txOutStore.GetOutboundItems(ctx)
	c.Assert(err, IsNil)
	c.Check(items, HasLen, 1, Commentf("Len %d", items))
	txOutStore.ClearOutboundItems(ctx)

	c.Assert(vMgr.ragnarokBond(ctx, 2), IsNil)
	activeNode, err = k.GetNodeAccount(ctx, activeNode.NodeAddress)
	c.Assert(err, IsNil)
	c.Check(activeNode.Bond.Equal(sdk.NewUint(72)), Equals, true)
	items, err = txOutStore.GetOutboundItems(ctx)
	c.Assert(err, IsNil)
	c.Check(items, HasLen, 1, Commentf("Len %d", items))
}

func (vtx *ValidatorMgrV1TestSuite) TestFindCounToRemove(c *C) {
	// remove one
	c.Check(findCounToRemove(0, 0, NodeAccounts{
		NodeAccount{LeaveHeight: 12},
		NodeAccount{},
		NodeAccount{},
		NodeAccount{},
		NodeAccount{},
	}), Equals, 1)

	// don't remove one
	c.Check(findCounToRemove(0, 0, NodeAccounts{
		NodeAccount{LeaveHeight: 12},
		NodeAccount{LeaveHeight: 12},
		NodeAccount{},
		NodeAccount{},
	}), Equals, 0)

	// remove one because of request to leave
	c.Check(findCounToRemove(0, 0, NodeAccounts{
		NodeAccount{LeaveHeight: 12, RequestedToLeave: true},
		NodeAccount{},
		NodeAccount{},
		NodeAccount{},
	}), Equals, 1)

	// don't remove more than 1/3rd of node accounts
	c.Check(findCounToRemove(0, 0, NodeAccounts{
		NodeAccount{LeaveHeight: 12},
		NodeAccount{LeaveHeight: 12},
		NodeAccount{LeaveHeight: 12},
		NodeAccount{LeaveHeight: 12},
		NodeAccount{LeaveHeight: 12},
		NodeAccount{LeaveHeight: 12},
		NodeAccount{LeaveHeight: 12},
		NodeAccount{LeaveHeight: 12},
		NodeAccount{LeaveHeight: 12},
		NodeAccount{LeaveHeight: 12},
		NodeAccount{LeaveHeight: 12},
		NodeAccount{LeaveHeight: 12},
	}), Equals, 3)
}

func (vts *ValidatorMgrV1TestSuite) TestFindMaxAbleToLeave(c *C) {
	c.Check(findMaxAbleToLeave(-1), Equals, 0)
	c.Check(findMaxAbleToLeave(0), Equals, 0)
	c.Check(findMaxAbleToLeave(1), Equals, 0)
	c.Check(findMaxAbleToLeave(2), Equals, 0)
	c.Check(findMaxAbleToLeave(3), Equals, 0)
	c.Check(findMaxAbleToLeave(4), Equals, 0)

	c.Check(findMaxAbleToLeave(5), Equals, 1)
	c.Check(findMaxAbleToLeave(6), Equals, 1)
	c.Check(findMaxAbleToLeave(7), Equals, 2)
	c.Check(findMaxAbleToLeave(8), Equals, 2)
	c.Check(findMaxAbleToLeave(9), Equals, 2)
	c.Check(findMaxAbleToLeave(10), Equals, 3)
	c.Check(findMaxAbleToLeave(11), Equals, 3)
	c.Check(findMaxAbleToLeave(12), Equals, 3)
}