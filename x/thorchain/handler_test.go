package thorchain

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/common"

	"gitlab.com/thorchain/bepswap/thornode/x/thorchain/types"
)

type HandlerSuite struct{}

var _ = Suite(&HandlerSuite{})

func (s *HandlerSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

// nolint: deadcode unused
// create a codec used only for testing
func makeTestCodec() *codec.Codec {
	var cdc = codec.New()
	bank.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	RegisterCodec(cdc)
	supply.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}

var (
	multiPerm    = "multiple permissions account"
	randomPerm   = "random permission"
	holder       = "holder"
	keyThorchain = sdk.NewKVStoreKey(StoreKey)
)

func setupKeeperForTest(c *C) (sdk.Context, Keeper) {
	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	keySupply := sdk.NewKVStoreKey(supply.StoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keySupply, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyThorchain, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	err := ms.LoadLatestVersion()
	c.Assert(err, IsNil)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "thorchain"}, false, log.NewNopLogger())
	cdc := makeTestCodec()

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)
	ak := auth.NewAccountKeeper(cdc, keyAcc, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	bk := bank.NewBaseKeeper(ak, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace, nil)

	maccPerms := map[string][]string{
		auth.FeeCollectorName: nil,
		holder:                nil,
		supply.Minter:         {supply.Minter},
		supply.Burner:         {supply.Burner},
		multiPerm:             {supply.Minter, supply.Burner, supply.Staking},
		randomPerm:            {"random"},
		ModuleName:            {supply.Minter},
	}
	supplyKeeper := supply.NewKeeper(cdc, keySupply, ak, bk, maccPerms)
	totalSupply := sdk.NewCoins(sdk.NewCoin("bep", sdk.NewInt(1000*common.One)))
	supplyKeeper.SetSupply(ctx, supply.NewSupply(totalSupply))
	k := NewKeeper(bk, supplyKeeper, keyThorchain, cdc)
	return ctx, k
}

type handlerTestWrapper struct {
	ctx                  sdk.Context
	keeper               Keeper
	poolAddrMgr          *PoolAddressManager
	validatorMgr         *ValidatorManager
	txOutStore           *TxOutStore
	activeNodeAccount    NodeAccount
	notActiveNodeAccount NodeAccount
}

func getHandlerTestWrapper(c *C, height int64, withActiveNode, withActieBNBPool bool) *handlerTestWrapper {
	ctx, k := setupKeeperForTest(c)
	ctx = ctx.WithBlockHeight(height)
	acc1 := GetRandomNodeAccount(NodeActive)
	if withActiveNode {
		k.SetNodeAccount(ctx, acc1)
	}
	if withActieBNBPool {
		p := k.GetPool(ctx, common.BNBAsset)
		p.Asset = common.BNBAsset
		p.Status = PoolEnabled
		p.BalanceRune = sdk.NewUint(100 * common.One)
		p.BalanceAsset = sdk.NewUint(100 * common.One)
		k.SetPool(ctx, p)
	}
	genesisPoolPubKey := common.NewPoolPubKey(common.BNBChain, 0, GetRandomPubKey())
	genesisPoolAddress := NewPoolAddresses(common.EmptyPoolPubKeys, common.PoolPubKeys{
		genesisPoolPubKey,
	}, common.EmptyPoolPubKeys, 100, 90)
	k.SetPoolAddresses(ctx, genesisPoolAddress)
	poolAddrMgr := NewPoolAddressManager(k)
	validatorMgr := NewValidatorManager(k)
	poolAddrMgr.BeginBlock(ctx)
	validatorMgr.BeginBlock(ctx, height)
	txOutStore := NewTxOutStore(k, poolAddrMgr)
	txOutStore.NewBlock(uint64(height))

	return &handlerTestWrapper{
		ctx:                  ctx,
		keeper:               k,
		poolAddrMgr:          poolAddrMgr,
		validatorMgr:         validatorMgr,
		txOutStore:           txOutStore,
		activeNodeAccount:    acc1,
		notActiveNodeAccount: GetRandomNodeAccount(NodeDisabled),
	}
}

