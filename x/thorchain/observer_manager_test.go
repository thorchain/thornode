package thorchain

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type ObserverManagerTestSuite struct{}

var _ = Suite(&ObserverManagerTestSuite{})

func (ObserverManagerTestSuite) TestObserverManager(c *C) {
	var err error
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
	addrs := mgr.List()
	// sort alphabetically
	sort.SliceStable(addrs, func(i, j int) bool { return addrs[i].String() > addrs[j].String() })
	expected := []sdk.AccAddress{a1, a2}
	sort.SliceStable(expected, func(i, j int) bool { return expected[i].String() > expected[j].String() })
	c.Check(addrs, DeepEquals, expected)

	mgr.EndBlock(ctx, k)
	addrs, err = k.GetObservingAddresses(ctx)
	c.Assert(err, IsNil)
	c.Check(addrs, HasLen, 2)
	// sort alphabetically
	sort.SliceStable(addrs, func(i, j int) bool { return addrs[i].String() > addrs[j].String() })
	expected = []sdk.AccAddress{a1, a2}
	sort.SliceStable(expected, func(i, j int) bool { return expected[i].String() > expected[j].String() })
	c.Check(addrs, DeepEquals, expected)
}
