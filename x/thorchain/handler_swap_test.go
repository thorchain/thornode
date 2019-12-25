package thorchain

import (
	"errors"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"

	. "gopkg.in/check.v1"
)

type HandlerSwapSuite struct{}

type TestSwapKeeper struct {
	KVStoreDummy
	activeNodeAccount NodeAccount
}

// IsActiveObserver see whether it is an active observer
func (k *TestSwapKeeper) IsActiveObserver(_ sdk.Context, addr sdk.AccAddress) bool {
	return k.activeNodeAccount.NodeAddress.Equals(addr)
}

var _ = Suite(&HandlerSwapSuite{})

func (s *HandlerSwapSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestSwapKeeper{
		activeNodeAccount: GetRandomNodeAccount(NodeActive),
	}

	txOutStore := NewTxStoreDummy()

	handler := NewSwapHandler(keeper, txOutStore)

	ver := semver.MustParse("0.1.0")
	txID := GetRandomTxHash()
	signerBNBAddr := GetRandomBNBAddress()
	observerAddr := keeper.activeNodeAccount.NodeAddress
	tx := common.NewTx(
		txID,
		signerBNBAddr,
		signerBNBAddr,
		common.Coins{
			common.NewCoin(common.RuneAsset(), sdk.OneUint()),
		},
		common.BNBGasFeeSingleton,
		"",
	)
	msg := NewMsgSwap(tx, common.BNBAsset, signerBNBAddr, sdk.ZeroUint(), observerAddr)
	err := handler.validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, badVersion)

	// invalid msg
	msg = MsgSwap{}
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, NotNil)

	// not signed observer
	msg = NewMsgSwap(tx, common.BNBAsset, signerBNBAddr, sdk.ZeroUint(), GetRandomBech32Addr())
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, Equals, notAuthorized)

}

type TestSwapHandleKeeper struct {
	KVStoreDummy
	currentPool       Pool
	activeNodeAccount NodeAccount
}

func (k *TestSwapHandleKeeper) PoolExist(_ sdk.Context, asset common.Asset) bool {
	return k.currentPool.Asset.Equals(asset)
}

func (k *TestSwapHandleKeeper) GetPool(_ sdk.Context, _ common.Asset) (Pool, error) {
	return k.currentPool, nil
}

func (k *TestSwapHandleKeeper) SetPool(_ sdk.Context, pool Pool) error {
	k.currentPool = pool
	return nil
}

// IsActiveObserver see whether it is an active observer
func (k *TestSwapHandleKeeper) IsActiveObserver(_ sdk.Context, addr sdk.AccAddress) bool {
	return k.activeNodeAccount.NodeAddress.Equals(addr)
}

func (k *TestSwapHandleKeeper) GetNodeAccount(_ sdk.Context, addr sdk.AccAddress) (NodeAccount, error) {
	if k.activeNodeAccount.NodeAddress.Equals(addr) {
		return k.activeNodeAccount, nil
	}
	return NodeAccount{}, errors.New("not exist")
}

func (k *TestSwapHandleKeeper) AddToLiquidityFees(_ sdk.Context, _ common.Asset, _ sdk.Uint) error {
	return nil
}
func (k *TestSwapHandleKeeper) UpsertEvent(ctx sdk.Context, event Event) error { return nil }

func (s *HandlerSwapSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestSwapHandleKeeper{
		activeNodeAccount: GetRandomNodeAccount(NodeActive),
	}

	txOutStore := NewTxStoreDummy()

	handler := NewSwapHandler(keeper, txOutStore)

	ver := semver.MustParse("0.1.0")
	constAccessor := constants.GetConstantValues(ver)
	txID := GetRandomTxHash()
	signerBNBAddr := GetRandomBNBAddress()
	observerAddr := keeper.activeNodeAccount.NodeAddress
	txOutStore.NewBlock(1, constAccessor)
	// no pool
	tx := common.NewTx(
		txID,
		signerBNBAddr,
		signerBNBAddr,
		common.Coins{
			common.NewCoin(common.RuneAsset(), sdk.OneUint()),
		},
		common.BNBGasFeeSingleton,
		"",
	)
	msg := NewMsgSwap(tx, common.BNBAsset, signerBNBAddr, sdk.ZeroUint(), observerAddr)
	res := handler.handle(ctx, msg, ver)
	c.Assert(res.Code, Equals, sdk.CodeInternal)
	pool := NewPool()
	pool.Asset = common.BNBAsset
	pool.BalanceAsset = sdk.NewUint(100 * common.One)
	pool.BalanceRune = sdk.NewUint(100 * common.One)
	c.Assert(keeper.SetPool(ctx, pool), IsNil)

	res = handler.handle(ctx, msg, ver)
	c.Assert(res.IsOK(), Equals, true)

	tx = common.NewTx(txID, signerBNBAddr, signerBNBAddr,
		common.Coins{
			common.NewCoin(common.RuneAsset(), sdk.OneUint()),
		},
		common.BNBGasFeeSingleton,
		"",
	)
	msgSwapPriceProtection := NewMsgSwap(tx, common.BNBAsset, signerBNBAddr, sdk.NewUint(2*common.One), observerAddr)
	res1 := handler.handle(ctx, msgSwapPriceProtection, ver)
	c.Assert(res1.IsOK(), Equals, false)
	c.Assert(res1.Code, Equals, sdk.CodeInternal)

	poolTCAN := NewPool()
	tCanAsset, err := common.NewAsset("BNB.TCAN-014")
	c.Assert(err, IsNil)
	poolTCAN.Asset = tCanAsset
	poolTCAN.BalanceAsset = sdk.NewUint(334850000)
	poolTCAN.BalanceRune = sdk.NewUint(2349500000)
	c.Assert(keeper.SetPool(ctx, poolTCAN), IsNil)

	m, err := ParseMemo("swap:RUNE-B1A:bnb18jtza8j86hfyuj2f90zec0g5gvjh823e5psn2u:124958592")
	txIn := NewObservedTx(
		common.NewTx(GetRandomTxHash(), signerBNBAddr, GetRandomBNBAddress(),
			common.Coins{
				common.NewCoin(tCanAsset, sdk.NewUint(20000000)),
			},
			common.BNBGasFeeSingleton,
			"swap:RUNE-B1A:bnb18jtza8j86hfyuj2f90zec0g5gvjh823e5psn2u:124958592",
		),
		1,
		GetRandomPubKey(),
	)
	msgSwapFromTxIn, err := getMsgSwapFromMemo(m.(SwapMemo), txIn, observerAddr)
	c.Assert(err, IsNil)

	c.Check(txOutStore.GetOutboundItems(), HasLen, 1)
	res2 := handler.handle(ctx, msgSwapFromTxIn.(MsgSwap), ver)
	c.Assert(res2.IsOK(), Equals, true)
	c.Assert(res2.Code, Equals, sdk.CodeOK)
	c.Check(txOutStore.GetOutboundItems(), HasLen, 2)
}