func (HandlerSuite) TestHandleMsgApply(c *C) {

	w := getHandlerTestWrapper(c, 1, false, false)
	bond := sdk.NewUint(100)
	bondAddr := GetRandomBNBAddress()
	// Not Authorized
	msgApply := NewMsgBond(w.activeNodeAccount.NodeAddress, bond, GetRandomTxHash(), bondAddr, w.activeNodeAccount.NodeAddress)
	c.Assert(msgApply.ValidateBasic(), IsNil)
	result := handleMsgBond(w.ctx, w.keeper, msgApply)
	c.Assert(result.IsOK(), Equals, false)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	// nodeAccoutn already exist
	w = getHandlerTestWrapper(c, 1, true, false)
	msgApply = NewMsgBond(w.activeNodeAccount.NodeAddress, bond, GetRandomTxHash(), bondAddr, w.activeNodeAccount.NodeAddress)
	result = handleMsgBond(w.ctx, w.keeper, msgApply)
	c.Assert(result.IsOK(), Equals, false)
	c.Assert(result.Code, Equals, sdk.CodeUnknownRequest)

	// invalid Msg
	invalidMsgApply := NewMsgBond(sdk.AccAddress{}, bond, GetRandomTxHash(), bondAddr, w.activeNodeAccount.NodeAddress)
	invalidMsgApplyResult := handleMsgBond(w.ctx, w.keeper, invalidMsgApply)
	c.Assert(invalidMsgApplyResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Assert(invalidMsgApplyResult.IsOK(), Equals, false)

	newAcc := GetRandomNodeAccount(NodeWhiteListed)
	// less than minimum bond
	msgApplyLessThanMinimumBond := NewMsgBond(newAcc.NodeAddress, sdk.NewUint(1000), GetRandomTxHash(), bondAddr, w.activeNodeAccount.NodeAddress)
	lessThanMinimumBondResult := handleMsgBond(w.ctx, w.keeper, msgApplyLessThanMinimumBond)
	c.Assert(lessThanMinimumBondResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Assert(lessThanMinimumBondResult.IsOK(), Equals, false)

	msgApply1 := NewMsgBond(newAcc.NodeAddress, sdk.NewUint(100*common.One), GetRandomTxHash(), bondAddr, w.activeNodeAccount.NodeAddress)
	result = handleMsgBond(w.ctx, w.keeper, msgApply1)
	c.Assert(result.IsOK(), Equals, true)
	c.Assert(result.Code, Equals, sdk.CodeOK)
	coins := w.keeper.coinKeeper.GetCoins(w.ctx, newAcc.NodeAddress)
	c.Assert(coins.AmountOf("bep").Int64(), Equals, int64(1000))

	// apply again shohuld fail
	resultExist := handleMsgBond(w.ctx, w.keeper, msgApply1)
	c.Assert(resultExist.IsOK(), Equals, false)
	c.Assert(resultExist.Code, Equals, sdk.CodeUnknownRequest)

}

func (HandlerSuite) TestHandleMsgSetTrustAccount(c *C) {
	ctx, k := setupKeeperForTest(c)
	ctx = ctx.WithBlockHeight(1)
	signer := GetRandomBech32Addr()
	// add observer
	bepConsPubKey := GetRandomBech32ConsensusPubKey()
	bondAddr := GetRandomBNBAddress()
	pubKeys := GetRandomPubkeys()
	emptyPubKeys := common.PubKeys{}

	msgTrustAccount := types.NewMsgSetTrustAccount(pubKeys, bepConsPubKey, signer)
	unAuthorizedResult := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(unAuthorizedResult.Code, Equals, sdk.CodeUnauthorized)
	c.Check(unAuthorizedResult.IsOK(), Equals, false)
	bond := sdk.NewUint(common.One * 100)
	nodeAccount := NewNodeAccount(signer, NodeActive, emptyPubKeys, "", bond, bondAddr, ctx.BlockHeight())
	k.SetNodeAccount(ctx, nodeAccount)

	activeFailResult := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(activeFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(activeFailResult.IsOK(), Equals, false)

	nodeAccount = NewNodeAccount(signer, NodeDisabled, emptyPubKeys, "", bond, bondAddr, ctx.BlockHeight())
	k.SetNodeAccount(ctx, nodeAccount)

	disabledFailResult := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(disabledFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(disabledFailResult.IsOK(), Equals, false)

	k.SetNodeAccount(ctx, NewNodeAccount(signer, NodeWhiteListed, pubKeys, bepConsPubKey, bond, bondAddr, ctx.BlockHeight()))

	notUniqueFailResult := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(notUniqueFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(notUniqueFailResult.IsOK(), Equals, false)

	nodeAccount = NewNodeAccount(signer, NodeWhiteListed, emptyPubKeys, "", bond, bondAddr, ctx.BlockHeight())
	k.SetNodeAccount(ctx, nodeAccount)

	success := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(success.Code, Equals, sdk.CodeOK)
	c.Check(success.IsOK(), Equals, true)

}

func (HandlerSuite) TestIsSignedByActiveObserver(c *C) {
	ctx, k := setupKeeperForTest(c)
	nodeAddr := GetRandomBech32Addr()
	c.Check(isSignedByActiveObserver(ctx, k, []sdk.AccAddress{nodeAddr}), Equals, false)
	c.Check(isSignedByActiveObserver(ctx, k, []sdk.AccAddress{}), Equals, false)
}

func (HandlerSuite) TestIsSignedByActiveNodeAccounts(c *C) {
	ctx, k := setupKeeperForTest(c)
	nodeAddr := GetRandomBech32Addr()
	c.Check(isSignedByActiveNodeAccounts(ctx, k, []sdk.AccAddress{}), Equals, false)
	c.Check(isSignedByActiveNodeAccounts(ctx, k, []sdk.AccAddress{nodeAddr}), Equals, false)
	nodeAccount1 := GetRandomNodeAccount(NodeWhiteListed)
	k.SetNodeAccount(ctx, nodeAccount1)
	c.Check(isSignedByActiveNodeAccounts(ctx, k, []sdk.AccAddress{nodeAccount1.NodeAddress}), Equals, false)
}

func (HandlerSuite) TestHandleOperatorMsgEndPool(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	acc1 := GetRandomNodeAccount(NodeWhiteListed)
	bnbAddr := GetRandomBNBAddress()
	txHash := GetRandomTxHash()
	tx := common.NewTx(
		txHash,
		bnbAddr,
		GetRandomBNBAddress(),
		common.Coins{common.NewCoin(common.BNBAsset, sdk.OneUint())},
		"",
	)
	msgEndPool := NewMsgEndPool(common.BNBAsset, tx, acc1.NodeAddress)
	result := handleOperatorMsgEndPool(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, msgEndPool)
	c.Assert(result.IsOK(), Equals, false)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)
	msgEndPool = NewMsgEndPool(common.BNBAsset, tx, w.activeNodeAccount.NodeAddress)
	w.poolAddrMgr.BeginBlock(w.ctx)
	stakeTxHash := GetRandomTxHash()
	tx = common.NewTx(
		stakeTxHash,
		bnbAddr,
		GetRandomBNBAddress(),
		common.Coins{common.NewCoin(common.BNBAsset, sdk.OneUint())},
		"",
	)
	msgSetStake := NewMsgSetStakeData(
		tx,
		common.BNBAsset,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		bnbAddr,
		bnbAddr,
		w.activeNodeAccount.NodeAddress)
	stakeResult := handleMsgSetStakeData(w.ctx, w.keeper, msgSetStake)
	c.Assert(stakeResult.Code, Equals, sdk.CodeOK)
	p := w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(p.Empty(), Equals, false)
	c.Assert(p.BalanceRune.Uint64(), Equals, msgSetStake.RuneAmount.Uint64())
	c.Assert(p.BalanceAsset.Uint64(), Equals, msgSetStake.AssetAmount.Uint64())
	c.Assert(p.Status, Equals, PoolEnabled)
	w.txOutStore.NewBlock(1)
	// EndPool again
	msgEndPool1 := NewMsgEndPool(common.BNBAsset, tx, w.activeNodeAccount.NodeAddress)
	result1 := handleOperatorMsgEndPool(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, msgEndPool1)
	c.Assert(result1.Code, Equals, sdk.CodeOK, Commentf("%+v\n", result1))
	p1 := w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Check(p1.Status, Equals, PoolSuspended)
	c.Check(p1.BalanceAsset.Uint64(), Equals, uint64(0))
	c.Check(p1.BalanceRune.Uint64(), Equals, uint64(0))
	txOut := w.txOutStore.blockOut
	c.Check(txOut, NotNil)
	c.Check(len(txOut.TxArray) > 0, Equals, true)
	c.Check(txOut.Height, Equals, uint64(1))
	totalAsset := sdk.ZeroUint()
	totalRune := sdk.ZeroUint()
	for _, item := range txOut.TxArray {
		c.Assert(item.Valid(), IsNil)
		c.Assert(item.ToAddress.Equals(bnbAddr), Equals, true)
		if item.Coin.Asset.IsRune() {
			totalRune = totalRune.Add(item.Coin.Amount)
		} else {
			totalAsset = totalAsset.Add(item.Coin.Amount)
		}
	}
	c.Assert(totalAsset.Equal(msgSetStake.AssetAmount.SubUint64(singleTransactionFee)), Equals, true, Commentf("%d %d", totalAsset.Uint64(), msgSetStake.AssetAmount.SubUint64(singleTransactionFee).Uint64()))
	c.Assert(totalRune.Equal(msgSetStake.RuneAmount), Equals, true)
}

func (HandlerSuite) TestHandleMsgSetPoolData(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)

	msgSetPoolData := NewMsgSetPoolData(common.BNBAsset, PoolEnabled, w.notActiveNodeAccount.NodeAddress)
	result := handleMsgSetPoolData(w.ctx, w.keeper, msgSetPoolData)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	msgSetPoolData = NewMsgSetPoolData(common.BNBAsset, PoolEnabled, w.activeNodeAccount.NodeAddress)
	result1 := handleMsgSetPoolData(w.ctx, w.keeper, msgSetPoolData)
	c.Assert(result1.Code, Equals, sdk.CodeOK)

	msgSetPoolData1 := NewMsgSetPoolData(common.Asset{}, PoolEnabled, w.activeNodeAccount.NodeAddress)
	result2 := handleMsgSetPoolData(w.ctx, w.keeper, msgSetPoolData1)
	c.Assert(result2.Code, Equals, sdk.CodeUnknownRequest)
}

func (HandlerSuite) TestHandleMsgSetStakeData(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	bnbAddr := GetRandomBNBAddress()
	stakeTxHash := GetRandomTxHash()
	tx := common.NewTx(
		stakeTxHash,
		bnbAddr,
		GetRandomBNBAddress(),
		common.Coins{common.NewCoin(common.BNBAsset, sdk.OneUint())},
		"",
	)
	msgSetStake := NewMsgSetStakeData(
		tx,
		common.BNBAsset,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		bnbAddr,
		bnbAddr,
		w.notActiveNodeAccount.NodeAddress)
	stakeResult := handleMsgSetStakeData(w.ctx, w.keeper, msgSetStake)
	c.Assert(stakeResult.Code, Equals, sdk.CodeUnauthorized)

	p := w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(p.Empty(), Equals, true)
	msgSetStake = NewMsgSetStakeData(
		tx,
		common.BNBAsset,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		bnbAddr,
		bnbAddr,
		w.activeNodeAccount.NodeAddress)
	stakeResult1 := handleMsgSetStakeData(w.ctx, w.keeper, msgSetStake)
	c.Assert(stakeResult1.Code, Equals, sdk.CodeOK)

	p = w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(p.Empty(), Equals, false)
	c.Assert(p.BalanceRune.Uint64(), Equals, msgSetStake.RuneAmount.Uint64())
	c.Assert(p.BalanceAsset.Uint64(), Equals, msgSetStake.AssetAmount.Uint64())
	e, err := w.keeper.GetCompletedEvent(w.ctx, 1)
	c.Assert(err, IsNil)
	c.Assert(e.Status.Valid(), IsNil)
	c.Assert(e.InTx.ID.Equals(stakeTxHash), Equals, true)

	// Suspended pool should not allow stake
	w.keeper.SetPoolData(w.ctx, common.BNBAsset, PoolSuspended)

	msgSetStake1 := NewMsgSetStakeData(
		tx,
		common.BNBAsset,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		GetRandomBNBAddress(),
		GetRandomBNBAddress(),
		w.activeNodeAccount.NodeAddress)
	stakeResult2 := handleMsgSetStakeData(w.ctx, w.keeper, msgSetStake1)
	c.Assert(stakeResult2.Code, Equals, sdk.CodeUnknownRequest)

	msgSetStake2 := NewMsgSetStakeData(
		tx,
		common.BNBAsset,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		"",
		"",
		w.activeNodeAccount.NodeAddress)
	stakeResult3 := handleMsgSetStakeData(w.ctx, w.keeper, msgSetStake2)
	c.Assert(stakeResult3.Code, Equals, sdk.CodeUnknownRequest)
}

func (HandlerSuite) TestHandleMsgConfirmNextPoolAddress(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	// invalid msg
	msgNextPoolAddrInvalid := NewMsgNextPoolAddress(
		GetRandomTxHash(),
		common.EmptyPubKey,
		GetRandomBNBAddress(), common.BNBChain,
		w.activeNodeAccount.NodeAddress)

	c.Assert(handleMsgConfirmNextPoolAddress(w.ctx, w.keeper, w.poolAddrMgr, w.validatorMgr, w.txOutStore, msgNextPoolAddrInvalid).Code, Equals, sdk.CodeUnknownRequest)
	// rotation window not open
	msgNextPoolAddr := NewMsgNextPoolAddress(
		GetRandomTxHash(),
		GetRandomPubKey(),
		GetRandomBNBAddress(),
		common.BNBChain,
		w.activeNodeAccount.NodeAddress)
	result := handleMsgConfirmNextPoolAddress(w.ctx, w.keeper, w.poolAddrMgr, w.validatorMgr, w.txOutStore, msgNextPoolAddr)
	c.Assert(result.Code, Equals, sdk.CodeUnknownRequest)
	// next pool had been confirmed already
	w.ctx = w.ctx.WithBlockHeight(w.poolAddrMgr.currentPoolAddresses.RotateWindowOpenAt)
	w.poolAddrMgr.BeginBlock(w.ctx)

	w.poolAddrMgr.currentPoolAddresses.Next = common.PoolPubKeys{
		common.NewPoolPubKey(common.BNBChain, 0, GetRandomPubKey()),
	}
	result = handleMsgConfirmNextPoolAddress(w.ctx, w.keeper, w.poolAddrMgr, w.validatorMgr, w.txOutStore, msgNextPoolAddr)
	c.Assert(result.Code, Equals, sdk.CodeUnknownRequest)
	chainSenderAddr := w.poolAddrMgr.currentPoolAddresses.Current.GetByChain(common.BNBChain)
	senderAddr, err := chainSenderAddr.GetAddress()
	c.Assert(err, IsNil)
	w.poolAddrMgr.currentPoolAddresses.Next = common.EmptyPoolPubKeys
	msgNextPoolAddr = NewMsgNextPoolAddress(
		GetRandomTxHash(),
		GetRandomPubKey(),
		senderAddr,
		common.BNBChain,
		w.activeNodeAccount.NodeAddress)
	w.txOutStore.NewBlock(1)
	result = handleMsgConfirmNextPoolAddress(w.ctx, w.keeper, w.poolAddrMgr, w.validatorMgr, w.txOutStore, msgNextPoolAddr)
	c.Assert(result.Code, Equals, sdk.CodeOK)
	c.Assert(w.txOutStore.blockOut, NotNil)
	c.Assert(w.txOutStore.blockOut.TxArray, HasLen, 1)
	tai := w.txOutStore.blockOut.TxArray[0]
	c.Assert(tai, NotNil)
	c.Assert(tai.Memo, Equals, "ack")
	c.Assert(tai.Coin.Amount.Uint64(), Equals, uint64(1))
}

func (HandlerSuite) TestHandleMsgSetTxIn(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	w.keeper.SetPool(w.ctx, Pool{
		Asset:        common.BNBAsset,
		BalanceRune:  sdk.NewUint(100 * common.One),
		BalanceAsset: sdk.NewUint(100 * common.One),
	})
	txIn := types.NewTxIn(
		common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One)),
			common.NewCoin(common.RuneAsset(), sdk.NewUint(100*common.One)),
		},
		"stake:BNB",
		GetRandomBNBAddress(),
		GetRandomBNBAddress(),
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
		sdk.NewUint(1024),
		currentChainPool.PubKey)
	msgSetTxIn1 := types.NewMsgSetTxIn(
		[]TxInVoter{
			types.NewTxInVoter(GetRandomTxHash(), []TxIn{txIn1}),
		},
		w.activeNodeAccount.NodeAddress)
	result2 := handleMsgSetTxIn(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, msgSetTxIn1)
	c.Assert(result2.Code, Equals, sdk.CodeOK)
	p1 := w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(p1.BalanceRune.Uint64(), Equals, uint64(200*common.One))
	c.Assert(p1.BalanceRune.Uint64(), Equals, uint64(200*common.One))
	// pool staker
	ps, err := w.keeper.GetPoolStaker(w.ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Assert(ps.TotalUnits.GT(sdk.ZeroUint()), Equals, true)
	c.Check(w.keeper.SupportedChain(w.ctx, common.BNBChain), Equals, true)
}

func (HandlerSuite) TestHandleTxInCreateMemo(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	currentChainPool := w.poolAddrMgr.currentPoolAddresses.Current.GetByChain(common.BNBChain)
	c.Assert(currentChainPool, NotNil)
	txIn := types.NewTxIn(
		common.Coins{
			common.NewCoin(common.RuneAsset(), sdk.NewUint(1*common.One)),
		},
		"create:BNB",
		GetRandomBNBAddress(),
		GetRandomBNBAddress(),
		sdk.NewUint(1024),
		currentChainPool.PubKey)

	msgSetTxIn := types.NewMsgSetTxIn(
		[]TxInVoter{
			types.NewTxInVoter(GetRandomTxHash(), []TxIn{txIn}),
		},
		w.activeNodeAccount.NodeAddress)

	result := handleMsgSetTxIn(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, msgSetTxIn)
	c.Assert(result.Code, Equals, sdk.CodeOK)

	pool := w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(pool.Empty(), Equals, false)
	c.Assert(pool.Status, Equals, PoolEnabled)
	c.Assert(pool.PoolUnits.Uint64(), Equals, uint64(0))
	c.Assert(pool.BalanceRune.Uint64(), Equals, uint64(0))
	c.Assert(pool.BalanceAsset.Uint64(), Equals, uint64(0))
}

func (HandlerSuite) TestHandleTxInWithdrawMemo(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	currentChainPool := w.poolAddrMgr.currentPoolAddresses.Current.GetByChain(common.BNBChain)
	c.Assert(currentChainPool, NotNil)
	staker := GetRandomBNBAddress()
	// lets do a stake first, otherwise nothing to withdraw
	txStake := types.NewTxIn(
		common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One)),
			common.NewCoin(common.RuneAsset(), sdk.NewUint(100*common.One)),
		},
		"stake:BNB",
		staker,
		GetRandomBNBAddress(),
		sdk.NewUint(1024),
		currentChainPool.PubKey)

	msgStake := types.NewMsgSetTxIn(
		[]TxInVoter{
			types.NewTxInVoter(GetRandomTxHash(), []TxIn{txStake}),
		},
		w.activeNodeAccount.NodeAddress)
	result := handleMsgSetTxIn(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, msgStake)
	c.Assert(result.Code, Equals, sdk.CodeOK)
	w.txOutStore.CommitBlock(w.ctx)

	txIn := types.NewTxIn(
		common.Coins{
			common.NewCoin(common.RuneAsset(), sdk.NewUint(1*common.One)),
		},
		"withdraw:BNB",
		staker,
		GetRandomBNBAddress(),
		sdk.NewUint(1025),
		currentChainPool.PubKey)

	msgSetTxIn := types.NewMsgSetTxIn(
		[]TxInVoter{
			types.NewTxInVoter(GetRandomTxHash(), []TxIn{txIn}),
		},
		w.activeNodeAccount.NodeAddress)
	w.txOutStore.NewBlock(2)
	result1 := handleMsgSetTxIn(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, msgSetTxIn)
	c.Assert(result1.Code, Equals, sdk.CodeOK)
	pool := w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(pool.Empty(), Equals, false)
	c.Assert(pool.Status, Equals, PoolBootstrap)
	c.Assert(pool.PoolUnits.Uint64(), Equals, uint64(0))
	c.Assert(pool.BalanceRune.Uint64(), Equals, uint64(0))
	c.Assert(pool.BalanceAsset.Uint64(), Equals, uint64(0))

}

