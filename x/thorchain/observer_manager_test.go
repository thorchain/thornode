package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type ObserverManagerTestSuite struct{}

var _ = Suite(&ObserverManagerTestSuite{})

func (ObserverManagerTestSuite) TestObserverManager(c *C) {
	ctx, k := setupKeeperForTest(c)
	mgr := NewObserverMgr()
	c.Assert(mgr, NotNil)
	mgr.BeginBlock()
	c.Check(mgr.List(), HasLen, 0)

	a1 := GetRandomBech32Addr()
	a2 := GetRandomBech32Addr()
	a3 := GetRandomBech32Addr()
	mgr.AppendObserver(common.BNBChain, []sdk.AccAddress{
		a1, a2, a3,
	})
	c.Check(mgr.List(), HasLen, 3)
	mgr.AppendObserver(common.BTCChain, []sdk.AccAddress{
		a1, a2,
	})
	c.Check(mgr.List(), HasLen, 2)
	c.Check(mgr.List(), DeepEquals, []sdk.AccAddress{a1, a2})

	mgr.EndBlock(ctx, k)
	addrs, err := k.GetObservingAddresses(ctx)
	c.Assert(err, IsNil)
	c.Check(addrs, HasLen, 2)
	c.Check(addrs, DeepEquals, []sdk.AccAddress{a1, a2})
}
