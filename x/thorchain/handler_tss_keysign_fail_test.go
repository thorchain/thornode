package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type HandlerTssKeysignFailSuite struct{}

var _ = Suite(&HandlerTssKeysignFailSuite{})

type tssKeysignFailHandlerTestHelper struct {
	ctx           sdk.Context
	version       semver.Version
	keeper        *tssKeysignFailKeeperHelper
	constAccessor constants.ConstantValues
	nodeAccount   NodeAccount
	vaultManager  VersionedVaultManager
	members       common.PubKeys
	blame         common.Blame
}

type tssKeysignFailKeeperHelper struct {
	Keeper
	errListActiveAccounts           bool
	errGetTssVoter                  bool
	errFailToGetNodeAccountByPubKey bool
	errFailSetNodeAccount           bool
}

func newTssKeysignFailKeeperHelper(keeper Keeper) *tssKeysignFailKeeperHelper {
	return &tssKeysignFailKeeperHelper{
		Keeper: keeper,
	}
}

func (k *tssKeysignFailKeeperHelper) GetNodeAccountByPubKey(ctx sdk.Context, pk common.PubKey) (NodeAccount, error) {
	if k.errFailToGetNodeAccountByPubKey {
		return NodeAccount{}, kaboom
	}
	return k.Keeper.GetNodeAccountByPubKey(ctx, pk)
}

func (k *tssKeysignFailKeeperHelper) SetNodeAccount(ctx sdk.Context, na NodeAccount) error {
	if k.errFailSetNodeAccount {
		return kaboom
	}
	return k.Keeper.SetNodeAccount(ctx, na)
}

func (k *tssKeysignFailKeeperHelper) GetTssKeysignFailVoter(ctx sdk.Context, id string) (TssKeysignFailVoter, error) {
	if k.errGetTssVoter {
		return TssKeysignFailVoter{}, kaboom
	}
	return k.Keeper.GetTssKeysignFailVoter(ctx, id)
}

func (k *tssKeysignFailKeeperHelper) ListActiveNodeAccounts(ctx sdk.Context) (NodeAccounts, error) {
	if k.errListActiveAccounts {
		return NodeAccounts{}, kaboom
	}
	return k.Keeper.ListActiveNodeAccounts(ctx)
}

func newTssKeysignFailHandlerTestHelper(c *C) tssKeysignFailHandlerTestHelper {
	ctx, k := setupKeeperForTest(c)
	ctx = ctx.WithBlockHeight(1023)
	version := semver.MustParse("0.1.0")
	keeper := newTssKeysignFailKeeperHelper(k)
	// active account
	nodeAccount := GetRandomNodeAccount(NodeActive)
	nodeAccount.Bond = sdk.NewUint(100 * common.One)
	c.Assert(keeper.SetNodeAccount(ctx, nodeAccount), IsNil)
	constAccessor := constants.GetConstantValues(version)
	versionedTxOutStore := NewVersionedTxOutStore()
	vaultMgr := NewVersionedVaultMgr(versionedTxOutStore)
	var members common.PubKeys
	for i := 0; i < 8; i++ {
		na := GetRandomNodeAccount(NodeStandby)
		members = append(members, na.PubKeySet.Secp256k1)
		_ = keeper.SetNodeAccount(ctx, na)
	}
	blame := common.Blame{
		FailReason: "whatever",
		BlameNodes: members,
	}
	asgardVault := NewVault(ctx.BlockHeight(), ActiveVault, AsgardVault, GetRandomPubKey())
	c.Assert(keeper.SetVault(ctx, asgardVault), IsNil)
	return tssKeysignFailHandlerTestHelper{
		ctx:           ctx,
		version:       version,
		keeper:        keeper,
		constAccessor: constAccessor,
		nodeAccount:   nodeAccount,
		vaultManager:  vaultMgr,
		members:       members,
		blame:         blame,
	}
}

