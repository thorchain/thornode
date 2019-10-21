package swapservice

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type ValidatorManagerTestSuite struct{}

var _ = Suite(&ValidatorManagerTestSuite{})

func (ps *ValidatorManagerTestSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

func (ValidatorManagerTestSuite) TestSetupValidatorNodes(c *C) {
	ctx, k := setupKeeperForTest(c)
	rotatePerBlockHeight := k.GetAdminConfigRotatePerBlockHeight(ctx, sdk.AccAddress{})
	validatorChangeWindow := k.GetAdminConfigValidatorsChangeWindow(ctx, sdk.AccAddress{})
	vMgr := NewValidatorManager(k)
	vMgr.Meta = &ValidatorMeta{}
	c.Assert(vMgr, NotNil)
	err := vMgr.setupValidatorNodes(ctx, 0)
	c.Assert(err, IsNil)

	// no node accounts at all
	err = vMgr.setupValidatorNodes(ctx, 1)
	c.Assert(err, NotNil)

	activeNode := GetRandomNodeAccount(NodeActive)
	k.SetNodeAccount(ctx, activeNode)

	err = vMgr.setupValidatorNodes(ctx, 1)
	c.Assert(err, IsNil)
	c.Assert(vMgr.Meta.RotateAtBlockHeight, Equals, int64(rotatePerBlockHeight+1))
	c.Assert(vMgr.Meta.RotateWindowOpenAtBlockHeight, Equals, int64(rotatePerBlockHeight+1-validatorChangeWindow))

	readyNode := GetRandomNodeAccount(NodeReady)
	k.SetNodeAccount(ctx, readyNode)

	// one active node and one ready node on start up
	// it should take both of the node as active
	vMgr1 := NewValidatorManager(k)
	vMgr1.BeginBlock(ctx, 1)
	c.Assert(vMgr1.Meta, NotNil)
	c.Assert(vMgr1.Meta.RotateAtBlockHeight, Equals, int64(rotatePerBlockHeight+1))
	c.Assert(vMgr1.Meta.RotateWindowOpenAtBlockHeight, Equals, int64(rotatePerBlockHeight+1-validatorChangeWindow))
	c.Assert(vMgr1.Meta.Queued.IsEmpty(), Equals, true)
	c.Assert(vMgr1.Meta.Nominated.IsEmpty(), Equals, true)
	activeNodes, err := k.ListActiveNodeAccounts(ctx)
	c.Assert(err, IsNil)
	c.Logf("active nodes:%s", activeNodes)
	c.Assert(len(activeNodes) == 2, Equals, true)

	activeNode1 := GetRandomNodeAccount(NodeActive)
	activeNode2 := GetRandomNodeAccount(NodeActive)
	k.SetNodeAccount(ctx, activeNode1)
	k.SetNodeAccount(ctx, activeNode2)

	// three active nodes and 1 ready nodes, it should take them all
	vMgr2 := NewValidatorManager(k)
	vMgr2.BeginBlock(ctx, 1)

	c.Assert(vMgr2.Meta, NotNil)
	c.Assert(vMgr2.Meta.RotateAtBlockHeight, Equals, int64(rotatePerBlockHeight+1))
	c.Assert(vMgr2.Meta.RotateWindowOpenAtBlockHeight, Equals, int64(rotatePerBlockHeight+1-validatorChangeWindow))
	c.Assert(vMgr2.Meta.Queued.IsEmpty(), Equals, true)
	c.Assert(vMgr2.Meta.Nominated.IsEmpty(), Equals, true)

	activeNodes1, err := k.ListActiveNodeAccounts(ctx)
	c.Assert(err, IsNil)
	c.Assert(len(activeNodes1) == 4, Equals, true)
	// No standby nodes
	validatorUpdates := vMgr2.EndBlock(ctx, rotatePerBlockHeight+1-validatorChangeWindow)
	c.Assert(validatorUpdates, IsNil)
	c.Assert(vMgr2.Meta.RotateWindowOpenAtBlockHeight, Equals, int64(rotatePerBlockHeight*2+1-validatorChangeWindow))
	c.Assert(vMgr2.Meta.Nominated.IsEmpty(), Equals, true)
	c.Assert(vMgr2.Meta.Queued.IsEmpty(), Equals, true)

	rotateHeight := rotatePerBlockHeight + 1
	validatorUpdates = vMgr2.EndBlock(ctx, int64(rotateHeight))
	c.Assert(validatorUpdates, IsNil)
	c.Assert(vMgr2.Meta.RotateWindowOpenAtBlockHeight, Equals, int64(rotatePerBlockHeight*2+1-validatorChangeWindow))
	c.Assert(vMgr2.Meta.RotateAtBlockHeight, Equals, int64(rotatePerBlockHeight*2+1))
	c.Assert(vMgr2.Meta.Nominated.IsEmpty(), Equals, true)
	c.Assert(vMgr2.Meta.Queued.IsEmpty(), Equals, true)

	standbyNode := GetRandomNodeAccount(NodeStandby)
	k.SetNodeAccount(ctx, standbyNode)

	openWindow := vMgr2.Meta.RotateWindowOpenAtBlockHeight
	validatorUpdates = vMgr2.EndBlock(ctx, openWindow)
	c.Assert(validatorUpdates, IsNil)
	c.Assert(vMgr2.Meta.Nominated.IsEmpty(), Equals, false)
	c.Assert(vMgr2.Meta.Queued.IsEmpty(), Equals, false)
	c.Assert(vMgr2.Meta.Nominated.Equals(standbyNode), Equals, true)

	allNodes := NodeAccounts{
		activeNode, activeNode1, activeNode2, readyNode,
	}

	sort.Sort(allNodes)

	// TODO fix this test
	// c.Assert(vMgr2.Meta.Queued.Equals(allNodes.First()), Equals, true, Commentf("%s %s", vMgr2.Meta.Queued.NodeAddress, allNodes.First().NodeAddress))

	nominatedNode := vMgr2.Meta.Nominated
	// nominated node is not in ready status abandon the rotation
	rotateAtHeight := vMgr2.Meta.RotateAtBlockHeight
	validatorUpdates = vMgr2.EndBlock(ctx, rotateAtHeight)
	c.Assert(validatorUpdates, IsNil)
	c.Assert(vMgr2.Meta.Nominated.IsEmpty(), Equals, true)
	c.Assert(vMgr2.Meta.Queued.IsEmpty(), Equals, true)
	nominatedNode, err = k.GetNodeAccount(ctx, nominatedNode.NodeAddress)
	c.Assert(err, IsNil)
	c.Assert(nominatedNode.Status, Equals, NodeStandby)

	// rotate validator, all good
	// nominatedNode need to be in ready status
	openWindow = vMgr2.Meta.RotateWindowOpenAtBlockHeight
	validatorUpdates = vMgr2.EndBlock(ctx, openWindow)
	c.Assert(validatorUpdates, IsNil)
	nominatedNode.UpdateStatus(NodeReady)
	k.SetNodeAccount(ctx, nominatedNode)
	queueNode := vMgr2.Meta.Queued

	rotateAtHeight = vMgr2.Meta.RotateAtBlockHeight
	validatorUpdates = vMgr2.EndBlock(ctx, rotateAtHeight)
	c.Assert(validatorUpdates, NotNil)
	c.Assert(vMgr2.Meta.Nominated.IsEmpty(), Equals, true)
	c.Assert(vMgr2.Meta.Queued.IsEmpty(), Equals, true)
	// get the node account from data store again
	nominatedNode, err = k.GetNodeAccount(ctx, nominatedNode.NodeAddress)
	c.Assert(err, IsNil)
	c.Assert(nominatedNode.Status, Equals, NodeActive)
	queueNode, err = k.GetNodeAccount(ctx, queueNode.NodeAddress)
	c.Assert(err, IsNil)
	c.Assert(queueNode.Status, Equals, NodeStandby)
}
