package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type HandlerPoolDataSuite struct{}

type TestPoolKeeper struct {
	KVStoreDummy
	na   NodeAccount
	pool Pool
}

func (k *TestPoolKeeper) GetNodeAccount(ctx sdk.Context, signer sdk.AccAddress) (NodeAccount, error) {
	return k.na, nil
}

func (k *TestPoolKeeper) GetPool(ctx sdk.Context, asset common.Asset) (Pool, error) {
	return k.pool, nil
}

func (k *TestPoolKeeper) SetPool(ctx sdk.Context, pool Pool) error {
	fmt.Printf("Setting Pool: %+v\n", pool)
	k.pool = pool
	return nil
}

var _ = Suite(&HandlerPoolDataSuite{})

func (s *HandlerPoolDataSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestPoolKeeper{
		na: GetRandomNodeAccount(NodeActive),
	}

	handler := NewPoolDataHandler(keeper)
	// happy path
	ver := semver.MustParse("0.1.0")
	msg := NewMsgSetPoolData(common.BNBAsset, PoolEnabled, keeper.na.NodeAddress)
	err := handler.Validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.Validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, badVersion)

	// inactive node account
	keeper.na = GetRandomNodeAccount(NodeStandby)
	msg = NewMsgSetPoolData(common.BNBAsset, PoolEnabled, keeper.na.NodeAddress)
	err = handler.Validate(ctx, msg, ver)
	c.Assert(err, Equals, notAuthorized)

	// invalid msg
	msg = MsgSetPoolData{}
	err = handler.Validate(ctx, msg, ver)
	c.Assert(err, NotNil)
}

func (s *HandlerPoolDataSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)
	ver := semver.MustParse("0.1.0")

	keeper := &TestPoolKeeper{}

	handler := NewPoolDataHandler(keeper)

	msg := NewMsgSetPoolData(common.BNBAsset, PoolEnabled, GetRandomBech32Addr())
	err := handler.Handle(ctx, msg, ver)
	c.Assert(err, IsNil)
	c.Check(keeper.pool.Asset.Equals(common.BNBAsset), Equals, true, Commentf("%+v\n", keeper.pool))
	c.Check(keeper.pool.Status, Equals, PoolEnabled)
}
