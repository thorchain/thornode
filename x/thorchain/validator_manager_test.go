package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type ValidatorManagerTestSuite struct{}

var _ = Suite(&ValidatorManagerTestSuite{})

func (vts *ValidatorManagerTestSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

func (vts *ValidatorManagerTestSuite) setDesireValidatorSet(c *C, ctx sdk.Context, k Keeper) {
	activeAccounts, err := k.ListActiveNodeAccounts(ctx)
	c.Assert(err, IsNil)
	for _, item := range activeAccounts {
		k.SetAdminConfig(ctx, NewAdminConfig(DesireValidatorSetKey, "4", item.NodeAddress))
	}
	currentDesireValidatorSet := k.GetAdminConfigDesireValidatorSet(ctx, nil)
	c.Logf("current desire validator set: %d", currentDesireValidatorSet)
}
func (vts *ValidatorManagerTestSuite) TestSetupValidatorNodes(c *C) {
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
	vMgr.rotationPolicy = GetValidatorRotationPolicy(ctx, vMgr.k)

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
	ctx = ctx.WithBlockHeight(rotatePerBlockHeight + 1 - validatorChangeWindow)
	validatorUpdates := vMgr2.EndBlock(ctx)
	c.Assert(validatorUpdates, IsNil)
	c.Assert(vMgr2.Meta.Nominated.IsEmpty(), Equals, true)
	c.Assert(vMgr2.Meta.Queued.IsEmpty(), Equals, true)

	rotateHeight := rotatePerBlockHeight + 1
	ctx = ctx.WithBlockHeight(rotateHeight)
	validatorUpdates = vMgr2.EndBlock(ctx)
	c.Assert(validatorUpdates, IsNil)
	c.Assert(vMgr2.Meta.RotateWindowOpenAtBlockHeight, Equals, int64(rotatePerBlockHeight*2+1-validatorChangeWindow))
	c.Assert(vMgr2.Meta.RotateAtBlockHeight, Equals, int64(rotatePerBlockHeight*2+1))
	c.Assert(vMgr2.Meta.Nominated.IsEmpty(), Equals, true)
	c.Assert(vMgr2.Meta.Queued.IsEmpty(), Equals, true)

	standbyNode := GetRandomNodeAccount(NodeStandby)
	k.SetNodeAccount(ctx, standbyNode)

	//vts.setDesireValidatorSet(c, ctx, k)
	vMgr2.rotationPolicy = GetValidatorRotationPolicy(ctx, k)
	openWindow := vMgr2.Meta.RotateWindowOpenAtBlockHeight
	ctx = ctx.WithBlockHeight(openWindow)
	validatorUpdates = vMgr2.EndBlock(ctx)
	c.Assert(validatorUpdates, IsNil)
	c.Assert(vMgr2.Meta.Nominated.IsEmpty(), Equals, false)
	c.Assert(vMgr2.Meta.Queued.IsEmpty(), Equals, true)
	c.Assert(vMgr2.Meta.Nominated, HasLen, 1)
	c.Assert(vMgr2.Meta.Nominated[0].Equals(standbyNode), Equals, true)

	nominatedNode := vMgr2.Meta.Nominated
	c.Assert(nominatedNode, HasLen, 1)
	// nominated node is not in ready status abandon the rotation
	rotateAtHeight := vMgr2.Meta.RotateAtBlockHeight
	ctx = ctx.WithBlockHeight(rotateAtHeight)
	validatorUpdates = vMgr2.EndBlock(ctx)
	c.Assert(validatorUpdates, IsNil)
	c.Assert(vMgr2.Meta.Nominated.IsEmpty(), Equals, true)
	c.Assert(vMgr2.Meta.Queued.IsEmpty(), Equals, true)
	nNode, err := k.GetNodeAccount(ctx, nominatedNode[0].NodeAddress)
	c.Assert(err, IsNil)
	c.Assert(nNode.Status, Equals, NodeStandby)

	// rotate validator, all good
	// nominatedNode need to be in ready status
	openWindow = vMgr2.Meta.RotateWindowOpenAtBlockHeight
	ctx = ctx.WithBlockHeight(openWindow)
	validatorUpdates = vMgr2.EndBlock(ctx)
	c.Assert(validatorUpdates, IsNil)
	nNode.UpdateStatus(NodeReady, openWindow)
	k.SetNodeAccount(ctx, nNode)

	rotateAtHeight = vMgr2.Meta.RotateAtBlockHeight
	ctx = ctx.WithBlockHeight(rotateAtHeight)
	validatorUpdates = vMgr2.EndBlock(ctx)
	c.Assert(validatorUpdates, NotNil)
	c.Assert(vMgr2.Meta.Nominated.IsEmpty(), Equals, true)
	c.Assert(vMgr2.Meta.Queued.IsEmpty(), Equals, true)
	// get the node account from data store again
	nNode, err = k.GetNodeAccount(ctx, nominatedNode[0].NodeAddress)
	c.Assert(err, IsNil)
	c.Assert(nNode.Status, Equals, NodeActive)

}
func setNodeAccountsStatus(ctx sdk.Context, k Keeper, nas NodeAccounts, status NodeStatus) {
	for _, item := range nas {
		item.UpdateStatus(status, ctx.BlockHeight())
		k.SetNodeAccount(ctx, item)
	}
}
func (vts *ValidatorManagerTestSuite) TestRotation(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	for i := 0; i < 10; i++ {
		node := GetRandomNodeAccount(NodeStandby)
		w.keeper.SetNodeAccount(w.ctx, node)
	}
	// we should rotate two in , and don't rotate out
	windowOpenAt := w.validatorMgr.Meta.RotateWindowOpenAtBlockHeight
	ctx := w.ctx.WithBlockHeight(windowOpenAt)
	w.validatorMgr.BeginBlock(ctx, windowOpenAt)
	validatorUpdates := w.validatorMgr.EndBlock(ctx)
	// nominated two nodes
	c.Assert(validatorUpdates, IsNil)
	c.Assert(w.validatorMgr.Meta.Nominated, HasLen, 2)
	c.Assert(w.validatorMgr.Meta.Queued, HasLen, 0)

	// set the nominated node as ready
	setNodeAccountsStatus(ctx, w.keeper, w.validatorMgr.Meta.Nominated, NodeReady)
	rotateAt := w.validatorMgr.Meta.RotateAtBlockHeight
	ctx = w.ctx.WithBlockHeight(rotateAt)
	w.validatorMgr.BeginBlock(ctx, rotateAt)
	validatorUpdates = w.validatorMgr.EndBlock(ctx)
	// we should have three active validators now
	c.Assert(validatorUpdates, HasLen, 3)
	c.Assert(w.validatorMgr.Meta.Queued, IsNil)
	c.Assert(w.validatorMgr.Meta.Nominated, IsNil)

	// do another two
	windowOpenAt = w.validatorMgr.Meta.RotateWindowOpenAtBlockHeight
	ctx = w.ctx.WithBlockHeight(windowOpenAt)
	w.validatorMgr.BeginBlock(ctx, windowOpenAt)
	validatorUpdates = w.validatorMgr.EndBlock(ctx)
	c.Assert(validatorUpdates, IsNil)

	setNodeAccountsStatus(ctx, w.keeper, w.validatorMgr.Meta.Nominated, NodeReady)
	rotateAt = w.validatorMgr.Meta.RotateAtBlockHeight
	ctx = w.ctx.WithBlockHeight(rotateAt)
	w.validatorMgr.BeginBlock(ctx, rotateAt)
	validatorUpdates = w.validatorMgr.EndBlock(ctx)
	c.Assert(validatorUpdates, HasLen, 5)

	for i := 0; i <= 27; i++ {
		node1 := GetRandomNodeAccount(NodeStandby)
		w.keeper.SetNodeAccount(w.ctx, node1)
		node2 := GetRandomNodeAccount(NodeStandby)
		w.keeper.SetNodeAccount(w.ctx, node2)

		windowOpenAt = w.validatorMgr.Meta.RotateWindowOpenAtBlockHeight
		ctx = w.ctx.WithBlockHeight(windowOpenAt)
		w.validatorMgr.BeginBlock(ctx, windowOpenAt)
		validatorUpdates = w.validatorMgr.EndBlock(ctx)
		c.Assert(validatorUpdates, IsNil)
		c.Assert(w.validatorMgr.Meta.Nominated, HasLen, 2)
		c.Assert(w.validatorMgr.Meta.Queued, HasLen, 1)

		setNodeAccountsStatus(ctx, w.keeper, w.validatorMgr.Meta.Nominated, NodeReady)
		rotateAt = w.validatorMgr.Meta.RotateAtBlockHeight
		ctx = w.ctx.WithBlockHeight(rotateAt)
		w.validatorMgr.BeginBlock(ctx, rotateAt)
		validatorUpdates = w.validatorMgr.EndBlock(ctx)
		c.Assert(validatorUpdates, HasLen, 7+i)
	}

	nodeA := GetRandomNodeAccount(NodeStandby)
	w.keeper.SetNodeAccount(w.ctx, nodeA)
	windowOpenAt = w.validatorMgr.Meta.RotateWindowOpenAtBlockHeight
	ctx = w.ctx.WithBlockHeight(windowOpenAt)
	w.validatorMgr.BeginBlock(ctx, windowOpenAt)
	validatorUpdates = w.validatorMgr.EndBlock(ctx)
	c.Assert(validatorUpdates, IsNil)
	c.Assert(w.validatorMgr.Meta.Nominated, HasLen, 1)
	c.Assert(w.validatorMgr.Meta.Queued, HasLen, 1)
	setNodeAccountsStatus(ctx, w.keeper, w.validatorMgr.Meta.Nominated, NodeReady)
	rotateAt = w.validatorMgr.Meta.RotateAtBlockHeight
	ctx = w.ctx.WithBlockHeight(rotateAt)
	w.validatorMgr.BeginBlock(ctx, rotateAt)
	validatorUpdates = w.validatorMgr.EndBlock(ctx)
	c.Assert(validatorUpdates, HasLen, 34)
}
