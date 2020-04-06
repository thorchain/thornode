package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
	. "gopkg.in/check.v1"
)

type HandlerSetNodeKeysSuite struct{}

type TestSetNodeKeysKeeper struct {
	KVStoreDummy
	na     NodeAccount
	ensure error
}

func (k *TestSetNodeKeysKeeper) GetNodeAccount(ctx sdk.Context, signer sdk.AccAddress) (NodeAccount, error) {
	return k.na, nil
}

func (k *TestSetNodeKeysKeeper) EnsureNodeKeysUnique(_ sdk.Context, _ string, _ common.PubKeySet) error {
	return k.ensure
}

var _ = Suite(&HandlerSetNodeKeysSuite{})

func (s *HandlerSetNodeKeysSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestSetNodeKeysKeeper{
		na:     GetRandomNodeAccount(NodeStandby),
		ensure: nil,
	}

	handler := NewSetNodeKeysHandler(keeper)

	// happy path
	ver := constants.SWVersion
	signer := GetRandomBech32Addr()
	c.Assert(signer.Empty(), Equals, false)
	consensPubKey := GetRandomBech32ConsensusPubKey()
	pubKeys := GetRandomPubKeySet()

	msg := NewMsgSetNodeKeys(pubKeys, consensPubKey, signer)
	err := handler.validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// cannot set node keys for active account
	keeper.na.Status = NodeActive
	msg = NewMsgSetNodeKeys(pubKeys, consensPubKey, keeper.na.NodeAddress)
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, NotNil)

	// cannot set node keys for disabled account
	keeper.na.Status = NodeDisabled
	msg = NewMsgSetNodeKeys(pubKeys, consensPubKey, keeper.na.NodeAddress)
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, NotNil)

	// cannot set node keys when duplicate
	keeper.na.Status = NodeStandby
	keeper.ensure = fmt.Errorf("duplicate keys")
	msg = NewMsgSetNodeKeys(keeper.na.PubKeySet, consensPubKey, keeper.na.NodeAddress)
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, ErrorMatches, "duplicate keys")
	keeper.ensure = nil

	// new version GT
	err = handler.validate(ctx, msg, semver.MustParse("2.0.0"))
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, errInvalidVersion)

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

func (k *TestSetNodeKeysHandleKeeper) EnsureNodeKeysUnique(_ sdk.Context, consensPubKey string, pubKeys common.PubKeySet) error {
	return nil
}

func (s *HandlerSetNodeKeysSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestSetNodeKeysHandleKeeper{
		na: GetRandomNodeAccount(NodeActive),
	}

	handler := NewSetNodeKeysHandler(keeper)

	ver := constants.SWVersion

	constAccessor := constants.GetConstantValues(ver)
	ctx = ctx.WithBlockHeight(1)
	signer := GetRandomBech32Addr()

	// add observer
	bepConsPubKey := GetRandomBech32ConsensusPubKey()
	bondAddr := GetRandomBNBAddress()
	pubKeys := GetRandomPubKeySet()
	emptyPubKeySet := common.PubKeySet{}

	msgNodeKeys := NewMsgSetNodeKeys(pubKeys, bepConsPubKey, signer)

	bond := sdk.NewUint(common.One * 100)
	nodeAccount := NewNodeAccount(signer, NodeActive, emptyPubKeySet, "", bond, bondAddr, ctx.BlockHeight())
	c.Assert(keeper.SetNodeAccount(ctx, nodeAccount), IsNil)

	nodeAccount = NewNodeAccount(signer, NodeWhiteListed, emptyPubKeySet, "", bond, bondAddr, ctx.BlockHeight())
	c.Assert(keeper.SetNodeAccount(ctx, nodeAccount), IsNil)

	// happy path
	success := handler.handle(ctx, msgNodeKeys, ver, constAccessor)
	c.Check(success.Code, Equals, sdk.CodeOK)
	c.Check(success.IsOK(), Equals, true)
	c.Assert(keeper.na.PubKeySet, Equals, pubKeys)
	c.Assert(keeper.na.ValidatorConsPubKey, Equals, bepConsPubKey)
	c.Assert(keeper.na.Status, Equals, NodeStandby)
	c.Assert(keeper.na.StatusSince, Equals, int64(1))

	// update version
	success2 := handler.handle(ctx, msgNodeKeys, semver.MustParse("2.0.0"), constAccessor)
	c.Check(success2.Code, Equals, sdk.CodeOK)
	c.Check(success2.IsOK(), Equals, true)
	c.Check(keeper.na.Version.String(), Equals, "2.0.0")
}
