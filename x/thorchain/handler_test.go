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

	"gitlab.com/thorchain/thornode/common"

	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

type HandlerSuite struct{}

var _ = Suite(&HandlerSuite{})

func (s *HandlerSuite) SetUpSuite(*C) {
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
	k := NewKVStore(bk, supplyKeeper, keyThorchain, cdc)
	return ctx, k
}

type handlerTestWrapper struct {
	ctx                  sdk.Context
	keeper               Keeper
	poolAddrMgr          PoolAddressManager
	validatorMgr         ValidatorManager
	txOutStore           TxOutStore
	activeNodeAccount    NodeAccount
	notActiveNodeAccount NodeAccount
}

func getHandlerTestWrapper(c *C, height int64, withActiveNode, withActieBNBPool bool) handlerTestWrapper {
	ctx, k := setupKeeperForTest(c)
	ctx = ctx.WithBlockHeight(height)
	acc1 := GetRandomNodeAccount(NodeActive)
	if withActiveNode {
		c.Assert(k.SetNodeAccount(ctx, acc1), IsNil)
	}
	if withActieBNBPool {
		p, err := k.GetPool(ctx, common.BNBAsset)
		c.Assert(err, IsNil)
		p.Asset = common.BNBAsset
		p.Status = PoolEnabled
		p.BalanceRune = sdk.NewUint(100 * common.One)
		p.BalanceAsset = sdk.NewUint(100 * common.One)
		c.Assert(k.SetPool(ctx, p), IsNil)
	}
	genesisPoolPubKey, err := common.NewPoolPubKey(common.BNBChain, 0, GetRandomPubKey())
	c.Assert(err, IsNil)
	genesisPoolAddress := NewPoolAddresses(common.EmptyPoolPubKeys, common.PoolPubKeys{
		genesisPoolPubKey,
	}, common.EmptyPoolPubKeys, 100, 90)
	k.SetPoolAddresses(ctx, genesisPoolAddress)
	poolAddrMgr := NewPoolAddressMgr(k)
	validatorMgr := NewValidatorMgr(k, poolAddrMgr)
	c.Assert(poolAddrMgr.BeginBlock(ctx), IsNil)
	validatorMgr.BeginBlock(ctx)
	txOutStore := NewTxOutStorage(k, poolAddrMgr)
	txOutStore.NewBlock(uint64(height))

	return handlerTestWrapper{
		ctx:                  ctx,
		keeper:               k,
		poolAddrMgr:          poolAddrMgr,
		validatorMgr:         validatorMgr,
		txOutStore:           txOutStore,
		activeNodeAccount:    acc1,
		notActiveNodeAccount: GetRandomNodeAccount(NodeDisabled),
	}
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
	c.Assert(k.SetNodeAccount(ctx, nodeAccount), IsNil)

	activeFailResult := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(activeFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(activeFailResult.IsOK(), Equals, false)

	nodeAccount = NewNodeAccount(signer, NodeDisabled, emptyPubKeys, "", bond, bondAddr, ctx.BlockHeight())
	c.Assert(k.SetNodeAccount(ctx, nodeAccount), IsNil)

	disabledFailResult := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(disabledFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(disabledFailResult.IsOK(), Equals, false)

	c.Assert(k.SetNodeAccount(ctx, NewNodeAccount(signer, NodeWhiteListed, pubKeys, bepConsPubKey, bond, bondAddr, ctx.BlockHeight())), IsNil)

	notUniqueFailResult := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(notUniqueFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(notUniqueFailResult.IsOK(), Equals, false)

	nodeAccount = NewNodeAccount(signer, NodeWhiteListed, emptyPubKeys, "", bond, bondAddr, ctx.BlockHeight())
	c.Assert(k.SetNodeAccount(ctx, nodeAccount), IsNil)

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
	c.Assert(k.SetNodeAccount(ctx, nodeAccount1), IsNil)
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
		common.BNBGasFeeSingleton,
		"",
	)
	msgEndPool := NewMsgEndPool(common.BNBAsset, tx, acc1.NodeAddress)
	result := handleOperatorMsgEndPool(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, msgEndPool)
	c.Assert(result.IsOK(), Equals, false)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)
	msgEndPool = NewMsgEndPool(common.BNBAsset, tx, w.activeNodeAccount.NodeAddress)
	c.Assert(w.poolAddrMgr.BeginBlock(w.ctx), IsNil)
	stakeTxHash := GetRandomTxHash()
	tx = common.NewTx(
		stakeTxHash,
		bnbAddr,
		GetRandomBNBAddress(),
		common.Coins{common.NewCoin(common.BNBAsset, sdk.OneUint())},
		common.BNBGasFeeSingleton,
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
	p, err := w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Assert(p.Empty(), Equals, false)
	c.Assert(p.BalanceRune.Uint64(), Equals, msgSetStake.RuneAmount.Uint64())
	c.Assert(p.BalanceAsset.Uint64(), Equals, msgSetStake.AssetAmount.Uint64())
	c.Assert(p.Status, Equals, PoolEnabled)
	w.txOutStore.NewBlock(1)
	// EndPool again
	msgEndPool1 := NewMsgEndPool(common.BNBAsset, tx, w.activeNodeAccount.NodeAddress)
	result1 := handleOperatorMsgEndPool(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, msgEndPool1)
	c.Assert(result1.Code, Equals, sdk.CodeOK, Commentf("%+v\n", result1))
	p1, err := w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Check(p1.Status, Equals, PoolSuspended)
	c.Check(p1.BalanceAsset.Uint64(), Equals, uint64(0))
	c.Check(p1.BalanceRune.Uint64(), Equals, uint64(0))
	txOut := w.txOutStore.getBlockOut()
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
	c.Assert(totalAsset.Equal(msgSetStake.AssetAmount), Equals, true, Commentf("%d %d", totalAsset.Uint64(), msgSetStake.AssetAmount.Uint64()))
	c.Assert(totalRune.Equal(msgSetStake.RuneAmount), Equals, true)
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
		common.BNBGasFeeSingleton,
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

	p, err := w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(err, IsNil)
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

	p, err = w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Assert(p.Empty(), Equals, false)
	c.Assert(p.BalanceRune.Uint64(), Equals, msgSetStake.RuneAmount.Uint64())
	c.Assert(p.BalanceAsset.Uint64(), Equals, msgSetStake.AssetAmount.Uint64())
	e, err := w.keeper.GetCompletedEvent(w.ctx, 1)
	c.Assert(err, IsNil)
	c.Assert(e.Status.Valid(), IsNil)
	c.Assert(e.InTx.ID.Equals(stakeTxHash), Equals, true)

	// Suspended pool should not allow stake
	p.Status = PoolSuspended
	c.Assert(w.keeper.SetPool(w.ctx, p), IsNil)

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
		GetRandomTx(),
		common.EmptyPubKey,
		GetRandomBNBAddress(), common.BNBChain,
		w.activeNodeAccount.NodeAddress)

	c.Assert(handleMsgConfirmNextPoolAddress(w.ctx, w.keeper, w.poolAddrMgr, w.validatorMgr, w.txOutStore, msgNextPoolAddrInvalid).Code, Equals, sdk.CodeUnknownRequest)
	// rotation window not open
	msgNextPoolAddr := NewMsgNextPoolAddress(
		GetRandomTx(),
		GetRandomPubKey(),
		GetRandomBNBAddress(),
		common.BNBChain,
		w.activeNodeAccount.NodeAddress)
	result := handleMsgConfirmNextPoolAddress(w.ctx, w.keeper, w.poolAddrMgr, w.validatorMgr, w.txOutStore, msgNextPoolAddr)
	c.Assert(result.Code, Equals, sdk.CodeUnknownRequest)
	// next pool had been confirmed already
	w.ctx = w.ctx.WithBlockHeight(w.poolAddrMgr.GetCurrentPoolAddresses().RotateWindowOpenAt)
	c.Assert(w.poolAddrMgr.BeginBlock(w.ctx), IsNil)

	pk1, err := common.NewPoolPubKey(common.BNBChain, 0, GetRandomPubKey())
	c.Assert(err, IsNil)
	w.poolAddrMgr.GetCurrentPoolAddresses().Next = common.PoolPubKeys{pk1}
	result = handleMsgConfirmNextPoolAddress(w.ctx, w.keeper, w.poolAddrMgr, w.validatorMgr, w.txOutStore, msgNextPoolAddr)
	c.Assert(result.Code, Equals, sdk.CodeUnknownRequest)
	chainSenderAddr := w.poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(common.BNBChain)
	senderAddr, err := chainSenderAddr.GetAddress()
	c.Assert(err, IsNil)
	w.poolAddrMgr.GetCurrentPoolAddresses().Next = common.EmptyPoolPubKeys
	msgNextPoolAddr = NewMsgNextPoolAddress(
		GetRandomTx(),
		GetRandomPubKey(),
		senderAddr,
		common.BNBChain,
		w.activeNodeAccount.NodeAddress)
	w.txOutStore.NewBlock(1)
	result = handleMsgConfirmNextPoolAddress(w.ctx, w.keeper, w.poolAddrMgr, w.validatorMgr, w.txOutStore, msgNextPoolAddr)
	c.Assert(result.Code, Equals, sdk.CodeOK)
	c.Assert(w.txOutStore.getBlockOut(), NotNil)
	c.Assert(w.txOutStore.GetOutboundItems(), HasLen, 1)
	tai := w.txOutStore.GetOutboundItems()[0]
	c.Assert(tai, NotNil)
	c.Assert(tai.Memo, Equals, "ack")
	c.Assert(tai.Coin.Amount.Uint64(), Equals, uint64(1))
}

func (HandlerSuite) TestHandleTxInCreateMemo(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	currentChainPool := w.poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(common.BNBChain)
	c.Assert(currentChainPool, NotNil)
	txIn := types.NewObservedTx(
		common.Tx{
			ID:          GetRandomTxHash(),
			Chain:       common.BNBChain,
			Coins:       common.Coins{common.NewCoin(common.RuneAsset(), sdk.NewUint(1*common.One))},
			Memo:        "create:BNB",
			FromAddress: GetRandomBNBAddress(),
			ToAddress:   currentChainPool.Address,
			Gas:         common.BNBGasFeeSingleton,
		},
		sdk.NewUint(1024),
		currentChainPool.PubKey,
	)

	msg := types.NewMsgObservedTxIn(
		ObservedTxs{
			txIn,
		},
		w.activeNodeAccount.NodeAddress,
	)

	handler := NewHandler(w.keeper, w.poolAddrMgr, w.txOutStore, w.validatorMgr)
	result := handler(w.ctx, msg)
	c.Assert(result.Code, Equals, sdk.CodeOK, Commentf("%s\n", result.Log))

	pool, err := w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Assert(pool.Empty(), Equals, false)
	c.Assert(pool.Status, Equals, PoolEnabled)
	c.Assert(pool.PoolUnits.Uint64(), Equals, uint64(0))
	c.Assert(pool.BalanceRune.Uint64(), Equals, uint64(0))
	c.Assert(pool.BalanceAsset.Uint64(), Equals, uint64(0))
}

func (HandlerSuite) TestHandleTxInWithdrawMemo(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	currentChainPool := w.poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(common.BNBChain)
	c.Assert(currentChainPool, NotNil)
	staker := GetRandomBNBAddress()
	// lets do a stake first, otherwise nothing to withdraw
	txStake := types.NewObservedTx(
		common.Tx{
			ID:    GetRandomTxHash(),
			Chain: common.BNBChain,
			Coins: common.Coins{
				common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One)),
				common.NewCoin(common.RuneAsset(), sdk.NewUint(100*common.One)),
			},
			Memo:        "stake:BNB",
			FromAddress: staker,
			ToAddress:   currentChainPool.Address,
			Gas:         common.BNBGasFeeSingleton,
		},
		sdk.NewUint(1024),
		currentChainPool.PubKey,
	)

	msg := types.NewMsgObservedTxIn(
		ObservedTxs{
			txStake,
		},
		w.activeNodeAccount.NodeAddress,
	)

	handler := NewHandler(w.keeper, w.poolAddrMgr, w.txOutStore, w.validatorMgr)
	result := handler(w.ctx, msg)
	c.Assert(result.Code, Equals, sdk.CodeOK, Commentf("%s\n", result.Log))

	txStake = types.NewObservedTx(
		common.Tx{
			ID:    GetRandomTxHash(),
			Chain: common.BNBChain,
			Coins: common.Coins{
				common.NewCoin(common.RuneAsset(), sdk.NewUint(1*common.One)),
			},
			Memo:        "withdraw:BNB",
			FromAddress: staker,
			ToAddress:   currentChainPool.Address,
			Gas:         common.BNBGasFeeSingleton,
		},
		sdk.NewUint(1024),
		currentChainPool.PubKey,
	)

	msg = types.NewMsgObservedTxIn(
		ObservedTxs{
			txStake,
		},
		w.activeNodeAccount.NodeAddress,
	)

	w.txOutStore.NewBlock(2)
	result = handler(w.ctx, msg)
	c.Assert(result.Code, Equals, sdk.CodeOK, Commentf("%s\n", result.Log))

	pool, err := w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Assert(pool.Empty(), Equals, false)
	c.Assert(pool.Status, Equals, PoolBootstrap)
	c.Assert(pool.PoolUnits.Uint64(), Equals, uint64(0))
	c.Assert(pool.BalanceRune.Uint64(), Equals, uint64(0))
	c.Assert(pool.BalanceAsset.Uint64(), Equals, uint64(0))

}