func (HandlerSuite) TestHandleMsgLeave(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)

	ygg := NewYggdrasil(w.activeNodeAccount.NodePubKey.Secp256k1)
	ygg.AddFunds(
		common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(500*common.One)),
			common.NewCoin(common.BTCAsset, sdk.NewUint(400*common.One)),
		},
	)
	w.keeper.SetYggdrasil(w.ctx, ygg)

	txID := GetRandomTxHash()
	senderBNB := GetRandomBNBAddress()
	tx := common.NewTx(
		txID,
		senderBNB,
		GetRandomBNBAddress(),
		common.Coins{common.NewCoin(common.BNBAsset, sdk.OneUint())},
		"",
	)
	msgLeave := NewMsgLeave(tx, w.notActiveNodeAccount.NodeAddress)
	c.Assert(msgLeave.ValidateBasic(), IsNil)
	result := handleMsgLeave(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, msgLeave)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	msgLeaveInvalidSender := NewMsgLeave(tx, w.activeNodeAccount.NodeAddress)
	// try to leave, invalid sender
	result1 := handleMsgLeave(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, msgLeaveInvalidSender)
	c.Assert(result1.Code, Equals, sdk.CodeUnknownRequest)

	// active node can't leave
	tx.ID = GetRandomTxHash()
	tx.FromAddress = w.activeNodeAccount.BondAddress
	msgLeaveActiveNode := NewMsgLeave(tx, w.activeNodeAccount.NodeAddress)
	resultActiveNode := handleMsgLeave(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, msgLeaveActiveNode)
	c.Assert(resultActiveNode.Code, Equals, sdk.CodeUnknownRequest)

	acc2 := GetRandomNodeAccount(NodeStandby)
	acc2.Bond = sdk.NewUint(100 * common.One)
	w.keeper.SetNodeAccount(w.ctx, acc2)

	tx.ID = ""
	tx.FromAddress = acc2.BondAddress
	invalidMsg := NewMsgLeave(tx, w.activeNodeAccount.NodeAddress)
	result3 := handleMsgLeave(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, invalidMsg)
	c.Assert(result3.Code, Equals, sdk.CodeUnknownRequest)

	tx.ID = GetRandomTxHash()
	msgLeave1 := NewMsgLeave(tx, w.activeNodeAccount.NodeAddress)
	result2 := handleMsgLeave(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, msgLeave1)
	c.Assert(result2.Code, Equals, sdk.CodeOK)
	c.Assert(w.txOutStore.blockOut.Valid(), IsNil)
	c.Assert(w.txOutStore.blockOut.IsEmpty(), Equals, false)
	c.Assert(w.txOutStore.blockOut.TxArray, HasLen, 2)

	// Ragnarok check. Ensure all bonders have a zero bond balance
	outbound := w.txOutStore.GetOutboundItems()
	c.Assert(outbound, HasLen, 2)
	memo := NewOutboundMemo(tx.ID)
	c.Check(outbound[0].Memo, Equals, memo.String())
	c.Check(outbound[1].Memo, Equals, "yggdrasil-")
}