func (h HandlerTssKeysignFailSuite) TestTssKeysignFailHandler(c *C) {
	testCases := []struct {
		name           string
		messageCreator func(helper tssKeysignFailHandlerTestHelper) sdk.Msg
		runner         func(handler TssKeysignFailHandler, msg sdk.Msg, helper tssKeysignFailHandlerTestHelper) sdk.Result
		validator      func(helper tssKeysignFailHandlerTestHelper, msg sdk.Msg, result sdk.Result, c *C)
		expectedResult sdk.CodeType
	}{
		{
			name: "invalid message should return an error",
			messageCreator: func(helper tssKeysignFailHandlerTestHelper) sdk.Msg {
				return NewMsgNoOp(helper.nodeAccount.NodeAddress)
			},
			runner: func(handler TssKeysignFailHandler, msg sdk.Msg, helper tssKeysignFailHandlerTestHelper) sdk.Result {
				return handler.Run(helper.ctx, msg, helper.version, helper.constAccessor)
			},
			expectedResult: CodeInvalidMessage,
		},
		{
			name: "bad version should return an error",
			messageCreator: func(helper tssKeysignFailHandlerTestHelper) sdk.Msg {
				return NewMsgTssKeysignFail(helper.ctx.BlockHeight(), helper.blame, "hello", common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(100))}, helper.nodeAccount.NodeAddress)
			},
			runner: func(handler TssKeysignFailHandler, msg sdk.Msg, helper tssKeysignFailHandlerTestHelper) sdk.Result {
				return handler.Run(helper.ctx, msg, semver.MustParse("0.0.1"), helper.constAccessor)
			},
			expectedResult: CodeBadVersion,
		},
		{
			name: "Not signed by an active account should return an error",
			messageCreator: func(helper tssKeysignFailHandlerTestHelper) sdk.Msg {
				return NewMsgTssKeysignFail(helper.ctx.BlockHeight(), helper.blame, "hello", common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(100))}, GetRandomBech32Addr())
			},
			runner: func(handler TssKeysignFailHandler, msg sdk.Msg, helper tssKeysignFailHandlerTestHelper) sdk.Result {
				return handler.Run(helper.ctx, msg, semver.MustParse("0.1.0"), helper.constAccessor)
			},
			expectedResult: sdk.CodeUnauthorized,
		},
		{
			name: "empty signer should return an error",
			messageCreator: func(helper tssKeysignFailHandlerTestHelper) sdk.Msg {
				return NewMsgTssKeysignFail(helper.ctx.BlockHeight(), helper.blame, "hello", common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(100))}, sdk.AccAddress{})
			},
			runner: func(handler TssKeysignFailHandler, msg sdk.Msg, helper tssKeysignFailHandlerTestHelper) sdk.Result {
				return handler.Run(helper.ctx, msg, semver.MustParse("0.1.0"), helper.constAccessor)
			},
			expectedResult: sdk.CodeInvalidAddress,
		},
		{
			name: "empty id should return an error",
			messageCreator: func(helper tssKeysignFailHandlerTestHelper) sdk.Msg {
				tssMsg := NewMsgTssKeysignFail(helper.ctx.BlockHeight(), helper.blame, "hello", common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(100))}, helper.nodeAccount.NodeAddress)
				tssMsg.ID = ""
				return tssMsg
			},
			runner: func(handler TssKeysignFailHandler, msg sdk.Msg, helper tssKeysignFailHandlerTestHelper) sdk.Result {
				return handler.Run(helper.ctx, msg, semver.MustParse("0.1.0"), helper.constAccessor)
			},
			expectedResult: sdk.CodeUnknownRequest,
		},
		{
			name: "empty member pubkeys should return an error",
			messageCreator: func(helper tssKeysignFailHandlerTestHelper) sdk.Msg {
				return NewMsgTssKeysignFail(helper.ctx.BlockHeight(), common.Blame{
					FailReason: "",
					BlameNodes: common.PubKeys{},
				}, "hello", common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(100))}, helper.nodeAccount.NodeAddress)
			},
			runner: func(handler TssKeysignFailHandler, msg sdk.Msg, helper tssKeysignFailHandlerTestHelper) sdk.Result {
				return handler.Run(helper.ctx, msg, semver.MustParse("0.1.0"), helper.constAccessor)
			},
			expectedResult: sdk.CodeUnknownRequest,
		},
		{
			name: "normal blame should works fine",
			messageCreator: func(helper tssKeysignFailHandlerTestHelper) sdk.Msg {
				return NewMsgTssKeysignFail(helper.ctx.BlockHeight(), helper.blame, "hello", common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(100))}, helper.nodeAccount.NodeAddress)
			},
			runner: func(handler TssKeysignFailHandler, msg sdk.Msg, helper tssKeysignFailHandlerTestHelper) sdk.Result {
				return handler.Run(helper.ctx, msg, semver.MustParse("0.1.0"), helper.constAccessor)
			},
			expectedResult: sdk.CodeOK,
		},
		{
			name: "fail to list active node accounts should return an error",
			messageCreator: func(helper tssKeysignFailHandlerTestHelper) sdk.Msg {
				return NewMsgTssKeysignFail(helper.ctx.BlockHeight(), helper.blame, "hello", common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(100))}, helper.nodeAccount.NodeAddress)
			},
			runner: func(handler TssKeysignFailHandler, msg sdk.Msg, helper tssKeysignFailHandlerTestHelper) sdk.Result {
				helper.keeper.errListActiveAccounts = true
				return handler.Run(helper.ctx, msg, semver.MustParse("0.1.0"), helper.constAccessor)
			},
			expectedResult: sdk.CodeInternal,
		},
		{
			name: "fail to get Tss Keysign fail voter should return an error",
			messageCreator: func(helper tssKeysignFailHandlerTestHelper) sdk.Msg {
				return NewMsgTssKeysignFail(helper.ctx.BlockHeight(), helper.blame, "hello", common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(100))}, helper.nodeAccount.NodeAddress)
			},
			runner: func(handler TssKeysignFailHandler, msg sdk.Msg, helper tssKeysignFailHandlerTestHelper) sdk.Result {
				helper.keeper.errGetTssVoter = true
				return handler.Run(helper.ctx, msg, semver.MustParse("0.1.0"), helper.constAccessor)
			},
			expectedResult: sdk.CodeInternal,
		},
		{
			name: "fail to get node account should return an error",
			messageCreator: func(helper tssKeysignFailHandlerTestHelper) sdk.Msg {
				return NewMsgTssKeysignFail(helper.ctx.BlockHeight(), helper.blame, "hello", common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(100))}, helper.nodeAccount.NodeAddress)
			},
			runner: func(handler TssKeysignFailHandler, msg sdk.Msg, helper tssKeysignFailHandlerTestHelper) sdk.Result {
				helper.keeper.errFailToGetNodeAccountByPubKey = true
				return handler.Run(helper.ctx, msg, semver.MustParse("0.1.0"), helper.constAccessor)
			},
			expectedResult: sdk.CodeInternal,
		},
		{
			name: "fail to set node account should return an error",
			messageCreator: func(helper tssKeysignFailHandlerTestHelper) sdk.Msg {
				return NewMsgTssKeysignFail(helper.ctx.BlockHeight(), helper.blame, "hello", common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(100))}, helper.nodeAccount.NodeAddress)
			},
			runner: func(handler TssKeysignFailHandler, msg sdk.Msg, helper tssKeysignFailHandlerTestHelper) sdk.Result {
				helper.keeper.errFailSetNodeAccount = true
				return handler.Run(helper.ctx, msg, semver.MustParse("0.1.0"), helper.constAccessor)
			},
			expectedResult: sdk.CodeInternal,
		},
	}
	for _, tc := range testCases {
		helper := newTssKeysignFailHandlerTestHelper(c)
		handler := NewTssKeysignFailHandler(helper.keeper)
		msg := tc.messageCreator(helper)
		result := tc.runner(handler, msg, helper)
		c.Assert(result.Code, Equals, tc.expectedResult, Commentf("name:%s", tc.name))
		if tc.validator != nil {
			tc.validator(helper, msg, result, c)
		}
	}
}