func (HandlerSuite) TestHandleMsgOutboundTx(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	currentChainPool := w.poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(common.BNBChain)
	handler := NewHandler(w.keeper, w.poolAddrMgr, w.txOutStore, w.validatorMgr)

	tx := common.Tx{
		ID:          GetRandomTxHash(),
		Chain:       common.BNBChain,
		Coins:       common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(1*common.One))},
		Memo:        "",
		FromAddress: GetRandomBNBAddress(),
		ToAddress:   currentChainPool.Address,
		Gas:         common.BNBGasFeeSingleton,
	}

	msgOutboundTx := NewMsgOutboundTx(tx, tx.ID, w.notActiveNodeAccount.NodeAddress)
	result := handleMsgOutboundTx(w.ctx, w.keeper, w.poolAddrMgr, msgOutboundTx)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	tx.ID = ""
	msgInvalidOutboundTx := NewMsgOutboundTx(tx, tx.ID, w.activeNodeAccount.NodeAddress)
	result1 := handleMsgOutboundTx(w.ctx, w.keeper, w.poolAddrMgr, msgInvalidOutboundTx)
	c.Assert(result1.Code, Equals, sdk.CodeUnknownRequest, Commentf("%+v\n", result1))

	tx.ID = GetRandomTxHash()
	msgInvalidPool := NewMsgOutboundTx(tx, tx.ID, w.activeNodeAccount.NodeAddress)
	result2 := handleMsgOutboundTx(w.ctx, w.keeper, w.poolAddrMgr, msgInvalidPool)
	c.Assert(result2.Code, Equals, sdk.CodeUnauthorized, Commentf("%+v\n", result2))

	w = getHandlerTestWrapper(c, 1, true, true)
	currentChainPool = w.poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(common.BNBChain)
	c.Assert(currentChainPool, NotNil)

	ygg := NewYggdrasil(currentChainPool.PubKey)
	ygg.Coins = common.Coins{
		common.NewCoin(common.BNBAsset, sdk.NewUint(500*common.One)),
		common.NewCoin(common.BTCAsset, sdk.NewUint(400*common.One)),
	}
	c.Assert(w.keeper.SetYggdrasil(w.ctx, ygg), IsNil)

	currentPoolAddr, err := currentChainPool.GetAddress()
	c.Assert(err, IsNil)
	tx.FromAddress = currentPoolAddr
	tx.Coins = common.Coins{
		common.NewCoin(common.BNBAsset, sdk.NewUint(200*common.One)),
		common.NewCoin(common.BTCAsset, sdk.NewUint(200*common.One)),
	}
	msgOutboundTxNormal := NewMsgOutboundTx(tx, tx.ID, w.activeNodeAccount.NodeAddress)
	result3 := handleMsgOutboundTx(w.ctx, w.keeper, w.poolAddrMgr, msgOutboundTxNormal)
	c.Assert(result3.Code, Equals, sdk.CodeOK, Commentf("%+v\n", result3))
	ygg, err = w.keeper.GetYggdrasil(w.ctx, currentChainPool.PubKey)
	c.Assert(err, IsNil)
	c.Check(ygg.GetCoin(common.BNBAsset).Amount.Equal(sdk.NewUint(300*common.One)), Equals, true)
	c.Check(ygg.GetCoin(common.BTCAsset).Amount.Equal(sdk.NewUint(200*common.One)), Equals, true)

	w.txOutStore.NewBlock(2)
	inTxID := GetRandomTxHash()

	txIn := types.NewObservedTx(
		common.Tx{
			ID:          inTxID,
			Chain:       common.BNBChain,
			Coins:       common.Coins{common.NewCoin(common.RuneAsset(), sdk.NewUint(1*common.One))},
			Memo:        "swap:BNB",
			FromAddress: GetRandomBNBAddress(),
			ToAddress:   currentChainPool.Address,
			Gas:         common.BNBGasFeeSingleton,
		},
		sdk.NewUint(1024),
		currentChainPool.PubKey,
	)

	msg := types.NewMsgObservedTxIn(ObservedTxs{txIn}, w.activeNodeAccount.NodeAddress)
	result = handler(w.ctx, msg)
	c.Assert(result.Code, Equals, sdk.CodeOK, Commentf("%s\n", result.Log))

	tx = common.Tx{
		ID:          GetRandomTxHash(),
		Chain:       common.BNBChain,
		Coins:       common.Coins{common.NewCoin(common.RuneAsset(), sdk.NewUint(1*common.One))},
		Memo:        "swap:BNB",
		FromAddress: currentChainPool.Address,
		ToAddress:   GetRandomBNBAddress(),
		Gas:         common.BNBGasFeeSingleton,
	}

	outMsg := NewMsgOutboundTx(tx, inTxID, w.activeNodeAccount.NodeAddress)
	ctx := w.ctx.WithBlockHeight(2)
	result4 := handleMsgOutboundTx(ctx, w.keeper, w.poolAddrMgr, outMsg)
	c.Assert(result4.Code, Equals, sdk.CodeOK, Commentf("%+v\n", result4))

	w.txOutStore.CommitBlock(ctx)
	tx.FromAddress = currentPoolAddr
	tx.ID = inTxID
	result = handler(ctx, msg)
	c.Assert(result.Code, Equals, sdk.CodeOK)

	iterator := w.keeper.GetCompleteEventIterator(w.ctx)
	found := false
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var evt Event
		w.keeper.Cdc().MustUnmarshalBinaryBare(iterator.Value(), &evt)
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
	tx := GetRandomTx()
	msgAdd := NewMsgAdd(tx, common.BNBAsset, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One), w.notActiveNodeAccount.NodeAddress)
	result := handleMsgAdd(w.ctx, w.keeper, msgAdd)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	msgInvalidAdd := NewMsgAdd(tx, common.Asset{}, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One), w.activeNodeAccount.NodeAddress)
	result1 := handleMsgAdd(w.ctx, w.keeper, msgInvalidAdd)
	c.Assert(result1.Code, Equals, sdk.CodeUnknownRequest)

	msgAdd = NewMsgAdd(tx, common.BNBAsset, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One), w.activeNodeAccount.NodeAddress)
	result2 := handleMsgAdd(w.ctx, w.keeper, msgAdd)
	c.Assert(result2.Code, Equals, sdk.CodeUnknownRequest)

	pool, err := w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	pool.Asset = common.BNBAsset
	pool.BalanceRune = sdk.NewUint(10 * common.One)
	pool.BalanceAsset = sdk.NewUint(20 * common.One)
	pool.Status = PoolEnabled
	c.Assert(w.keeper.SetPool(w.ctx, pool), IsNil)
	result3 := handleMsgAdd(w.ctx, w.keeper, msgAdd)
	c.Assert(result3.Code, Equals, sdk.CodeOK)
	pool, err = w.keeper.GetPool(w.ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Assert(pool.Status, Equals, PoolEnabled)
	c.Assert(pool.BalanceAsset.Uint64(), Equals, sdk.NewUint(120*common.One).Uint64())
	c.Assert(pool.BalanceRune.Uint64(), Equals, sdk.NewUint(110*common.One).Uint64())
	c.Assert(pool.PoolUnits.Uint64(), Equals, uint64(0))

}

