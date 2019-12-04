package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type KeeperObserverSuite struct{}

var _ = Suite(&KeeperObserverSuite{})

func (s *KeeperObserverSuite) TestObserver(c *C) {
	ctx, k := setupKeeperForTest(c)

	addr := GetRandomBech32Addr()

	k.SetActiveObserver(ctx, addr)
	c.Check(k.IsActiveObserver(ctx, addr), Equals, true)
	k.RemoveActiveObserver(ctx, addr)
	c.Check(k.IsActiveObserver(ctx, addr), Equals, false)

	k.AddObservingAddresses(ctx, []sdk.AccAddress{addr})
	addrs, err := k.GetObservingAddresses(ctx)
	c.Assert(err, IsNil)
	c.Assert(addrs, HasLen, 1)
	c.Check(addrs[0].Equals(addr), Equals, true)

	k.ClearObservingAddresses(ctx)
	addrs, err := k.GetObservingAddresses(ctx)
	c.Assert(err, IsNil)
	c.Assert(addrs, HasLen, 0)
}
