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
	poolAddrMgr := NewPoolAddressMgr(k)
	vMgr := NewValidatorMgr(k, poolAddrMgr)
	vMgr.meta = &ValidatorMeta{}
	c.Assert(vMgr, NotNil)
	err := vMgr.setupValidatorNodes(ctx, 0)
	c.Assert(err, IsNil)

	// no node accounts at all
	err = vMgr.setupValidatorNodes(ctx, 1)
	c.Assert(err, NotNil)

	activeNode := GetRandomNodeAccount(NodeActive)
	c.Assert(k.SetNodeAccount(ctx, activeNode), IsNil)
	vMgr.rotationPolicy = GetValidatorRotationPolicy()

	err = vMgr.setupValidatorNodes(ctx, 1)
	c.Assert(err, IsNil)
	c.Assert(vMgr.meta.RotateAtBlockHeight, Equals, int64(rotatePerBlockHeight+1))
	c.Assert(vMgr.meta.RotateWindowOpenAtBlockHeight, Equals, int64(rotatePerBlockHeight+1-validatorChangeWindow))

	readyNode := GetRandomNodeAccount(NodeReady)
	c.Assert(k.SetNodeAccount(ctx, readyNode), IsNil)

	// one active node and one ready node on start up
	// it should take both of the node as active
	vMgr1 := NewValidatorMgr(k, poolAddrMgr)

	vMgr1.BeginBlock(ctx)
	c.Assert(vMgr1.meta, NotNil)
	c.Assert(vMgr1.meta.RotateAtBlockHeight, Equals, int64(rotatePerBlockHeight+1))
	c.Assert(vMgr1.meta.RotateWindowOpenAtBlockHeight, Equals, int64(rotatePerBlockHeight+1-validatorChangeWindow))
	c.Assert(vMgr1.meta.Queued.IsEmpty(), Equals, true)
	c.Assert(vMgr1.meta.Nominated.IsEmpty(), Equals, true)
	activeNodes, err := k.ListActiveNodeAccounts(ctx)
	c.Assert(err, IsNil)
	c.Logf("active nodes:%s", activeNodes)
	c.Assert(len(activeNodes) == 2, Equals, true)

	activeNode1 := GetRandomNodeAccount(NodeActive)
	activeNode2 := GetRandomNodeAccount(NodeActive)
	c.Assert(k.SetNodeAccount(ctx, activeNode1), IsNil)
	c.Assert(k.SetNodeAccount(ctx, activeNode2), IsNil)

	// three active nodes and 1 ready nodes, it should take them all
	vMgr2 := NewValidatorMgr(k, poolAddrMgr)
	vMgr2.BeginBlock(ctx)

	c.Assert(vMgr2.meta, NotNil)
	c.Assert(vMgr2.meta.RotateAtBlockHeight, Equals, int64(rotatePerBlockHeight+1))
	c.Assert(vMgr2.meta.RotateWindowOpenAtBlockHeight, Equals, int64(rotatePerBlockHeight+1-validatorChangeWindow))
	c.Assert(vMgr2.meta.Queued.IsEmpty(), Equals, true)
	c.Assert(vMgr2.meta.Nominated.IsEmpty(), Equals, true)

	activeNodes1, err := k.ListActiveNodeAccounts(ctx)
	c.Assert(err, IsNil)
	c.Assert(len(activeNodes1) == 4, Equals, true)
	// No standby nodes
	ctx = ctx.WithBlockHeight(rotatePerBlockHeight + 1 - validatorChangeWindow)
	txOutStore := NewTxOutStorage(k, poolAddrMgr)
	txOutStore.NewBlock(uint64(rotatePerBlockHeight + 1 - validatorChangeWindow))
	validatorUpdates := vMgr2.EndBlock(ctx, txOutStore)
	c.Assert(validatorUpdates, IsNil)
	c.Assert(vMgr2.meta.Nominated.IsEmpty(), Equals, true)
	c.Assert(vMgr2.meta.Queued.IsEmpty(), Equals, true)

	rotateHeight := rotatePerBlockHeight + 1
	ctx = ctx.WithBlockHeight(rotateHeight)
	txOutStore.NewBlock(uint64(rotateHeight))
	validatorUpdates = vMgr2.EndBlock(ctx, txOutStore)
	c.Assert(validatorUpdates, IsNil)
	c.Assert(vMgr2.meta.RotateWindowOpenAtBlockHeight, Equals, int64(rotatePerBlockHeight*2+1-validatorChangeWindow))
	c.Assert(vMgr2.meta.RotateAtBlockHeight, Equals, int64(rotatePerBlockHeight*2+1))
	c.Assert(vMgr2.meta.Nominated.IsEmpty(), Equals, true)
	c.Assert(vMgr2.meta.Queued.IsEmpty(), Equals, true)

	standbyNode := GetRandomNodeAccount(NodeStandby)
	c.Assert(k.SetNodeAccount(ctx, standbyNode), IsNil)

	// vts.setDesireValidatorSet(c, ctx, k)
	vMgr2.rotationPolicy = GetValidatorRotationPolicy()
	openWindow := vMgr2.meta.RotateWindowOpenAtBlockHeight
	ctx = ctx.WithBlockHeight(openWindow)
	txOutStore.NewBlock(uint64(openWindow))
	validatorUpdates = vMgr2.EndBlock(ctx, txOutStore)
	c.Assert(validatorUpdates, IsNil)
	c.Assert(vMgr2.meta.Nominated.IsEmpty(), Equals, false)
	c.Assert(vMgr2.meta.Queued.IsEmpty(), Equals, true)
	c.Assert(vMgr2.meta.Nominated, HasLen, 1)
	c.Assert(vMgr2.meta.Nominated[0].Equals(standbyNode), Equals, true)

	nominatedNode := vMgr2.meta.Nominated
	c.Assert(nominatedNode, HasLen, 1)
	// nominated node is not in ready status abandon the rotation
	rotateAtHeight := vMgr2.meta.RotateAtBlockHeight
	ctx = ctx.WithBlockHeight(rotateAtHeight)
	txOutStore.NewBlock(uint64(rotateAtHeight))
	validatorUpdates = vMgr2.EndBlock(ctx, txOutStore)
	c.Assert(validatorUpdates, IsNil)
	c.Assert(vMgr2.meta.Nominated.IsEmpty(), Equals, true)
	c.Assert(vMgr2.meta.Queued.IsEmpty(), Equals, true)
	nNode, err := k.GetNodeAccount(ctx, nominatedNode[0].NodeAddress)
	c.Assert(err, IsNil)
	c.Assert(nNode.Status, Equals, NodeStandby)

	// rotate validator, all good
	// nominatedNode need to be in ready status
	openWindow = vMgr2.meta.RotateWindowOpenAtBlockHeight
	ctx = ctx.WithBlockHeight(openWindow)
	txOutStore.NewBlock(uint64(openWindow))
	validatorUpdates = vMgr2.EndBlock(ctx, txOutStore)
	c.Assert(validatorUpdates, IsNil)
	nNode.UpdateStatus(NodeReady, openWindow)
	c.Assert(k.SetNodeAccount(ctx, nNode), IsNil)

	rotateAtHeight = vMgr2.meta.RotateAtBlockHeight
	ctx = ctx.WithBlockHeight(rotateAtHeight)
	txOutStore.NewBlock(uint64(rotateAtHeight))
	validatorUpdates = vMgr2.EndBlock(ctx, txOutStore)
	c.Assert(validatorUpdates, NotNil)
	c.Assert(vMgr2.meta.Nominated.IsEmpty(), Equals, true)
	c.Assert(vMgr2.meta.Queued.IsEmpty(), Equals, true)
	// get the node account from data store again
	nNode, err = k.GetNodeAccount(ctx, nominatedNode[0].NodeAddress)
	c.Assert(err, IsNil)
	c.Assert(nNode.Status, Equals, NodeActive)

}
func setNodeAccountsStatus(ctx sdk.Context, k Keeper, nas NodeAccounts, status NodeStatus, c *C) {
	for _, item := range nas {
		item.UpdateStatus(status, ctx.BlockHeight())
		c.Assert(k.SetNodeAccount(ctx, item), IsNil)
	}
}
func (vts *ValidatorManagerTestSuite) TestRotation(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	for i := 0; i < 10; i++ {
		node := GetRandomNodeAccount(NodeStandby)
		c.Assert(w.keeper.SetNodeAccount(w.ctx, node), IsNil)
	}
	// THORNode should rotate two in , and don't rotate out
	windowOpenAt := w.validatorMgr.Meta().RotateWindowOpenAtBlockHeight
	ctx := w.ctx.WithBlockHeight(windowOpenAt)
	w.validatorMgr.BeginBlock(ctx)
	w.txOutStore.NewBlock(uint64(windowOpenAt))
	validatorUpdates := w.validatorMgr.EndBlock(ctx, w.txOutStore)
	// nominated two nodes
	c.Assert(validatorUpdates, IsNil)
	c.Assert(w.validatorMgr.Meta().Nominated, HasLen, 2)
	c.Assert(w.validatorMgr.Meta().Queued, HasLen, 0)

	// set the nominated node as ready
	setNodeAccountsStatus(ctx, w.keeper, w.validatorMgr.Meta().Nominated, NodeReady, c)
	rotateAt := w.validatorMgr.Meta().RotateAtBlockHeight
	ctx = w.ctx.WithBlockHeight(rotateAt)
	w.validatorMgr.BeginBlock(ctx)
	w.txOutStore.NewBlock(uint64(windowOpenAt))
	validatorUpdates = w.validatorMgr.EndBlock(ctx, w.txOutStore)
	// THORNode should have three active validators now
	c.Assert(validatorUpdates, HasLen, 3)
	c.Assert(w.validatorMgr.Meta().Queued, IsNil)
	c.Assert(w.validatorMgr.Meta().Nominated, IsNil)

	// do another two
	windowOpenAt = w.validatorMgr.Meta().RotateWindowOpenAtBlockHeight
	ctx = w.ctx.WithBlockHeight(windowOpenAt)
	w.validatorMgr.BeginBlock(ctx)
	w.txOutStore.NewBlock(uint64(windowOpenAt))
	validatorUpdates = w.validatorMgr.EndBlock(ctx, w.txOutStore)
	c.Assert(validatorUpdates, IsNil)

	setNodeAccountsStatus(ctx, w.keeper, w.validatorMgr.Meta().Nominated, NodeReady, c)
	rotateAt = w.validatorMgr.Meta().RotateAtBlockHeight
	ctx = w.ctx.WithBlockHeight(rotateAt)
	w.validatorMgr.BeginBlock(ctx)
	w.txOutStore.NewBlock(uint64(rotateAt))
	validatorUpdates = w.validatorMgr.EndBlock(ctx, w.txOutStore)
	c.Assert(validatorUpdates, HasLen, 5)

	for i := 0; i <= 27; i++ {
		node1 := GetRandomNodeAccount(NodeStandby)
		c.Assert(w.keeper.SetNodeAccount(w.ctx, node1), IsNil)
		node2 := GetRandomNodeAccount(NodeStandby)
		c.Assert(w.keeper.SetNodeAccount(w.ctx, node2), IsNil)

		windowOpenAt = w.validatorMgr.Meta().RotateWindowOpenAtBlockHeight
		ctx = w.ctx.WithBlockHeight(windowOpenAt)
		w.validatorMgr.BeginBlock(ctx)
		w.txOutStore.NewBlock(uint64(windowOpenAt))
		validatorUpdates = w.validatorMgr.EndBlock(ctx, w.txOutStore)
		c.Assert(validatorUpdates, IsNil)
		c.Assert(w.validatorMgr.Meta().Nominated, HasLen, 2)
		c.Assert(w.validatorMgr.Meta().Queued, HasLen, 1)

		setNodeAccountsStatus(ctx, w.keeper, w.validatorMgr.Meta().Nominated, NodeReady, c)
		rotateAt = w.validatorMgr.Meta().RotateAtBlockHeight
		ctx = w.ctx.WithBlockHeight(rotateAt)
		w.validatorMgr.BeginBlock(ctx)
		w.txOutStore.NewBlock(uint64(rotateAt))
		validatorUpdates = w.validatorMgr.EndBlock(ctx, w.txOutStore)
		c.Assert(validatorUpdates, HasLen, 7+i)
	}

	nodeA := GetRandomNodeAccount(NodeStandby)
	c.Assert(w.keeper.SetNodeAccount(w.ctx, nodeA), IsNil)
	windowOpenAt = w.validatorMgr.Meta().RotateWindowOpenAtBlockHeight
	ctx = w.ctx.WithBlockHeight(windowOpenAt)
	w.validatorMgr.BeginBlock(ctx)
	w.txOutStore.NewBlock(uint64(windowOpenAt))
	validatorUpdates = w.validatorMgr.EndBlock(ctx, w.txOutStore)
	c.Assert(validatorUpdates, IsNil)
	c.Assert(w.validatorMgr.Meta().Nominated, HasLen, 1)
	c.Assert(w.validatorMgr.Meta().Queued, HasLen, 1)
	setNodeAccountsStatus(ctx, w.keeper, w.validatorMgr.Meta().Nominated, NodeReady, c)
	rotateAt = w.validatorMgr.Meta().RotateAtBlockHeight
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
		c.Assert(w.keeper.SetNodeAccount(w.ctx, node), IsNil)
		allNodes = append(allNodes, node)
	}
	// this should trick validator manager to take four nodes as active.
	ctx := w.ctx.WithBlockHeight(1)
	validatorMgr := NewValidatorMgr(w.keeper, w.poolAddrMgr)
	validatorMgr.BeginBlock(ctx)
	w.validatorMgr = validatorMgr
	// set one node to leave
	w.validatorMgr.Meta().LeaveQueue = append(w.validatorMgr.Meta().LeaveQueue, allNodes[0])
	height := w.validatorMgr.Meta().LeaveOpenWindow
	w.txOutStore.NewBlock(uint64(height))
	ctx = w.ctx.WithBlockHeight(height)
	w.validatorMgr.BeginBlock(ctx)
	w.validatorMgr.EndBlock(ctx, w.txOutStore)
	c.Assert(w.validatorMgr.Meta().Queued, HasLen, 1)
	// THORNode don't have enough standby node to be rotate in
	c.Assert(w.validatorMgr.Meta().Nominated, HasLen, 0)
	// make sure THORNode trigger a pool rotation as well
	c.Assert(w.poolAddrMgr.GetCurrentPoolAddresses().RotateWindowOpenAt, Equals, ctx.BlockHeight()+1)
	c.Assert(w.poolAddrMgr.GetCurrentPoolAddresses().RotateAt, Equals, w.validatorMgr.Meta().LeaveProcessAt)

	ctx = w.ctx.WithBlockHeight(w.validatorMgr.Meta().LeaveProcessAt)
	rotateWindowOpen := w.validatorMgr.Meta().RotateWindowOpenAtBlockHeight
	rotateAt := w.validatorMgr.Meta().RotateAtBlockHeight
	leaveWindowOpen := w.validatorMgr.Meta().LeaveOpenWindow
	leaveAt := w.validatorMgr.Meta().LeaveProcessAt
	w.validatorMgr.BeginBlock(ctx)
	w.txOutStore.NewBlock(uint64(w.validatorMgr.Meta().LeaveProcessAt))
	updates := w.validatorMgr.EndBlock(ctx, w.txOutStore)
	// THORNode don't have yggdrasil fund
	c.Assert(w.txOutStore.GetOutboundItems(), HasLen, 0)
	c.Assert(updates, HasLen, 7)
	// make sure scheduled rotation window get extended
	c.Assert(w.validatorMgr.Meta().RotateAtBlockHeight, Equals, rotateAt+w.validatorMgr.RotationPolicy().LeaveProcessPerBlockHeight)
	c.Assert(w.validatorMgr.Meta().RotateWindowOpenAtBlockHeight, Equals, rotateWindowOpen+w.validatorMgr.RotationPolicy().LeaveProcessPerBlockHeight)
	c.Assert(w.validatorMgr.Meta().LeaveOpenWindow, Equals, leaveWindowOpen+w.validatorMgr.RotationPolicy().LeaveProcessPerBlockHeight)
	c.Assert(w.validatorMgr.Meta().LeaveProcessAt, Equals, leaveAt+w.validatorMgr.RotationPolicy().LeaveProcessPerBlockHeight)

}

func (ValidatorManagerTestSuite) TestRagnarokProtocol(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	allNodes := NodeAccounts{}
	// add 6 ndoes
	for i := 1; i <= 6; i++ {
		node := GetRandomNodeAccount(NodeReady)
		c.Assert(w.keeper.SetNodeAccount(w.ctx, node), IsNil)
		allNodes = append(allNodes, node)
	}
	// this should trick validator manager to take four nodes as active.
	ctx := w.ctx.WithBlockHeight(1)
	validatorMgr := NewValidatorMgr(w.keeper, w.poolAddrMgr)
	validatorMgr.BeginBlock(ctx)
	w.validatorMgr = validatorMgr
	// set one node to leave
	for i := 0; i <= 4; i++ {
		w.validatorMgr.Meta().LeaveQueue = append(w.validatorMgr.Meta().LeaveQueue, allNodes[i])
	}
	height := w.validatorMgr.Meta().LeaveOpenWindow
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
	c.Assert(w.validatorMgr.Meta().Ragnarok, Equals, true)
}