func (HandlerSuite) TestRefund(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)

	pool := Pool{
		Asset:        common.BNBAsset,
		BalanceRune:  sdk.NewUint(100 * common.One),
		BalanceAsset: sdk.NewUint(100 * common.One),
	}
	c.Assert(w.keeper.SetPool(w.ctx, pool), IsNil)

	currentPoolAddr := w.poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(common.BNBChain)
	c.Assert(currentPoolAddr, NotNil)

	txin := NewObservedTx(
		common.Tx{
			ID:    GetRandomTxHash(),
			Chain: common.BNBChain,
			Coins: common.Coins{
				common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One)),
			},
			Memo:        "withdraw:BNB",
			FromAddress: GetRandomBNBAddress(),
			ToAddress:   currentPoolAddr.Address,
			Gas:         common.BNBGasFeeSingleton,
		},
		sdk.NewUint(1024),
		currentPoolAddr.PubKey,
	)

	c.Assert(refundTx(w.ctx, txin, w.txOutStore, w.keeper, currentPoolAddr.PubKey, currentPoolAddr.Chain, true), IsNil)
	c.Assert(w.txOutStore.GetOutboundItems(), HasLen, 1)

	// check THORNode DONT create a refund transaction when THORNode don't have a pool for
	// the asset sent.
	lokiAsset, _ := common.NewAsset(fmt.Sprintf("BNB.LOKI"))
	txin.Tx.Coins = common.Coins{
		common.NewCoin(lokiAsset, sdk.NewUint(100*common.One)),
	}

	c.Assert(refundTx(w.ctx, txin, w.txOutStore, w.keeper, currentPoolAddr.PubKey, currentPoolAddr.Chain, true), IsNil)
	c.Assert(w.txOutStore.GetOutboundItems(), HasLen, 1)
	var err error
	pool, err = w.keeper.GetPool(w.ctx, lokiAsset)
	c.Assert(err, IsNil)
	// pool should be zero since we drop coins we don't recognize on the floor
	c.Assert(pool.BalanceAsset.Equal(sdk.ZeroUint()), Equals, true, Commentf("%d", pool.BalanceAsset.Uint64()))

	// doing it a second time should keep it at zero
	c.Assert(refundTx(w.ctx, txin, w.txOutStore, w.keeper, currentPoolAddr.PubKey, currentPoolAddr.Chain, true), IsNil)
	c.Assert(w.txOutStore.GetOutboundItems(), HasLen, 1)
	pool, err = w.keeper.GetPool(w.ctx, lokiAsset)
	c.Assert(err, IsNil)
	c.Assert(pool.BalanceAsset.Equal(sdk.ZeroUint()), Equals, true)
}

