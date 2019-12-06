package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type HandlerObservedTxInSuite struct{}

type TestObservedTxInValidateKeeper struct {
	KVStoreDummy
	isActive bool
}

func (k *TestObservedTxInValidateKeeper) IsActiveObserver(ctx sdk.Context, signer sdk.AccAddress) bool {
	return k.isActive
}

var _ = Suite(&HandlerObservedTxInSuite{})

func (s *HandlerObservedTxInSuite) TestValidate(c *C) {
	var err error
	ctx, _ := setupKeeperForTest(c)
	w := getHandlerTestWrapper(c, 1, true, false)

	keeper := &TestObservedTxInValidateKeeper{
		isActive: true,
	}

	handler := NewObservedTxInHandler(keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr)

	// happy path
	ver := semver.MustParse("0.1.0")
	pk := GetRandomPubKey()
	txs := ObservedTxs{NewObservedTx(GetRandomTx(), sdk.NewUint(12), pk)}
	txs[0].Tx.ToAddress, err = pk.GetAddress(txs[0].Tx.Coins[0].Asset.Chain)
	c.Assert(err, IsNil)
	msg := NewMsgObservedTxIn(txs, GetRandomBech32Addr())
	err = handler.Validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.Validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, badVersion)

	// inactive node account
	keeper.isActive = false
	msg = NewMsgObservedTxIn(txs, GetRandomBech32Addr())
	err = handler.Validate(ctx, msg, ver)
	c.Assert(err, Equals, notAuthorized)

	// invalid msg
	msg = MsgObservedTxIn{}
	err = handler.Validate(ctx, msg, ver)
	c.Assert(err, NotNil)
}

type TestObservedTxInHandleKeeper struct {
	KVStoreDummy
	activeNodeAccounts NodeAccounts
	voter              ObservedTxVoter
}

func (k TestObservedTxInHandleKeeper) ListActiveNodeAccounts(ctx sdk.Context) (NodeAccounts, error) {
	return k.activeNodeAccounts, nil
}

func (k TestObservedTxInHandleKeeper) GetObservedTxVoter(ctx sdk.Context, _ common.TxID) (ObservedTxVoter, error) {
	return k.voter, nil
}

func (s *HandlerObservedTxInSuite) TestHandle(c *C) {
	var err error
	ctx, _ := setupKeeperForTest(c)
	w := getHandlerTestWrapper(c, 1, true, false)

	ver := semver.MustParse("0.1.0")

	keeper := &TestObservedTxInHandleKeeper{}

	handler := NewObservedTxInHandler(keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr)

	pk := GetRandomPubKey()
	txs := ObservedTxs{NewObservedTx(GetRandomTx(), sdk.NewUint(12), pk)}
	txs[0].Tx.ToAddress, err = pk.GetAddress(txs[0].Tx.Coins[0].Asset.Chain)
	c.Assert(err, IsNil)
	msg := NewMsgObservedTxIn(txs, GetRandomBech32Addr())
	err = handler.Handle(ctx, msg, ver)
	c.Assert(err, IsNil)
}

/*
func (HandlerSuite) TestHandleMsgSetTxIn(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	err := w.keeper.SetPool(w.ctx, Pool{
		Asset:        common.BNBAsset,
		BalanceRune:  sdk.NewUint(100 * common.One),
		BalanceAsset: sdk.NewUint(100 * common.One),
	})
	c.Assert(err, IsNil)
	txIn := types.NewTxIn(
		common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One)),
			common.NewCoin(common.RuneAsset(), sdk.NewUint(100*common.One)),
		},
		"stake:BNB",
		GetRandomBNBAddress(),
		GetRandomBNBAddress(),
		common.BNBGasFeeSingleton,
		sdk.NewUint(1024),
		GetRandomPubKey())

	msgSetTxIn := types.NewMsgSetTxIn(
		[]TxInVoter{
			types.NewTxInVoter(GetRandomTxHash(), []TxIn{txIn}),
		},
		w.notActiveNodeAccount.NodeAddress)
	result := handleMsgSetTxIn(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, msgSetTxIn)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	w = getHandlerTestWrapper(c, 1, true, true)

	msgSetTxIn = types.NewMsgSetTxIn(
		[]TxInVoter{
			types.NewTxInVoter(GetRandomTxHash(), []TxIn{txIn}),
		},
		w.activeNodeAccount.NodeAddress)
	// send to wrong pool address, refund
	result1 := handleMsgSetTxIn(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, msgSetTxIn)
	c.Assert(result1.Code, Equals, sdk.CodeOK)
	c.Assert(w.txOutStore.blockOut, NotNil)
	c.Assert(w.txOutStore.blockOut.Valid(), IsNil)
	c.Assert(w.txOutStore.blockOut.IsEmpty(), Equals, false)
	c.Assert(len(w.txOutStore.blockOut.TxArray), Equals, 2)
	// expect to refund two coins
	c.Assert(w.txOutStore.GetOutboundItems(), HasLen, 2, Commentf("Len %d", len(w.txOutStore.GetOutboundItems())))

	currentChainPool := w.poolAddrMgr.currentPoolAddresses.Current.GetByChain(common.BNBChain)
	c.Assert(currentChainPool, NotNil)
	txIn1 := types.NewTxIn(
		common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One)),
			common.NewCoin(common.RuneAsset(), sdk.NewUint(100*common.One)),
		},
		"stake:BNB",
		GetRandomBNBAddress(),
		GetRandomBNBAddress(),
		common.BNBGasFeeSingleton,
		sdk.NewUint(1024),
		currentChainPool.PubKey)
	msgSetTxIn1 := types.NewMsgSetTxIn(
		[]TxInVoter{
			types.NewTxInVoter(GetRandomTxHash(), []TxIn{txIn1}),
		},
		w.activeNodeAccount.NodeAddress)
	result2 := handleMsgSetTxIn(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, msgSetTxIn1)
	c.Assert(result2.Code, Equals, sdk.CodeOK)
	p1, err := w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Assert(p1.BalanceRune.Uint64(), Equals, uint64(200*common.One))
	c.Assert(p1.BalanceAsset.Uint64(), Equals, uint64(200*common.One))
	// pool staker
	ps, err := w.keeper.GetPoolStaker(w.ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Assert(ps.TotalUnits.GT(sdk.ZeroUint()), Equals, true)
	chains, err := w.keeper.GetChains(w.ctx)
	c.Assert(err, IsNil)
	c.Check(chains.Has(common.BNBChain), Equals, true)
}
*/
