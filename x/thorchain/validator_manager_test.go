package thorchain

import (
	. "gopkg.in/check.v1"

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
	poolAddrMgr := NewPoolAddressDummyMgr()
	k.SetPoolAddresses(ctx, poolAddrMgr.GetCurrentPoolAddresses())
	vMgr := NewValidatorMgr(k, poolAddrMgr)
	c.Assert(vMgr, NotNil)
	err := vMgr.setupValidatorNodes(ctx, 0)
	c.Assert(err, IsNil)

	// no node accounts at all
	err = vMgr.setupValidatorNodes(ctx, 1)
	c.Assert(err, NotNil)

	activeNode := GetRandomNodeAccount(NodeActive)
	c.Assert(k.SetNodeAccount(ctx, activeNode), IsNil)

	err = vMgr.setupValidatorNodes(ctx, 1)
	c.Assert(err, IsNil)

	readyNode := GetRandomNodeAccount(NodeReady)
	c.Assert(k.SetNodeAccount(ctx, readyNode), IsNil)

	// one active node and one ready node on start up
	// it should take both of the node as active
	vMgr1 := NewValidatorMgr(k, poolAddrMgr)

	vMgr1.BeginBlock(ctx)
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

	activeNodes1, err := k.ListActiveNodeAccounts(ctx)
	c.Assert(err, IsNil)
	c.Assert(len(activeNodes1) == 4, Equals, true)
	// No standby nodes
	ctx = ctx.WithBlockHeight(rotatePerBlockHeight + 1 - validatorChangeWindow)
	txOutStore := NewTxOutStorage(k, poolAddrMgr)
	txOutStore.NewBlock(uint64(rotatePerBlockHeight + 1 - validatorChangeWindow))
	validatorUpdates := vMgr2.EndBlock(ctx, txOutStore)
	c.Assert(validatorUpdates, HasLen, 0)

	rotateHeight := rotatePerBlockHeight + 1
	ctx = ctx.WithBlockHeight(rotateHeight)
	txOutStore.NewBlock(uint64(rotateHeight))
	validatorUpdates = vMgr2.EndBlock(ctx, txOutStore)
	c.Assert(validatorUpdates, HasLen, 0)

	standbyNode := GetRandomNodeAccount(NodeStandby)
	c.Assert(k.SetNodeAccount(ctx, standbyNode), IsNil)

	// vts.setDesireValidatorSet(c, ctx, k)
	validatorUpdates = vMgr2.EndBlock(ctx, txOutStore)
	c.Assert(validatorUpdates, HasLen, 0)
}