func (HandlerSuite) TestGetMsgSwapFromMemo(c *C) {
	m, err := ParseMemo("swap:BNB")
	swapMemo, ok := m.(SwapMemo)
	c.Assert(ok, Equals, true)
	c.Assert(err, IsNil)

	txin := types.NewObservedTx(
		common.Tx{
			ID:    GetRandomTxHash(),
			Chain: common.BNBChain,
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
			Memo:        "withdraw:BNB",
			FromAddress: GetRandomBNBAddress(),
			ToAddress:   GetRandomBNBAddress(),
			Gas:         common.BNBGasFeeSingleton,
		},
		sdk.NewUint(1024),
		common.EmptyPubKey,
	)

	// more than one coin
	resultMsg, err := getMsgSwapFromMemo(swapMemo, txin, GetRandomBech32Addr())
	c.Assert(err, NotNil)
	c.Assert(resultMsg, IsNil)

	txin.Tx.Coins = common.Coins{
		common.NewCoin(
			common.BNBAsset,
			sdk.NewUint(100*common.One),
		),
	}

	// coin and the ticker is the same, thus no point to swap
	resultMsg1, err := getMsgSwapFromMemo(swapMemo, txin, GetRandomBech32Addr())
	c.Assert(resultMsg1, IsNil)
	c.Assert(err, NotNil)
}

