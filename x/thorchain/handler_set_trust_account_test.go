package thorchain

import (
	"errors"
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type HandlerSetTrustAccountSuite struct{}

type TestSetTrustAccountKeeper struct {
	KVStoreDummy
	na NodeAccount
}

func (k *TestSetTrustAccountKeeper) GetNodeAccount(ctx sdk.Context, signer sdk.AccAddress) (NodeAccount, error) {
	return k.na, nil
}

var _ = Suite(&HandlerSetTrustAccountSuite{})

func (s *HandlerSetTrustAccountSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestSetTrustAccountKeeper{
		na: GetRandomNodeAccount(NodeActive),
	}

	handler := NewSetTrustAccountHandler(keeper)

	// happy path
	ver := semver.MustParse("0.1.0")
	signer := GetRandomBech32Addr()
	c.Assert(signer.Empty(), Equals, false)
	consensPubKey := GetRandomBech32ConsensusPubKey()
	pubKeys := GetRandomPubkeys()

	msg := NewMsgSetTrustAccount(pubKeys, consensPubKey, signer)
	err := handler.validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, badVersion)

	// invalid msg
	msg = MsgSetTrustAccount{}
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, NotNil)
}

type TestSetTrustAccountHandleKeeper struct {
	KVStoreDummy
	na NodeAccount
}

func (k *TestSetTrustAccountHandleKeeper) GetNodeAccount(ctx sdk.Context, signer sdk.AccAddress) (NodeAccount, error) {
	return k.na, nil
}

func (k *TestSetTrustAccountHandleKeeper) SetNodeAccount(_ sdk.Context, na NodeAccount) error {
	k.na = na
	return nil
}

func (k *TestSetTrustAccountHandleKeeper) EnsureTrustAccountUnique(_ sdk.Context, consensPubKey string, pubKeys common.PubKeys) error {
	return nil
}

func (s *HandlerSetTrustAccountSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestSetTrustAccountHandleKeeper{
		na: GetRandomNodeAccount(NodeActive),
	}

	handler := NewSetTrustAccountHandler(keeper)

	ver := semver.MustParse("0.1.0")

	ctx = ctx.WithBlockHeight(1)
	signer := GetRandomBech32Addr()

	// add observer
	bepConsPubKey := GetRandomBech32ConsensusPubKey()
	bondAddr := GetRandomBNBAddress()
	pubKeys := GetRandomPubkeys()
	emptyPubKeys := common.PubKeys{}

	msgTrustAccount := NewMsgSetTrustAccount(pubKeys, bepConsPubKey, signer)

	bond := sdk.NewUint(common.One * 100)
	nodeAccount := NewNodeAccount(signer, NodeActive, emptyPubKeys, "", bond, bondAddr, ctx.BlockHeight())
	c.Assert(keeper.SetNodeAccount(ctx, nodeAccount), IsNil)

	activeFailResult := handler.handle(ctx, msgTrustAccount, ver)
	c.Check(activeFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(activeFailResult.IsOK(), Equals, false)

	nodeAccount = NewNodeAccount(signer, NodeDisabled, emptyPubKeys, "", bond, bondAddr, ctx.BlockHeight())
	c.Assert(keeper.SetNodeAccount(ctx, nodeAccount), IsNil)

	disabledFailResult := handler.handle(ctx, msgTrustAccount, ver)
	c.Check(disabledFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(disabledFailResult.IsOK(), Equals, false)

	nodeAccount = NewNodeAccount(signer, NodeWhiteListed, emptyPubKeys, "", bond, bondAddr, ctx.BlockHeight())
	c.Assert(keeper.SetNodeAccount(ctx, nodeAccount), IsNil)

	success := handler.handle(ctx, msgTrustAccount, ver)
	c.Check(success.Code, Equals, sdk.CodeOK)
	c.Check(success.IsOK(), Equals, true)
	c.Assert(keeper.na.NodePubKey, Equals, emptyPubKeys)
	c.Assert(keeper.na.ValidatorConsPubKey, Equals, bepConsPubKey)
}

type TestSetTrustAccountHandleFailUniqueKeeper struct {
	KVStoreDummy
	na NodeAccount
}

func (k *TestSetTrustAccountHandleFailUniqueKeeper) GetNodeAccount(ctx sdk.Context, signer sdk.AccAddress) (NodeAccount, error) {
	return k.na, nil
}

func (k *TestSetTrustAccountHandleFailUniqueKeeper) SetNodeAccount(_ sdk.Context, na NodeAccount) error {
	k.na = na
	return nil
}

func (k *TestSetTrustAccountHandleFailUniqueKeeper) EnsureTrustAccountUnique(_ sdk.Context, consensPubKey string, pubKeys common.PubKeys) error {
	return errors.New("not unique")
}

func (s *HandlerSetTrustAccountSuite) TestHandleFailUnique(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestSetTrustAccountHandleFailUniqueKeeper{
		na: GetRandomNodeAccount(NodeActive),
	}

	handler := NewSetTrustAccountHandler(keeper)

	ver := semver.MustParse("0.1.0")

	ctx = ctx.WithBlockHeight(1)
	signer := GetRandomBech32Addr()

	// add observer
	bepConsPubKey := GetRandomBech32ConsensusPubKey()
	pubKeys := GetRandomPubkeys()

	msgTrustAccount := NewMsgSetTrustAccount(pubKeys, bepConsPubKey, signer)
	notUniqueFailResult := handler.handle(ctx, msgTrustAccount, ver)
	c.Check(notUniqueFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(notUniqueFailResult.IsOK(), Equals, false)
}
