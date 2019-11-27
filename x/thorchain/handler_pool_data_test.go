package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type HandlerPoolDataSuite struct{}

type TestPoolKeeper struct {
	KVStoreDummy
	na     NodeAccount
	asset  common.Asset
	status PoolStatus
}

func (k *TestPoolKeeper) GetNodeAccount(ctx sdk.Context, signer sdk.AccAddress) (NodeAccount, error) {
	return k.na, nil
}

func (k *TestPoolKeeper) SetPoolData(ctx sdk.Context, asset common.Asset, status PoolStatus) {
	k.asset = asset
	k.status = status
}

var _ = Suite(&HandlerPoolDataSuite{})

func (s *HandlerPoolDataSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestPoolKeeper{
		na: GetRandomNodeAccount(NodeActive),
	}

	handler := NewPoolDataHandler(keeper)
	// happy path
	msg := NewMsgSetPoolData(common.BNBAsset, PoolEnabled, keeper.na.NodeAddress)
	err := handler.Validate(ctx, msg, 1)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.Validate(ctx, msg, 0)
	c.Assert(err, Equals, badVersion)

	// inactive node account
	keeper.na = GetRandomNodeAccount(NodeStandby)
	msg = NewMsgSetPoolData(common.BNBAsset, PoolEnabled, keeper.na.NodeAddress)
	err = handler.Validate(ctx, msg, 1)
	c.Assert(err, Equals, notAuthorized)

	// invalid msg
	msg = MsgSetPoolData{}
	err = handler.Validate(ctx, msg, 1)
	c.Assert(err, NotNil)
}

func (s *HandlerPoolDataSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestPoolKeeper{}

	handler := NewPoolDataHandler(keeper)

	msg := NewMsgSetPoolData(common.BNBAsset, PoolEnabled, GetRandomBech32Addr())
	err := handler.Handle(ctx, msg, 1)
	c.Assert(err, IsNil)
	c.Check(keeper.asset.Equals(common.BNBAsset), Equals, true)
	c.Check(keeper.status, Equals, PoolEnabled)
}
