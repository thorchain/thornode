package thorchain

import (
	"errors"
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
	. "gopkg.in/check.v1"
)

type HandlerSetNodeKeysSuite struct{}

type TestSetNodeKeysKeeper struct {
	KVStoreDummy
	na NodeAccount
}

func (k *TestSetNodeKeysKeeper) GetNodeAccount(ctx sdk.Context, signer sdk.AccAddress) (NodeAccount, error) {
	return k.na, nil
}

var _ = Suite(&HandlerSetNodeKeysSuite{})

func (s *HandlerSetNodeKeysSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestSetNodeKeysKeeper{
		na: GetRandomNodeAccount(NodeActive),
	}

	handler := NewSetNodeKeysHandler(keeper)

	// happy path
	ver := semver.MustParse("0.1.0")
	signer := GetRandomBech32Addr()
	c.Assert(signer.Empty(), Equals, false)
	consensPubKey := GetRandomBech32ConsensusPubKey()
	pubKeys := GetRandomPubkeys()

	msg := NewMsgSetNodeKeys(pubKeys, consensPubKey, signer)
	err := handler.validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// new version GT
	err = handler.validate(ctx, msg, semver.MustParse("2.0.0"))
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, badVersion)

	// invalid msg
	msg = MsgSetNodeKeys{}
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, NotNil)
}

type TestSetNodeKeysHandleKeeper struct {
	KVStoreDummy
	na NodeAccount
}

func (k *TestSetNodeKeysHandleKeeper) GetNodeAccount(ctx sdk.Context, signer sdk.AccAddress) (NodeAccount, error) {
	return k.na, nil
}

func (k *TestSetNodeKeysHandleKeeper) SetNodeAccount(_ sdk.Context, na NodeAccount) error {
	k.na = na
	return nil
}

func (k *TestSetNodeKeysHandleKeeper) EnsureNodeKeysUnique(_ sdk.Context, consensPubKey string, pubKeys common.PubKeys) error {
	return nil
}

func (s *HandlerSetNodeKeysSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestSetNodeKeysHandleKeeper{
		na: GetRandomNodeAccount(NodeActive),
	}

	handler := NewSetNodeKeysHandler(keeper)

	ver := semver.MustParse("0.1.0")

	constAccessor := constants.GetConstantValues(ver)
	ctx = ctx.WithBlockHeight(1)
	signer := GetRandomBech32Addr()

	// add observer
	bepConsPubKey := GetRandomBech32ConsensusPubKey()
	bondAddr := GetRandomBNBAddress()
	pubKeys := GetRandomPubkeys()
	emptyPubKeys := common.PubKeys{}

	msgNodeKeys := NewMsgSetNodeKeys(pubKeys, bepConsPubKey, signer)

	bond := sdk.NewUint(common.One * 100)
	nodeAccount := NewNodeAccount(signer, NodeActive, emptyPubKeys, "", bond, bondAddr, ctx.BlockHeight())
	c.Assert(keeper.SetNodeAccount(ctx, nodeAccount), IsNil)

	activeFailResult := handler.handle(ctx, msgNodeKeys, ver, constAccessor)
	c.Check(activeFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(activeFailResult.IsOK(), Equals, false)

	nodeAccount = NewNodeAccount(signer, NodeDisabled, emptyPubKeys, "", bond, bondAddr, ctx.BlockHeight())
	c.Assert(keeper.SetNodeAccount(ctx, nodeAccount), IsNil)

	disabledFailResult := handler.handle(ctx, msgNodeKeys, ver, constAccessor)
	c.Check(disabledFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(disabledFailResult.IsOK(), Equals, false)

	nodeAccount = NewNodeAccount(signer, NodeWhiteListed, emptyPubKeys, "", bond, bondAddr, ctx.BlockHeight())
	c.Assert(keeper.SetNodeAccount(ctx, nodeAccount), IsNil)

	// happy path
	success := handler.handle(ctx, msgNodeKeys, ver, constAccessor)
	c.Check(success.Code, Equals, sdk.CodeOK)
	c.Check(success.IsOK(), Equals, true)
	c.Assert(keeper.na.NodePubKey, Equals, pubKeys)
	c.Assert(keeper.na.ValidatorConsPubKey, Equals, bepConsPubKey)
	c.Assert(keeper.na.Status, Equals, NodeStandby)
	c.Assert(keeper.na.StatusSince, Equals, int64(1))

	// update version
	success2 := handler.handle(ctx, msgNodeKeys, semver.MustParse("2.0.0"), constAccessor)
	c.Check(success2.Code, Equals, sdk.CodeOK)
	c.Check(success2.IsOK(), Equals, true)
	c.Check(keeper.na.Version.String(), Equals, "2.0.0")
}

type TestSetNodeKeysHandleFailUniqueKeeper struct {
	KVStoreDummy
	na NodeAccount
}

func (k *TestSetNodeKeysHandleFailUniqueKeeper) GetNodeAccount(ctx sdk.Context, signer sdk.AccAddress) (NodeAccount, error) {
	return k.na, nil
}

func (k *TestSetNodeKeysHandleFailUniqueKeeper) SetNodeAccount(_ sdk.Context, na NodeAccount) error {
	k.na = na
	return nil
}

func (k *TestSetNodeKeysHandleFailUniqueKeeper) EnsureNodeKeysUnique(_ sdk.Context, consensPubKey string, pubKeys common.PubKeys) error {
	return errors.New("not unique")
}

func (s *HandlerSetNodeKeysSuite) TestHandleFailUnique(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestSetNodeKeysHandleFailUniqueKeeper{
		na: GetRandomNodeAccount(NodeActive),
	}

	handler := NewSetNodeKeysHandler(keeper)

	ver := semver.MustParse("0.1.0")

	constAccessor := constants.GetConstantValues(ver)
	ctx = ctx.WithBlockHeight(1)
	signer := GetRandomBech32Addr()

	// add observer
	bepConsPubKey := GetRandomBech32ConsensusPubKey()
	pubKeys := GetRandomPubkeys()

	msgNodeKeys := NewMsgSetNodeKeys(pubKeys, bepConsPubKey, signer)
	notUniqueFailResult := handler.handle(ctx, msgNodeKeys, ver, constAccessor)
	c.Check(notUniqueFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(notUniqueFailResult.IsOK(), Equals, false)
}