func (HandlerSuite) TestHandleMsgOutboundTx(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	bnbAddr := GetRandomBNBAddress()
	txID := GetRandomTxHash()
	tx := common.NewTx(
		txID,
		bnbAddr,
		GetRandomBNBAddress(),
		common.Coins{common.NewCoin(common.BNBAsset, sdk.OneUint())},
		"",
	)
	msgOutboundTx := NewMsgOutboundTx(tx, txID, w.notActiveNodeAccount.NodeAddress)
	result := handleMsgOutboundTx(w.ctx, w.keeper, w.poolAddrMgr, msgOutboundTx)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	tx.ID = ""
	msgInvalidOutboundTx := NewMsgOutboundTx(tx, txID, w.activeNodeAccount.NodeAddress)
	result1 := handleMsgOutboundTx(w.ctx, w.keeper, w.poolAddrMgr, msgInvalidOutboundTx)
	c.Assert(result1.Code, Equals, sdk.CodeUnknownRequest, Commentf("%+v\n", result1))

	tx.ID = txID
	msgInvalidPool := NewMsgOutboundTx(tx, txID, w.activeNodeAccount.NodeAddress)
	result2 := handleMsgOutboundTx(w.ctx, w.keeper, w.poolAddrMgr, msgInvalidPool)
	c.Assert(result2.Code, Equals, sdk.CodeUnauthorized, Commentf("%+v\n", result2))

	w = getHandlerTestWrapper(c, 1, true, true)
	currentChainPool := w.poolAddrMgr.currentPoolAddresses.Current.GetByChain(common.BNBChain)
	c.Assert(currentChainPool, NotNil)

	ygg := NewYggdrasil(currentChainPool.PubKey)
	ygg.Coins = common.Coins{
		common.NewCoin(common.BNBAsset, sdk.NewUint(500*common.One)),
		common.NewCoin(common.BTCAsset, sdk.NewUint(400*common.One)),
	}
	w.keeper.SetYggdrasil(w.ctx, ygg)

	currentPoolAddr, err := currentChainPool.GetAddress()
	c.Assert(err, IsNil)
	tx.FromAddress = currentPoolAddr
	tx.Coins = common.Coins{
		common.NewCoin(common.BNBAsset, sdk.NewUint(200*common.One)),
		common.NewCoin(common.BTCAsset, sdk.NewUint(200*common.One)),
	}
	msgOutboundTxNormal := NewMsgOutboundTx(tx, txID, w.activeNodeAccount.NodeAddress)
	result3 := handleMsgOutboundTx(w.ctx, w.keeper, w.poolAddrMgr, msgOutboundTxNormal)
	c.Assert(result3.Code, Equals, sdk.CodeOK, Commentf("%+v\n", result3))
	ygg = w.keeper.GetYggdrasil(w.ctx, currentChainPool.PubKey)
	c.Check(ygg.GetCoin(common.BNBAsset).Amount.Equal(sdk.NewUint(300*common.One)), Equals, true)
	c.Check(ygg.GetCoin(common.BTCAsset).Amount.Equal(sdk.NewUint(200*common.One)), Equals, true)

	w.txOutStore.NewBlock(2)
	inTxID := GetRandomTxHash()
	// set a txin
	txIn1 := types.NewTxIn(
		common.Coins{
			common.NewCoin(common.RuneAsset(), sdk.NewUint(1*common.One)),
		},
		"swap:BNB",
		GetRandomBNBAddress(),
		GetRandomBNBAddress(),
		sdk.NewUint(1024),
		currentChainPool.PubKey)
	msgSetTxIn1 := types.NewMsgSetTxIn(
		[]TxInVoter{
			types.NewTxInVoter(inTxID, []TxIn{txIn1}),
		},
		w.activeNodeAccount.NodeAddress)
	ctx := w.ctx.WithBlockHeight(2)
	resultTxIn := handleMsgSetTxIn(ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, msgSetTxIn1)
	c.Assert(resultTxIn.Code, Equals, sdk.CodeOK)
	w.txOutStore.CommitBlock(ctx)
	tx.FromAddress = currentPoolAddr
	tx.ID = inTxID
	msgOutboundTxNormal1 := NewMsgOutboundTx(tx, inTxID, w.activeNodeAccount.NodeAddress)
	result4 := handleMsgOutboundTx(ctx, w.keeper, w.poolAddrMgr, msgOutboundTxNormal1)
	c.Assert(result4.Code, Equals, sdk.CodeOK)
	iterator := w.keeper.GetCompleteEventIterator(w.ctx)
	found := false
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var evt Event
		w.keeper.cdc.MustUnmarshalBinaryBare(iterator.Value(), &evt)
		if evt.InTx.ID.Equals(inTxID) {
			found = true
			break
		}
	}
	c.Assert(found, Equals, true)
}

