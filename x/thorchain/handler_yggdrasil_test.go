package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"

	. "gopkg.in/check.v1"
)

type HandlerYggdrasilSuite struct{}

type TestYggdrasilValidateKeeper struct {
	KVStoreDummy
	na NodeAccount
}

func (k *TestYggdrasilValidateKeeper) GetNodeAccount(_ sdk.Context, signer sdk.AccAddress) (NodeAccount, error) {
	return k.na, nil
}

var _ = Suite(&HandlerYggdrasilSuite{})

func (s *HandlerYggdrasilSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestYggdrasilValidateKeeper{
		na: GetRandomNodeAccount(NodeActive),
	}

	vaultMgr := NewVaultMgrDummy()
	validatorMgr := NewValidatorMgr(keeper, vaultMgr)
	txOutStore := NewTxStoreDummy()

	handler := NewYggdrasilHandler(keeper, txOutStore, validatorMgr)

	// happy path
	ver := semver.MustParse("0.1.0")
	pubKey := GetRandomPubKey()
	coins := common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One))}
	txID := GetRandomTxHash()
	signer := GetRandomBech32Addr()
	msg := NewMsgYggdrasil(pubKey, true, coins, txID, signer)
	err := handler.validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, badVersion)

	// invalid msg
	msg = MsgYggdrasil{}
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, NotNil)
}

type TestYggdrasilHandleKeeper struct {
	KVStoreDummy
	ygg  Vault
	na   NodeAccount
	pool Pool
}

func (k *TestYggdrasilHandleKeeper) GetVault(ctx sdk.Context, pubKey common.PubKey) (Vault, error) {
	return k.ygg, nil
}

func (k *TestYggdrasilHandleKeeper) GetNodeAccountByPubKey(ctx sdk.Context, pubKey common.PubKey) (NodeAccount, error) {
	return k.na, nil
}

func (k *TestYggdrasilHandleKeeper) SetVault(ctx sdk.Context, ygg Vault) error {
	k.ygg = ygg
	return nil
}

func (k *TestYggdrasilHandleKeeper) HasValidVaultPools(_ sdk.Context) (bool, error) {
	return true, nil
}

func (k *TestYggdrasilHandleKeeper) GetPool(_ sdk.Context, _ common.Asset) (Pool, error) {
	return k.pool, nil
}

func (k *TestYggdrasilHandleKeeper) TotalActiveNodeAccount(_ sdk.Context) (int, error) {
	return 1, nil
}

func (k *TestYggdrasilHandleKeeper) SetNodeAccount(_ sdk.Context, na NodeAccount) error {
	k.na = na
	return nil
}

func (s *HandlerYggdrasilSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)
	ctx = ctx.WithBlockHeight(12)

	pubKey := GetRandomPubKey()
	ygg := NewVault(ctx.BlockHeight(), ActiveVault, YggdrasilVault, pubKey)
	ygg.Coins = common.Coins{
		common.NewCoin(common.RuneAsset(), sdk.NewUint(1022*common.One)),
		common.NewCoin(common.BNBAsset, sdk.NewUint(33*common.One)),
	}
	keeper := &TestYggdrasilHandleKeeper{
		ygg: ygg,
		na:  GetRandomNodeAccount(NodeActive),
		pool: Pool{
			Asset:        common.BNBAsset,
			BalanceRune:  sdk.NewUint(234 * common.One),
			BalanceAsset: sdk.NewUint(765 * common.One),
		},
	}
	ver := semver.MustParse("0.1.0")
	constAccessor := constants.GetConstantValues(ver)
	vaultMgr := NewVaultMgrDummy()
	validatorMgr := NewValidatorMgr(keeper, vaultMgr)
	validatorMgr.BeginBlock(ctx, constAccessor)
	txOutStore := NewTxStoreDummy()

	handler := NewYggdrasilHandler(keeper, txOutStore, validatorMgr)

	// check yggdrasil balance on add funds
	coins := common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One))}
	txID := GetRandomTxHash()
	signer := GetRandomBech32Addr()
	msg := NewMsgYggdrasil(pubKey, true, coins, txID, signer)
	result := handler.handle(ctx, msg, ver, constAccessor)
	c.Assert(result.Code, Equals, sdk.CodeOK, Commentf("%+v\n", result))

	ygg, err := keeper.GetVault(ctx, pubKey)
	c.Assert(err, IsNil)
	coin := ygg.GetCoin(common.BNBAsset)
	c.Check(coin.Amount.Uint64(), Equals, sdk.NewUint(133*common.One).Uint64(), Commentf("%d vs %d", coin.Amount.Uint64(), sdk.NewUint(133*common.One).Uint64()))

	// check yggdrasil balance on sub funds
	msg = NewMsgYggdrasil(pubKey, false, coins, txID, signer)
	result = handler.handle(ctx, msg, ver, constAccessor)
	c.Assert(result.Code, Equals, sdk.CodeOK)

	ygg, err = keeper.GetVault(ctx, pubKey)
	c.Assert(err, IsNil)
	coin = ygg.GetCoin(common.BNBAsset)
	c.Check(coin.Amount.Uint64(), Equals, sdk.NewUint(33*common.One).Uint64(), Commentf("%d vs %d", coin.Amount.Uint64(), sdk.NewUint(33*common.One).Uint64()))
}

// TODO test handleRagnarokProtocolStep2