func (HandlerSuite) TestGetMsgStakeFromMemo(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	// Stake BNB, however THORNode send T-CAN as coin , which is incorrect, should result in an error
	m, err := ParseMemo("stake:BNB")
	c.Assert(err, IsNil)
	stakeMemo, ok := m.(StakeMemo)
	c.Assert(ok, Equals, true)
	tcanAsset, err := common.NewAsset("BNB.TCAN-014")
	c.Assert(err, IsNil)
	runeAsset := common.RuneAsset()
	c.Assert(err, IsNil)

	txin := types.NewObservedTx(
		common.Tx{
			ID:    GetRandomTxHash(),
			Chain: common.BNBChain,
			Coins: common.Coins{
				common.NewCoin(tcanAsset,
					sdk.NewUint(100*common.One)),
				common.NewCoin(runeAsset,
					sdk.NewUint(100*common.One)),
			},
			Memo:        "withdraw:BNB",
			FromAddress: GetRandomBNBAddress(),
			ToAddress:   GetRandomBNBAddress(),
			Gas:         common.BNBGasFeeSingleton,
		},
		sdk.NewUint(1024),
		common.EmptyPubKey,
	)

	msg, err := getMsgStakeFromMemo(w.ctx, stakeMemo, txin, GetRandomBech32Addr())
	c.Assert(msg, IsNil)
	c.Assert(err, NotNil)

	// Asymentic stake should works fine, only RUNE
	txin.Tx.Coins = common.Coins{
		common.NewCoin(runeAsset,
			sdk.NewUint(100*common.One)),
	}

	// stake only rune should be fine
	msg1, err1 := getMsgStakeFromMemo(w.ctx, stakeMemo, txin, GetRandomBech32Addr())
	c.Assert(msg1, NotNil)
	c.Assert(err1, IsNil)

	bnbAsset, err := common.NewAsset("BNB.BNB")
	c.Assert(err, IsNil)
	txin.Tx.Coins = common.Coins{
		common.NewCoin(bnbAsset,
			sdk.NewUint(100*common.One)),
	}

	// stake only token(BNB) should be fine
	msg2, err2 := getMsgStakeFromMemo(w.ctx, stakeMemo, txin, GetRandomBech32Addr())
	c.Assert(msg2, NotNil)
	c.Assert(err2, IsNil)

	lokiAsset, _ := common.NewAsset(fmt.Sprintf("BNB.LOKI"))
	txin.Tx.Coins = common.Coins{
		common.NewCoin(tcanAsset,
			sdk.NewUint(100*common.One)),
		common.NewCoin(lokiAsset,
			sdk.NewUint(100*common.One)),
	}

	// stake only token should be fine
	msg3, err3 := getMsgStakeFromMemo(w.ctx, stakeMemo, txin, GetRandomBech32Addr())
	c.Assert(msg3, IsNil)
	c.Assert(err3, NotNil)

	// Make sure the RUNE Address and Asset Address set correctly
	txin.Tx.Coins = common.Coins{
		common.NewCoin(runeAsset,
			sdk.NewUint(100*common.One)),
		common.NewCoin(lokiAsset,
			sdk.NewUint(100*common.One)),
	}

	lokiStakeMemo, err := ParseMemo("stake:BNB.LOKI")
	c.Assert(err, IsNil)
	msg4, err4 := getMsgStakeFromMemo(w.ctx, lokiStakeMemo.(StakeMemo), txin, GetRandomBech32Addr())
	c.Assert(err4, IsNil)
	c.Assert(msg4, NotNil)
	msgStake := msg4.(MsgSetStakeData)
	c.Assert(msgStake, NotNil)
	c.Assert(msgStake.RuneAddress, Equals, txin.Tx.FromAddress)
	c.Assert(msgStake.AssetAddress, Equals, txin.Tx.FromAddress)
}