func (HandlerSuite) TestHandleMsgSetAdminConfig(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)

	tx := GetRandomTx()
	msgSetAdminCfg := NewMsgSetAdminConfig(tx, GSLKey, "0.5", w.notActiveNodeAccount.NodeAddress)
	result := handleMsgSetAdminConfig(w.ctx, w.keeper, msgSetAdminCfg)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	msgSetAdminCfg = NewMsgSetAdminConfig(tx, GSLKey, "0.5", w.activeNodeAccount.NodeAddress)
	result1 := handleMsgSetAdminConfig(w.ctx, w.keeper, msgSetAdminCfg)
	c.Assert(result1.Code, Equals, sdk.CodeOK)

	msgInvalidSetAdminCfg := NewMsgSetAdminConfig(tx, "Whatever", "blablab", w.activeNodeAccount.NodeAddress)
	result2 := handleMsgSetAdminConfig(w.ctx, w.keeper, msgInvalidSetAdminCfg)
	c.Assert(result2.Code, Equals, sdk.CodeUnknownRequest)
}

func (HandlerSuite) TestHandleMsgAdd(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	msgAdd := NewMsgAdd(common.BNBAsset, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One), GetRandomTxHash(), w.notActiveNodeAccount.NodeAddress)
	result := handleMsgAdd(w.ctx, w.keeper, msgAdd)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	msgInvalidAdd := NewMsgAdd(common.Asset{}, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One), GetRandomTxHash(), w.activeNodeAccount.NodeAddress)
	result1 := handleMsgAdd(w.ctx, w.keeper, msgInvalidAdd)
	c.Assert(result1.Code, Equals, sdk.CodeUnknownRequest)

	msgAdd = NewMsgAdd(common.BNBAsset, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One), GetRandomTxHash(), w.activeNodeAccount.NodeAddress)
	result2 := handleMsgAdd(w.ctx, w.keeper, msgAdd)
	c.Assert(result2.Code, Equals, sdk.CodeUnknownRequest)

	pool := w.keeper.GetPool(w.ctx, common.BNBAsset)
	pool.Asset = common.BNBAsset
	pool.Status = PoolEnabled
	w.keeper.SetPool(w.ctx, pool)
	result3 := handleMsgAdd(w.ctx, w.keeper, msgAdd)
	c.Assert(result3.Code, Equals, sdk.CodeOK)
	pool = w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(pool.Status, Equals, PoolEnabled)
	c.Assert(pool.BalanceAsset.Uint64(), Equals, sdk.NewUint(100*common.One).Uint64())
	c.Assert(pool.BalanceRune.Uint64(), Equals, sdk.NewUint(100*common.One).Uint64())
	c.Assert(pool.PoolUnits.Uint64(), Equals, uint64(0))

}
func (HandlerSuite) TestHandleMsgAck(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	txID := GetRandomTxHash()
	sender := GetRandomBNBAddress()
	signer := GetRandomBech32Addr()
	nextPoolPubKey := GetRandomPubKey()
	// invalid msg
	msgAckInvalid := NewMsgAck("", sender, common.BNBChain, signer)
	result := handleMsgAck(w.ctx, w.keeper, w.poolAddrMgr, w.validatorMgr, msgAckInvalid)
	c.Assert(result.Code, Equals, sdk.CodeUnknownRequest)

	// Pool rotation window didn't open
	msgAck := NewMsgAck(txID, sender, common.BNBChain, signer)
	result = handleMsgAck(w.ctx, w.keeper, w.poolAddrMgr, w.validatorMgr, msgAck)
	c.Assert(result.Code, Equals, sdk.CodeUnknownRequest)

	w.ctx = w.ctx.WithBlockHeight(w.poolAddrMgr.currentPoolAddresses.RotateWindowOpenAt)
	// open the window
	w.poolAddrMgr.BeginBlock(w.ctx)
	// didn't observe next pool address
	result = handleMsgAck(w.ctx, w.keeper, w.poolAddrMgr, w.validatorMgr, msgAck)
	c.Assert(result.Code, Equals, sdk.CodeUnknownRequest)
	nextChainPoolPubKey := common.NewPoolPubKey(common.BNBChain, 0, nextPoolPubKey)
	w.poolAddrMgr.ObservedNextPoolAddrPubKey = common.PoolPubKeys{
		nextChainPoolPubKey,
	}
	// sender is not the same as the observed next pool public key
	result = handleMsgAck(w.ctx, w.keeper, w.poolAddrMgr, w.validatorMgr, msgAck)
	c.Assert(result.Code, Equals, sdk.CodeUnknownRequest)
	senderAddr, err := nextPoolPubKey.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)
	msgAck1 := NewMsgAck(txID, senderAddr, common.BNBChain, signer)
	result = handleMsgAck(w.ctx, w.keeper, w.poolAddrMgr, w.validatorMgr, msgAck1)
	c.Assert(result.Code, Equals, sdk.CodeOK)
	c.Assert(w.poolAddrMgr.ObservedNextPoolAddrPubKey.IsEmpty(), Equals, true)
	c.Assert(w.poolAddrMgr.currentPoolAddresses.Next.IsEmpty(), Equals, false)
	nodeAccount, err := w.keeper.GetNodeAccount(w.ctx, w.activeNodeAccount.NodeAddress)
	c.Assert(err, IsNil)
	c.Assert(nodeAccount.SignerMembership, HasLen, 1)
}

