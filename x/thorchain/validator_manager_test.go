package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type ValidatorManagerTestSuite struct{}

var _ = Suite(&ValidatorManagerTestSuite{})

func (vts *ValidatorManagerTestSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

func (vts *ValidatorManagerTestSuite) TestSetupValidatorNodes(c *C) {
	ctx, k := setupKeeperForTest(c)
	ctx = ctx.WithBlockHeight(1)
	rotatePerBlockHeight := int64(constants.RotatePerBlockHeight)
	validatorChangeWindow := int64(constants.ValidatorsChangeWindow)
	poolAddrMgr := NewPoolAddressManager(k)
	vMgr := NewValidatorManager(k, poolAddrMgr)
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
	vMgr1 := NewValidatorManager(k, poolAddrMgr)

	vMgr1.BeginBlock(ctx)
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
	vMgr2 := NewValidatorManager(k, poolAddrMgr)
	vMgr2.BeginBlock(ctx)

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
	txOutStore := NewTxOutStore(k, poolAddrMgr)
	txOutStore.NewBlock(uint64(rotatePerBlockHeight + 1 - validatorChangeWindow))
	validatorUpdates := vMgr2.EndBlock(ctx, txOutStore)
	c.Assert(validatorUpdates, IsNil)
	c.Assert(vMgr2.Meta.Nominated.IsEmpty(), Equals, true)
	c.Assert(vMgr2.Meta.Queued.IsEmpty(), Equals, true)

	rotateHeight := rotatePerBlockHeight + 1
	ctx = ctx.WithBlockHeight(rotateHeight)
	txOutStore.NewBlock(uint64(rotateHeight))
	validatorUpdates = vMgr2.EndBlock(ctx, txOutStore)
	c.Assert(validatorUpdates, IsNil)
	c.Assert(vMgr2.Meta.RotateWindowOpenAtBlockHeight, Equals, int64(rotatePerBlockHeight*2+1-validatorChangeWindow))
	c.Assert(vMgr2.Meta.RotateAtBlockHeight, Equals, int64(rotatePerBlockHeight*2+1))
	c.Assert(vMgr2.Meta.Nominated.IsEmpty(), Equals, true)
	c.Assert(vMgr2.Meta.Queued.IsEmpty(), Equals, true)

	standbyNode := GetRandomNodeAccount(NodeStandby)
	k.SetNodeAccount(ctx, standbyNode)

	// vts.setDesireValidatorSet(c, ctx, k)
	vMgr2.rotationPolicy = GetValidatorRotationPolicy(ctx, k)
	openWindow := vMgr2.Meta.RotateWindowOpenAtBlockHeight
	ctx = ctx.WithBlockHeight(openWindow)
	txOutStore.NewBlock(uint64(openWindow))
	validatorUpdates = vMgr2.EndBlock(ctx, txOutStore)
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
	txOutStore.NewBlock(uint64(rotateAtHeight))
	validatorUpdates = vMgr2.EndBlock(ctx, txOutStore)
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
	txOutStore.NewBlock(uint64(openWindow))
	validatorUpdates = vMgr2.EndBlock(ctx, txOutStore)
	c.Assert(validatorUpdates, IsNil)
	nNode.UpdateStatus(NodeReady, openWindow)
	k.SetNodeAccount(ctx, nNode)

	rotateAtHeight = vMgr2.Meta.RotateAtBlockHeight
	ctx = ctx.WithBlockHeight(rotateAtHeight)
	txOutStore.NewBlock(uint64(rotateAtHeight))
	validatorUpdates = vMgr2.EndBlock(ctx, txOutStore)
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
	w.validatorMgr.BeginBlock(ctx)
	w.txOutStore.NewBlock(uint64(windowOpenAt))
	validatorUpdates := w.validatorMgr.EndBlock(ctx, w.txOutStore)
	// nominated two nodes
	c.Assert(validatorUpdates, IsNil)
	c.Assert(w.validatorMgr.Meta.Nominated, HasLen, 2)
	c.Assert(w.validatorMgr.Meta.Queued, HasLen, 0)

	// set the nominated node as ready
	setNodeAccountsStatus(ctx, w.keeper, w.validatorMgr.Meta.Nominated, NodeReady)
	rotateAt := w.validatorMgr.Meta.RotateAtBlockHeight
	ctx = w.ctx.WithBlockHeight(rotateAt)
	w.validatorMgr.BeginBlock(ctx)
	w.txOutStore.NewBlock(uint64(windowOpenAt))
	validatorUpdates = w.validatorMgr.EndBlock(ctx, w.txOutStore)
	// we should have three active validators now
	c.Assert(validatorUpdates, HasLen, 3)
	c.Assert(w.validatorMgr.Meta.Queued, IsNil)
	c.Assert(w.validatorMgr.Meta.Nominated, IsNil)

	// do another two
	windowOpenAt = w.validatorMgr.Meta.RotateWindowOpenAtBlockHeight
	ctx = w.ctx.WithBlockHeight(windowOpenAt)
	w.validatorMgr.BeginBlock(ctx)
	w.txOutStore.NewBlock(uint64(windowOpenAt))
	validatorUpdates = w.validatorMgr.EndBlock(ctx, w.txOutStore)
	c.Assert(validatorUpdates, IsNil)

	setNodeAccountsStatus(ctx, w.keeper, w.validatorMgr.Meta.Nominated, NodeReady)
	rotateAt = w.validatorMgr.Meta.RotateAtBlockHeight
	ctx = w.ctx.WithBlockHeight(rotateAt)
	w.validatorMgr.BeginBlock(ctx)
	w.txOutStore.NewBlock(uint64(rotateAt))
	validatorUpdates = w.validatorMgr.EndBlock(ctx, w.txOutStore)
	c.Assert(validatorUpdates, HasLen, 5)

	for i := 0; i <= 27; i++ {
		node1 := GetRandomNodeAccount(NodeStandby)
		w.keeper.SetNodeAccount(w.ctx, node1)
		node2 := GetRandomNodeAccount(NodeStandby)
		w.keeper.SetNodeAccount(w.ctx, node2)

		windowOpenAt = w.validatorMgr.Meta.RotateWindowOpenAtBlockHeight
		ctx = w.ctx.WithBlockHeight(windowOpenAt)
		w.validatorMgr.BeginBlock(ctx)
		w.txOutStore.NewBlock(uint64(windowOpenAt))
		validatorUpdates = w.validatorMgr.EndBlock(ctx, w.txOutStore)
		c.Assert(validatorUpdates, IsNil)
		c.Assert(w.validatorMgr.Meta.Nominated, HasLen, 2)
		c.Assert(w.validatorMgr.Meta.Queued, HasLen, 1)

		setNodeAccountsStatus(ctx, w.keeper, w.validatorMgr.Meta.Nominated, NodeReady)
		rotateAt = w.validatorMgr.Meta.RotateAtBlockHeight
		ctx = w.ctx.WithBlockHeight(rotateAt)
		w.validatorMgr.BeginBlock(ctx)
		w.txOutStore.NewBlock(uint64(rotateAt))
		validatorUpdates = w.validatorMgr.EndBlock(ctx, w.txOutStore)
		c.Assert(validatorUpdates, HasLen, 7+i)
	}

	nodeA := GetRandomNodeAccount(NodeStandby)
	w.keeper.SetNodeAccount(w.ctx, nodeA)
	windowOpenAt = w.validatorMgr.Meta.RotateWindowOpenAtBlockHeight
	ctx = w.ctx.WithBlockHeight(windowOpenAt)
	w.validatorMgr.BeginBlock(ctx)
	w.txOutStore.NewBlock(uint64(windowOpenAt))
	validatorUpdates = w.validatorMgr.EndBlock(ctx, w.txOutStore)
	c.Assert(validatorUpdates, IsNil)
	c.Assert(w.validatorMgr.Meta.Nominated, HasLen, 1)
	c.Assert(w.validatorMgr.Meta.Queued, HasLen, 1)
	setNodeAccountsStatus(ctx, w.keeper, w.validatorMgr.Meta.Nominated, NodeReady)
	rotateAt = w.validatorMgr.Meta.RotateAtBlockHeight
	ctx = w.ctx.WithBlockHeight(rotateAt)
	w.validatorMgr.BeginBlock(ctx)
	w.txOutStore.NewBlock(uint64(rotateAt))
	validatorUpdates = w.validatorMgr.EndBlock(ctx, w.txOutStore)
	c.Assert(validatorUpdates, HasLen, 34)
}

func (ValidatorManagerTestSuite) TestValidatorsLeave(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	allNodes := NodeAccounts{}
	// add 6 ndoes
	for i := 1; i <= 6; i++ {
		node := GetRandomNodeAccount(NodeReady)
		w.keeper.SetNodeAccount(w.ctx, node)
		allNodes = append(allNodes, node)
	}
	// this should trick validator manager to take four nodes as active.
	ctx := w.ctx.WithBlockHeight(1)
	validatorMgr := NewValidatorManager(w.keeper, w.poolAddrMgr)
	validatorMgr.BeginBlock(ctx)
	w.validatorMgr = validatorMgr
	// set one node to leave
	w.validatorMgr.Meta.LeaveQueue = append(w.validatorMgr.Meta.LeaveQueue, allNodes[0])
	height := w.validatorMgr.Meta.LeaveOpenWindow
	w.txOutStore.NewBlock(uint64(height))
	ctx = w.ctx.WithBlockHeight(height)
	w.validatorMgr.BeginBlock(ctx)
	w.validatorMgr.EndBlock(ctx, w.txOutStore)
	c.Assert(w.validatorMgr.Meta.Queued, HasLen, 1)
	// we don't have enough standby node to be rotate in
	c.Assert(w.validatorMgr.Meta.Nominated, HasLen, 0)
	// make sure we trigger a pool rotation as well
	c.Assert(w.poolAddrMgr.currentPoolAddresses.RotateWindowOpenAt, Equals, ctx.BlockHeight()+1)
	c.Assert(w.poolAddrMgr.currentPoolAddresses.RotateAt, Equals, w.validatorMgr.Meta.LeaveProcessAt)

	ctx = w.ctx.WithBlockHeight(w.validatorMgr.Meta.LeaveProcessAt)
	rotateWindowOpen := w.validatorMgr.Meta.RotateWindowOpenAtBlockHeight
	rotateAt := w.validatorMgr.Meta.RotateAtBlockHeight
	leaveWindowOpen := w.validatorMgr.Meta.LeaveOpenWindow
	leaveAt := w.validatorMgr.Meta.LeaveProcessAt
	w.validatorMgr.BeginBlock(ctx)
	w.txOutStore.NewBlock(uint64(w.validatorMgr.Meta.LeaveProcessAt))
	updates := w.validatorMgr.EndBlock(ctx, w.txOutStore)
	// we don't have yggdrasil fund
	c.Assert(w.txOutStore.blockOut.TxArray, HasLen, 0)
	c.Assert(updates, HasLen, 7)
	// make sure scheduled rotation window get extended
	c.Assert(w.validatorMgr.Meta.RotateAtBlockHeight, Equals, rotateAt+w.validatorMgr.rotationPolicy.LeaveProcessPerBlockHeight)
	c.Assert(w.validatorMgr.Meta.RotateWindowOpenAtBlockHeight, Equals, rotateWindowOpen+w.validatorMgr.rotationPolicy.LeaveProcessPerBlockHeight)
	c.Assert(w.validatorMgr.Meta.LeaveOpenWindow, Equals, leaveWindowOpen+w.validatorMgr.rotationPolicy.LeaveProcessPerBlockHeight)
	c.Assert(w.validatorMgr.Meta.LeaveProcessAt, Equals, leaveAt+w.validatorMgr.rotationPolicy.LeaveProcessPerBlockHeight)

}

func (ValidatorManagerTestSuite) TestRagnarokProtocol(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	allNodes := NodeAccounts{}
	// add 6 ndoes
	for i := 1; i <= 6; i++ {
		node := GetRandomNodeAccount(NodeReady)
		w.keeper.SetNodeAccount(w.ctx, node)
		allNodes = append(allNodes, node)
	}
	// this should trick validator manager to take four nodes as active.
	ctx := w.ctx.WithBlockHeight(1)
	validatorMgr := NewValidatorManager(w.keeper, w.poolAddrMgr)
	validatorMgr.BeginBlock(ctx)
	w.validatorMgr = validatorMgr
	// set one node to leave
	for i := 0; i <= 4; i++ {
		w.validatorMgr.Meta.LeaveQueue = append(w.validatorMgr.Meta.LeaveQueue, allNodes[i])
	}
	height := w.validatorMgr.Meta.LeaveOpenWindow
	tx := common.NewTx(
		GetRandomTxHash(),
		GetRandomBNBAddress(),
		GetRandomBNBAddress(),
		common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(common.One*100)),
			common.NewCoin(common.RuneAsset(), sdk.NewUint(common.One*100)),
		},
		nil,
		"stake:BNB",
	)
	msg := NewMsgSetStakeData(tx,
		common.BNBAsset,
		sdk.NewUint(common.One*100),
		sdk.NewUint(common.One*100),
		tx.FromAddress,
		tx.FromAddress,
		allNodes[0].NodeAddress)
	// add a staker
	handleMsgSetStakeData(w.ctx, w.keeper, msg)

	w.txOutStore.NewBlock(uint64(height))
	ctx = w.ctx.WithBlockHeight(height)
	w.validatorMgr.BeginBlock(ctx)
	w.validatorMgr.EndBlock(ctx, w.txOutStore)
	c.Assert(w.validatorMgr.Meta.Ragnarok, Equals, true)
}
