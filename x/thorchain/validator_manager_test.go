package thorchain

import (
	"github.com/blang/semver"
	"gitlab.com/thorchain/thornode/constants"
	. "gopkg.in/check.v1"
)

type ValidatorManagerTestSuite struct{}

var _ = Suite(&ValidatorManagerTestSuite{})

func (vts *ValidatorManagerTestSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

func (vts *ValidatorManagerTestSuite) TestSetupValidatorNodes(c *C) {
	ctx, k := setupKeeperForTest(c)
	ctx = ctx.WithBlockHeight(1)
	poolAddrMgr := NewPoolAddressDummyMgr()
	k.SetPoolAddresses(ctx, poolAddrMgr.GetCurrentPoolAddresses())
	vaultMgr := NewVaultMgrDummy()
	vMgr := NewValidatorMgr(k, poolAddrMgr, vaultMgr)
	c.Assert(vMgr, NotNil)
	ver := semver.MustParse("0.1.0")
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
	vMgr1 := NewValidatorMgr(k, poolAddrMgr, vaultMgr)

	vMgr1.BeginBlock(ctx, constAccessor)
	activeNodes, err := k.ListActiveNodeAccounts(ctx)
	c.Assert(err, IsNil)
	c.Logf("active nodes:%s", activeNodes)
	c.Assert(len(activeNodes) == 2, Equals, true)

	activeNode1 := GetRandomNodeAccount(NodeActive)
	activeNode2 := GetRandomNodeAccount(NodeActive)
	c.Assert(k.SetNodeAccount(ctx, activeNode1), IsNil)
	c.Assert(k.SetNodeAccount(ctx, activeNode2), IsNil)

	// three active nodes and 1 ready nodes, it should take them all
	vMgr2 := NewValidatorMgr(k, poolAddrMgr, vaultMgr)
	vMgr2.BeginBlock(ctx, constAccessor)

	activeNodes1, err := k.ListActiveNodeAccounts(ctx)
	c.Assert(err, IsNil)
	c.Assert(len(activeNodes1) == 4, Equals, true)
}