func (HandlerSuite) TestRefund(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)

	pool := Pool{
		Asset:        common.BNBAsset,
		BalanceRune:  sdk.NewUint(100 * common.One),
		BalanceAsset: sdk.NewUint(100 * common.One),
	}
	w.keeper.SetPool(w.ctx, pool)

	// test we create a refund transaction
	txin := TxIn{
		Sender: GetRandomBNBAddress(),
		Coins: common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One)),
		},
	}
	currentPoolAddr := w.poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(common.BNBChain)
	c.Assert(currentPoolAddr, NotNil)
	refundTx(w.ctx, GetRandomTxHash(), txin, w.txOutStore, w.keeper, currentPoolAddr.PubKey, currentPoolAddr.Chain, true)
	c.Assert(w.txOutStore.GetOutboundItems(), HasLen, 1)

	// check we DONT create a refund transaction when we don't have a pool for
	// the asset sent.
	lokiAsset, _ := common.NewAsset(fmt.Sprintf("BNB.LOKI"))
	txin = TxIn{
		Sender: GetRandomBNBAddress(),
		Coins: common.Coins{
			common.NewCoin(lokiAsset, sdk.NewUint(100*common.One)),
		},
	}

	refundTx(w.ctx, GetRandomTxHash(), txin, w.txOutStore, w.keeper, currentPoolAddr.PubKey, currentPoolAddr.Chain, true)
	c.Assert(w.txOutStore.GetOutboundItems(), HasLen, 1)
	pool = w.keeper.GetPool(w.ctx, lokiAsset)
	c.Assert(pool.BalanceAsset.Equal(sdk.NewUint(100*common.One)), Equals, true)

	// doing it a second time should add the assets again.
	refundTx(w.ctx, GetRandomTxHash(), txin, w.txOutStore, w.keeper, currentPoolAddr.PubKey, currentPoolAddr.Chain, true)
	c.Assert(w.txOutStore.GetOutboundItems(), HasLen, 1)
	pool = w.keeper.GetPool(w.ctx, lokiAsset)
	c.Assert(pool.BalanceAsset.Equal(sdk.NewUint(200*common.One)), Equals, true)
}

