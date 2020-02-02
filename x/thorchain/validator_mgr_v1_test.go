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
