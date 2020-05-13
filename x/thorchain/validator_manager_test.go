package thorchain

import (
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
	versionedTxOutStoreDummy := NewVersionedTxOutStoreDummy()
	versionedVaultMgrDummy := NewVersionedVaultMgrDummy(versionedTxOutStoreDummy)
	versionedEventManagerDummy := NewDummyVersionedEventMgr()

	vMgr := newValidatorMgrV1(k, versionedTxOutStoreDummy, versionedVaultMgrDummy, versionedEventManagerDummy)
	c.Assert(vMgr, NotNil)
	ver := constants.SWVersion
	constAccessor := constants.GetConstantValues(ver)
	err := vMgr.setupValidatorNodes(ctx, 0, constAccessor)
	c.Assert(err, IsNil)

	// no node accounts at all
	err = vMgr.setupValidatorNodes(ctx, 1, constAccessor)
	c.Assert(err, NotNil)

	activeNode := GetRandomNodeAccount(NodeActive)
	c.Assert(k.SetNodeAccount(ctx, activeNode), IsNil)

	err = vMgr.setupValidatorNodes(ctx, 1, constAccessor)
	c.Assert(err, IsNil)

	readyNode := GetRandomNodeAccount(NodeReady)
	c.Assert(k.SetNodeAccount(ctx, readyNode), IsNil)

	// one active node and one ready node on start up
	// it should take both of the node as active
	vMgr1 := newValidatorMgrV1(k, versionedTxOutStoreDummy, versionedVaultMgrDummy, versionedEventManagerDummy)

	c.Assert(vMgr1.BeginBlock(ctx, constAccessor), IsNil)
	activeNodes, err := k.ListActiveNodeAccounts(ctx)
	c.Assert(err, IsNil)
	c.Logf("active nodes:%s", activeNodes)
	c.Assert(len(activeNodes) == 2, Equals, true)

	activeNode1 := GetRandomNodeAccount(NodeActive)
	activeNode2 := GetRandomNodeAccount(NodeActive)
	c.Assert(k.SetNodeAccount(ctx, activeNode1), IsNil)
	c.Assert(k.SetNodeAccount(ctx, activeNode2), IsNil)

	// three active nodes and 1 ready nodes, it should take them all
	vMgr2 := newValidatorMgrV1(k, versionedTxOutStoreDummy, versionedVaultMgrDummy, versionedEventManagerDummy)
	vMgr2.BeginBlock(ctx, constAccessor)

	activeNodes1, err := k.ListActiveNodeAccounts(ctx)
	c.Assert(err, IsNil)
	c.Assert(len(activeNodes1) == 4, Equals, true)
}

func (vts *ValidatorManagerTestSuite) TestRagnarokForChaosnet(c *C) {
	ctx, k := setupKeeperForTest(c)
	versionedTxOutStoreDummy := NewVersionedTxOutStoreDummy()
	versionedVaultMgrDummy := NewVersionedVaultMgrDummy(versionedTxOutStoreDummy)
	versionedEventManagerDummy := NewDummyVersionedEventMgr()

	vMgr := newValidatorMgrV1(k, versionedTxOutStoreDummy, versionedVaultMgrDummy, versionedEventManagerDummy)
	c.Assert(vMgr, NotNil)

	constAccessor := constants.NewDummyConstants(map[constants.ConstantName]int64{
		constants.DesireValidatorSet:            12,
		constants.ArtificialRagnarokBlockHeight: 1024,
		constants.BadValidatorRate:              256,
		constants.OldValidatorRate:              256,
		constants.MinimumNodesForBFT:            4,
		constants.RotatePerBlockHeight:          256,
		constants.RotateRetryBlocks:             720,
	}, map[constants.ConstantName]bool{
		constants.StrictBondStakeRatio: false,
	}, map[constants.ConstantName]string{})
	for i := 0; i < 12; i++ {
		node := GetRandomNodeAccount(NodeReady)
		c.Assert(k.SetNodeAccount(ctx, node), IsNil)
	}
	c.Assert(vMgr.setupValidatorNodes(ctx, 1, constAccessor), IsNil)
	nodeAccounts, err := k.ListNodeAccountsByStatus(ctx, NodeActive)
	c.Assert(err, IsNil)
	c.Assert(len(nodeAccounts), Equals, 12)

	// trigger ragnarok
	ctx = ctx.WithBlockHeight(1024)
	c.Assert(vMgr.BeginBlock(ctx, constAccessor), IsNil)
	vault := NewVault(ctx.BlockHeight(), ActiveVault, AsgardVault, GetRandomPubKey(), common.Chains{common.BNBChain})
	for _, item := range versionedVaultMgrDummy.vaultMgrDummy.nas {
		vault.Membership = append(vault.Membership, item.PubKeySet.Secp256k1)
	}
	c.Assert(k.SetVault(ctx, vault), IsNil)
	updates := vMgr.EndBlock(ctx, constAccessor)
	// ragnarok , no one leaves
	c.Assert(updates, IsNil)
	ragnarokHeight, err := k.GetRagnarokBlockHeight(ctx)
	c.Assert(err, IsNil)
	c.Assert(ragnarokHeight == 1024, Equals, true, Commentf("%d == %d", ragnarokHeight, 1024))
}