func (HandlerSuite) TestGetMsgSwapFromMemo(c *C) {
	m, err := ParseMemo("swap:BNB")
	swapMemo, ok := m.(SwapMemo)
	c.Assert(ok, Equals, true)
	c.Assert(err, IsNil)
	txin := TxIn{
		Sender: GetRandomBNBAddress(),
		Coins: common.Coins{
			common.NewCoin(
				common.BNBAsset,
				sdk.NewUint(100*common.One),
			),
			common.NewCoin(
				common.RuneAsset(),
				sdk.NewUint(100*common.One),
			),
		},
	}
	// more than one coin
	resultMsg, err := getMsgSwapFromMemo(swapMemo, GetRandomTxHash(), txin, GetRandomBech32Addr())
	c.Assert(err, NotNil)
	c.Assert(resultMsg, IsNil)

	txin1 := TxIn{
		Sender: GetRandomBNBAddress(),
		Coins: common.Coins{
			common.NewCoin(
				common.BNBAsset,
				sdk.NewUint(100*common.One),
			),
		},
	}

	// coin and the ticker is the same, thus no point to swap
	resultMsg1, err := getMsgSwapFromMemo(swapMemo, GetRandomTxHash(), txin1, GetRandomBech32Addr())
	c.Assert(resultMsg1, IsNil)
	c.Assert(err, NotNil)
}

