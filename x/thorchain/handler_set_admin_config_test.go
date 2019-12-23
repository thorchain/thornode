package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type HandlerSetAdminConfigSuite struct{}

type TestSetAdminConfigKeeper struct {
	KVStoreDummy
	na NodeAccount
}

func (k *TestSetAdminConfigKeeper) GetNodeAccount(_ sdk.Context, _ sdk.AccAddress) (NodeAccount, error) {
	return k.na, nil
}

var _ = Suite(&HandlerSetAdminConfigSuite{})

func (s *HandlerSetAdminConfigSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	na := GetRandomNodeAccount(NodeActive)
	keeper := &TestSetAdminConfigKeeper{
		na: na,
	}

	handler := NewSetAdminConfigHandler(keeper)

	// happy path
	ver := semver.MustParse("0.1.0")
	tx := GetRandomTx()
	msg := NewMsgSetAdminConfig(tx, PoolRefundGasKey, "1000", na.NodeAddress)
	err := handler.validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, badVersion)

	// invalid msg
	msg = MsgSetAdminConfig{}
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, NotNil)

	// not active node
	na = GetRandomNodeAccount(NodeWhiteListed)
	keeper = &TestSetAdminConfigKeeper{
		na: na,
	}
	handler = NewSetAdminConfigHandler(keeper)
	msg = NewMsgSetAdminConfig(tx, PoolRefundGasKey, "1000", na.NodeAddress)
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, Equals, notAuthorized)
}

type TestSetAdminConfigHandleKeeper struct {
	KVStoreDummy
	na NodeAccount
	ac AdminConfig
}

func (k *TestSetAdminConfigHandleKeeper) GetNodeAccount(_ sdk.Context, _ sdk.AccAddress) (NodeAccount, error) {
	return k.na, nil
}

func (k *TestSetAdminConfigHandleKeeper) GetAdminConfigValue(_ sdk.Context, _ AdminConfigKey, _ sdk.AccAddress) (string, error) {
	return k.ac.Value, nil
}

func (k *TestSetAdminConfigHandleKeeper) SetAdminConfig(_ sdk.Context, ac AdminConfig) {
	k.ac = ac
}
func (k *TestSetAdminConfigHandleKeeper) UpsertEvent(_ sdk.Context, _ Event) error {
	return nil
}

func (s *HandlerSetAdminConfigSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)

	na := GetRandomNodeAccount(NodeActive)
	keeper := &TestSetAdminConfigHandleKeeper{
		na: na,
		ac: AdminConfig{
			Key:     PoolRefundGasKey,
			Value:   "500",
			Address: na.NodeAddress,
		},
	}

	handler := NewSetAdminConfigHandler(keeper)

	ver := semver.MustParse("0.1.0")
	tx := GetRandomTx()
	msg := NewMsgSetAdminConfig(tx, PoolRefundGasKey, "1000", na.NodeAddress)
	result1 := handler.handle(ctx, msg, ver)
	c.Assert(result1.Code, Equals, sdk.CodeOK)
}
