package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type HandlerYggdrasilSuite struct{}

type TestYggdrasilValidateKeeper struct {
	KVStoreDummy
	na NodeAccount
}

func (k *TestYggdrasilValidateKeeper) GetNodeAccount(ctx sdk.Context, signer sdk.AccAddress) (NodeAccount, error) {
	return k.na, nil
}

var _ = Suite(&HandlerYggdrasilSuite{})

func (s *HandlerYggdrasilSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestYggdrasilValidateKeeper{
		na: GetRandomNodeAccount(NodeActive),
	}

	poolAddrMgr := NewPoolAddressMgr(keeper)
	validatorMgr := NewValidatorMgr(keeper, poolAddrMgr)
	txOutStore := NewTxStoreDummy()

	handler := NewYggdrasilHandler(keeper, txOutStore, poolAddrMgr, validatorMgr)

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
	ygg  Yggdrasil
	na   NodeAccount
	pool Pool
}

func (k *TestYggdrasilHandleKeeper) GetYggdrasil(ctx sdk.Context, pubKey common.PubKey) (Yggdrasil, error) {
	return k.ygg, nil
}

func (k *TestYggdrasilHandleKeeper) GetNodeAccountByPubKey(ctx sdk.Context, pubKey common.PubKey) (NodeAccount, error) {
	return k.na, nil
}

func (k *TestYggdrasilHandleKeeper) SetYggdrasil(ctx sdk.Context, ygg Yggdrasil) error {
	k.ygg = ygg
	return nil
}

func (k *TestYggdrasilHandleKeeper) GetPool(_ sdk.Context, _ common.Asset) (Pool, error) {
	return k.pool, nil
}

func (k *TestYggdrasilHandleKeeper) SetNodeAccount(_ sdk.Context, na NodeAccount) error {
	k.na = na
	return nil
}

func (s *HandlerYggdrasilSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)

	pubKey := GetRandomPubKey()
	keeper := &TestYggdrasilHandleKeeper{
		ygg: Yggdrasil{
			PubKey: pubKey,
			Coins: common.Coins{
				common.NewCoin(common.RuneAsset(), sdk.NewUint(1022*common.One)),
				common.NewCoin(common.BNBAsset, sdk.NewUint(33*common.One)),
			},
		},
		na: GetRandomNodeAccount(NodeActive),
		pool: Pool{
			Asset:        common.BNBAsset,
			BalanceRune:  sdk.NewUint(234 * common.One),
			BalanceAsset: sdk.NewUint(765 * common.One),
		},
	}

	poolAddrMgr := NewPoolAddressMgr(keeper)
	validatorMgr := NewValidatorMgr(keeper, poolAddrMgr)
	validatorMgr.BeginBlock(ctx)
	txOutStore := NewTxStoreDummy()

	handler := NewYggdrasilHandler(keeper, txOutStore, poolAddrMgr, validatorMgr)

	// check yggdrasil balance on add funds
	ver := semver.MustParse("0.1.0")
	coins := common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One))}
	txID := GetRandomTxHash()
	signer := GetRandomBech32Addr()
	msg := NewMsgYggdrasil(pubKey, true, coins, txID, signer)
	result := handler.handle(ctx, msg, ver)
	c.Assert(result.Code, Equals, sdk.CodeOK)

	ygg, err := keeper.GetYggdrasil(ctx, pubKey)
	c.Assert(err, IsNil)
	coin := ygg.GetCoin(common.BNBAsset)
	c.Check(coin.Amount.Uint64(), Equals, sdk.NewUint(133*common.One).Uint64(), Commentf("%d vs %d", coin.Amount.Uint64(), sdk.NewUint(133*common.One).Uint64()))

	// check yggdrasil balance on sub funds
	msg = NewMsgYggdrasil(pubKey, false, coins, txID, signer)
	result = handler.handle(ctx, msg, ver)
	c.Assert(result.Code, Equals, sdk.CodeOK)

	ygg, err = keeper.GetYggdrasil(ctx, pubKey)
	c.Assert(err, IsNil)
	coin = ygg.GetCoin(common.BNBAsset)
	c.Check(coin.Amount.Uint64(), Equals, sdk.NewUint(33*common.One).Uint64(), Commentf("%d vs %d", coin.Amount.Uint64(), sdk.NewUint(33*common.One).Uint64()))
}

// TODO test handleRagnarokProtocolStep2