func (HandlerSuite) TestGetMsgStakeFromMemo(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	m, err := ParseMemo("stake:BNB")
	c.Assert(err, IsNil)
	stakeMemo, ok := m.(StakeMemo)
	c.Assert(ok, Equals, true)
	c.Assert(err, IsNil)
	tcanAsset, err := common.NewAsset("BNB.TCAN-014")
	c.Assert(err, IsNil)
	runeAsset := common.RuneAsset()
	c.Assert(err, IsNil)
	txin := TxIn{
		Sender: GetRandomBNBAddress(),
		Coins: common.Coins{
			common.NewCoin(tcanAsset,
				sdk.NewUint(100*common.One)),
			common.NewCoin(runeAsset,
				sdk.NewUint(100*common.One)),
		},
	}
	msg, err := getMsgStakeFromMemo(w.ctx, stakeMemo, GetRandomTxHash(), &txin, GetRandomBech32Addr())
	c.Assert(msg, IsNil)
	c.Assert(err, NotNil)
	txin1 := TxIn{
		Sender: GetRandomBNBAddress(),
		Coins: common.Coins{
			common.NewCoin(runeAsset,
				sdk.NewUint(100*common.One)),
		},
	}
	// stake only rune should be fine
	msg1, err1 := getMsgStakeFromMemo(w.ctx, stakeMemo, GetRandomTxHash(), &txin1, GetRandomBech32Addr())
	c.Assert(msg1, NotNil)
	c.Assert(err1, IsNil)
	bnbAsset, err := common.NewAsset("BNB.BNB")
	c.Assert(err, IsNil)
	txin2 := TxIn{
		Sender: GetRandomBNBAddress(),
		Coins: common.Coins{
			common.NewCoin(bnbAsset,
				sdk.NewUint(100*common.One)),
		},
	}
	// stake only token should be fine
	msg2, err2 := getMsgStakeFromMemo(w.ctx, stakeMemo, GetRandomTxHash(), &txin2, GetRandomBech32Addr())
	c.Assert(msg2, NotNil)
	c.Assert(err2, IsNil)
	lokiAsset, _ := common.NewAsset(fmt.Sprintf("BNB.LOKI"))
	txin3 := TxIn{
		Sender: GetRandomBNBAddress(),
		Coins: common.Coins{
			common.NewCoin(tcanAsset,
				sdk.NewUint(100*common.One)),
			common.NewCoin(lokiAsset,
				sdk.NewUint(100*common.One)),
		},
	}
	// stake only token should be fine
	msg3, err3 := getMsgStakeFromMemo(w.ctx, stakeMemo, GetRandomTxHash(), &txin3, GetRandomBech32Addr())
	c.Assert(msg3, IsNil)
	c.Assert(err3, NotNil)
}
