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
	pools             map[common.Asset]Pool
	activeNodeAccount NodeAccount
	event             []Event
	hasEvent          bool
}

func (k *TestSwapHandleKeeper) PoolExist(_ sdk.Context, asset common.Asset) bool {
	_, ok := k.pools[asset]
	return ok
}

func (k *TestSwapHandleKeeper) GetPool(_ sdk.Context, asset common.Asset) (Pool, error) {
	return k.pools[asset], nil
}

func (k *TestSwapHandleKeeper) SetPool(_ sdk.Context, pool Pool) error {
	k.pools[pool.Asset] = pool
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
func (k *TestSwapHandleKeeper) UpsertEvent(ctx sdk.Context, event Event) error {
	k.event = append(k.event, event)
	return nil
}
func (k *TestSwapHandleKeeper) clearEvent() {
	k.event = nil
}

func (s *HandlerSwapSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)
	keeper := &TestSwapHandleKeeper{
		pools:             make(map[common.Asset]Pool),
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
	keeper.clearEvent()
	msg := NewMsgSwap(tx, common.BNBAsset, signerBNBAddr, sdk.ZeroUint(), observerAddr)
	res := handler.handle(ctx, msg, ver, constAccessor)
	c.Assert(res.Code, Equals, CodeSwapFailPoolNotExist)
	c.Assert(keeper.event, IsNil)
	pool := NewPool()
	pool.Asset = common.BNBAsset
	pool.BalanceAsset = sdk.NewUint(100 * common.One)
	pool.BalanceRune = sdk.NewUint(100 * common.One)
	c.Assert(keeper.SetPool(ctx, pool), IsNil)
	keeper.clearEvent()
	// fund is not enough to pay for transaction fee
	res = handler.handle(ctx, msg, ver, constAccessor)
	c.Assert(res.IsOK(), Equals, false)
	c.Assert(keeper.event, IsNil)

	tx = common.NewTx(txID, signerBNBAddr, signerBNBAddr,
		common.Coins{
			common.NewCoin(common.RuneAsset(), sdk.NewUint(2*common.One)),
		},
		common.BNBGasFeeSingleton,
		"",
	)
	keeper.clearEvent()
	msgSwapPriceProtection := NewMsgSwap(tx, common.BNBAsset, signerBNBAddr, sdk.NewUint(2*common.One), observerAddr)
	res1 := handler.handle(ctx, msgSwapPriceProtection, ver, constAccessor)
	c.Assert(res1.IsOK(), Equals, false)
	c.Assert(res1.Code, Equals, CodeSwapFailTradeTarget)
	c.Assert(keeper.event, IsNil)

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
	keeper.clearEvent()
	c.Check(txOutStore.GetOutboundItems(), HasLen, 0)
	res2 := handler.handle(ctx, msgSwapFromTxIn.(MsgSwap), ver, constAccessor)
	c.Assert(res2.IsOK(), Equals, true)
	c.Assert(res2.Code, Equals, sdk.CodeOK)
	c.Assert(keeper.event, NotNil)
	c.Check(txOutStore.GetOutboundItems(), HasLen, 1)
}

func (s *HandlerSwapSuite) TestDoubleSwap(c *C) {
	ctx, _ := setupKeeperForTest(c)
	keeper := &TestSwapHandleKeeper{
		pools:             make(map[common.Asset]Pool),
		activeNodeAccount: GetRandomNodeAccount(NodeActive),
	}
	txOutStore := NewTxStoreDummy()
	handler := NewSwapHandler(keeper, txOutStore)
	ver := semver.MustParse("0.1.0")
	constAccessor := constants.GetConstantValues(ver)

	pool := NewPool()
	pool.Asset = common.BNBAsset
	pool.BalanceAsset = sdk.NewUint(100 * common.One)
	pool.BalanceRune = sdk.NewUint(100 * common.One)
	c.Assert(keeper.SetPool(ctx, pool), IsNil)

	poolTCAN := NewPool()
	tCanAsset, err := common.NewAsset("BNB.TCAN-014")
	c.Assert(err, IsNil)
	poolTCAN.Asset = tCanAsset
	poolTCAN.BalanceAsset = sdk.NewUint(334850000)
	poolTCAN.BalanceRune = sdk.NewUint(2349500000)
	c.Assert(keeper.SetPool(ctx, poolTCAN), IsNil)

	signerBNBAddr := GetRandomBNBAddress()
	observerAddr := keeper.activeNodeAccount.NodeAddress
	txOutStore.NewBlock(1, constAccessor)

	// double swap - happy path
	m, err := ParseMemo("swap:BNB:bnb18jtza8j86hfyuj2f90zec0g5gvjh823e5psn2u")
	txIn := NewObservedTx(
		common.NewTx(GetRandomTxHash(), signerBNBAddr, GetRandomBNBAddress(),
			common.Coins{
				common.NewCoin(tCanAsset, sdk.NewUint(20000000)),
			},
			common.BNBGasFeeSingleton,
			"swap:BNB:bnb18jtza8j86hfyuj2f90zec0g5gvjh823e5psn2u",
		),
		1,
		GetRandomPubKey(),
	)
	msgSwapFromTxIn, err := getMsgSwapFromMemo(m.(SwapMemo), txIn, observerAddr)
	c.Assert(err, IsNil)

	c.Check(txOutStore.GetOutboundItems(), HasLen, 0)
	res := handler.handle(ctx, msgSwapFromTxIn.(MsgSwap), ver, constAccessor)
	c.Assert(res.IsOK(), Equals, true)
	c.Assert(res.Code, Equals, sdk.CodeOK)
	c.Assert(keeper.event, NotNil)
	c.Assert(len(keeper.event), Equals, 2)
	c.Check(txOutStore.GetOutboundItems(), HasLen, 1)
	keeper.clearEvent()
	// double swap , RUNE not enough to pay for transaction fee

	m1, err := ParseMemo("swap:BNB:bnb18jtza8j86hfyuj2f90zec0g5gvjh823e5psn2u")
	txIn1 := NewObservedTx(
		common.NewTx(GetRandomTxHash(), signerBNBAddr, GetRandomBNBAddress(),
			common.Coins{
				common.NewCoin(tCanAsset, sdk.NewUint(10000000)),
			},
			common.BNBGasFeeSingleton,
			"swap:BNB:bnb18jtza8j86hfyuj2f90zec0g5gvjh823e5psn2u",
		),
		1,
		GetRandomPubKey(),
	)
	msgSwapFromTxIn1, err := getMsgSwapFromMemo(m1.(SwapMemo), txIn1, observerAddr)
	c.Assert(err, IsNil)
	txOutStore.ClearOutboundItems()
	res1 := handler.handle(ctx, msgSwapFromTxIn1.(MsgSwap), ver, constAccessor)
	c.Assert(res1.IsOK(), Equals, false)
	c.Assert(res1.Code, Equals, CodeSwapFailNotEnoughFee)
	c.Assert(keeper.event, IsNil)
	c.Check(txOutStore.GetOutboundItems(), HasLen, 0)
}
