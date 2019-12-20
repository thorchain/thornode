package thorchain

import (
	"fmt"

	"github.com/blang/semver"
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
	"gitlab.com/thorchain/thornode/constants"

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
	genesisPoolPubKey, err := common.NewPoolPubKey(common.BNBChain, nil, GetRandomPubKey())
	c.Assert(err, IsNil)
	genesisPoolAddress := NewPoolAddresses(common.EmptyPoolPubKeys, common.PoolPubKeys{
		genesisPoolPubKey,
	}, common.EmptyPoolPubKeys)
	k.SetPoolAddresses(ctx, genesisPoolAddress)
	ver := semver.MustParse("0.1.0")
	constAccessor := constants.GetConstantValues(ver)
	poolAddrMgr := NewPoolAddressMgr(k)
	poolAddrMgr.currentPoolAddresses = NewPoolAddresses(GetRandomPoolPubKeys(), GetRandomPoolPubKeys(), GetRandomPoolPubKeys())
	validatorMgr := NewValidatorMgr(k, poolAddrMgr)
	validatorMgr.BeginBlock(ctx, constAccessor)
	txOutStore := NewTxOutStorage(k, poolAddrMgr)
	txOutStore.NewBlock(uint64(height), constAccessor)

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
	ver := semver.MustParse("0.1.0")
	constAccessor := constants.GetConstantValues(ver)
	w.txOutStore.NewBlock(2, constAccessor)
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

	tx := NewObservedTx(common.Tx{
		ID:          GetRandomTxHash(),
		Chain:       common.BNBChain,
		Coins:       common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(1*common.One))},
		Memo:        "",
		FromAddress: GetRandomBNBAddress(),
		ToAddress:   currentChainPool.Address,
		Gas:         common.BNBGasFeeSingleton,
	}, sdk.NewUint(12), GetRandomPubKey())

	msgOutboundTx := NewMsgOutboundTx(tx, tx.Tx.ID, w.notActiveNodeAccount.NodeAddress)
	result := handleMsgOutboundTx(w.ctx, w.keeper, w.poolAddrMgr, msgOutboundTx)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	tx.Tx.ID = ""
	msgInvalidOutboundTx := NewMsgOutboundTx(tx, tx.Tx.ID, w.activeNodeAccount.NodeAddress)
	result1 := handleMsgOutboundTx(w.ctx, w.keeper, w.poolAddrMgr, msgInvalidOutboundTx)
	c.Assert(result1.Code, Equals, sdk.CodeUnknownRequest, Commentf("%+v\n", result1))

	tx.Tx.ID = GetRandomTxHash()
	msgInvalidPool := NewMsgOutboundTx(tx, tx.Tx.ID, w.activeNodeAccount.NodeAddress)
	result2 := handleMsgOutboundTx(w.ctx, w.keeper, w.poolAddrMgr, msgInvalidPool)
	c.Assert(result2.Code, Equals, sdk.CodeUnauthorized, Commentf("%+v\n", result2))

	w = getHandlerTestWrapper(c, 1, true, true)
	currentChainPool = w.poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(common.BNBChain)
	c.Assert(currentChainPool, NotNil)

	ygg := NewVault(w.ctx.BlockHeight(), ActiveVault, YggdrasilVault, currentChainPool.PubKey)
	ygg.Coins = common.Coins{
		common.NewCoin(common.BNBAsset, sdk.NewUint(500*common.One)),
		common.NewCoin(common.BTCAsset, sdk.NewUint(400*common.One)),
	}
	c.Assert(w.keeper.SetVault(w.ctx, ygg), IsNil)

	currentPoolAddr, err := currentChainPool.GetAddress()
	c.Assert(err, IsNil)
	tx.ObservedPubKey = currentChainPool.PubKey
	tx.Tx.FromAddress = currentPoolAddr
	tx.Tx.Coins = common.Coins{
		common.NewCoin(common.BNBAsset, sdk.NewUint(200*common.One)),
		common.NewCoin(common.BTCAsset, sdk.NewUint(200*common.One)),
	}
	msgOutboundTxNormal := NewMsgOutboundTx(tx, tx.Tx.ID, w.activeNodeAccount.NodeAddress)
	result3 := handleMsgOutboundTx(w.ctx, w.keeper, w.poolAddrMgr, msgOutboundTxNormal)
	c.Assert(result3.Code, Equals, sdk.CodeOK, Commentf("%+v\n", result3))
	ygg, err = w.keeper.GetVault(w.ctx, currentChainPool.PubKey)
	c.Assert(err, IsNil)
	c.Check(ygg.GetCoin(common.BNBAsset).Amount.Equal(sdk.NewUint(29999962500)), Equals, true) // 300 - Gas
	c.Check(ygg.GetCoin(common.BTCAsset).Amount.Equal(sdk.NewUint(200*common.One)), Equals, true)
	ver := semver.MustParse("0.1.0")
	constAccessor := constants.GetConstantValues(ver)
	w.txOutStore.NewBlock(2, constAccessor)
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

	tx = NewObservedTx(common.Tx{
		ID:          GetRandomTxHash(),
		Chain:       common.BNBChain,
		Coins:       common.Coins{common.NewCoin(common.RuneAsset(), sdk.NewUint(1*common.One))},
		Memo:        "swap:BNB",
		FromAddress: currentChainPool.Address,
		ToAddress:   GetRandomBNBAddress(),
		Gas:         common.BNBGasFeeSingleton,
	}, sdk.NewUint(12), GetRandomPubKey())

	outMsg := NewMsgOutboundTx(tx, inTxID, w.activeNodeAccount.NodeAddress)
	ctx := w.ctx.WithBlockHeight(2)
	result4 := handleMsgOutboundTx(ctx, w.keeper, w.poolAddrMgr, outMsg)
	c.Assert(result4.Code, Equals, sdk.CodeOK, Commentf("%+v\n", result4))

	w.txOutStore.CommitBlock(ctx)
	tx.Tx.FromAddress = currentPoolAddr
	tx.Tx.ID = inTxID
	result = handler(ctx, msg)
	c.Assert(result.Code, Equals, sdk.CodeOK)

	iterator := w.keeper.GetEventsIterator(w.ctx)
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

	c.Assert(refundTx(w.ctx, txin, w.txOutStore, w.keeper, true), IsNil)
	c.Assert(w.txOutStore.GetOutboundItems(), HasLen, 1)

	// check THORNode DONT create a refund transaction when THORNode don't have a pool for
	// the asset sent.
	lokiAsset, _ := common.NewAsset(fmt.Sprintf("BNB.LOKI"))
	txin.Tx.Coins = common.Coins{
		common.NewCoin(lokiAsset, sdk.NewUint(100*common.One)),
	}

	c.Assert(refundTx(w.ctx, txin, w.txOutStore, w.keeper, true), IsNil)
	c.Assert(w.txOutStore.GetOutboundItems(), HasLen, 1)
	var err error
	pool, err = w.keeper.GetPool(w.ctx, lokiAsset)
	c.Assert(err, IsNil)
	// pool should be zero since we drop coins we don't recognize on the floor
	c.Assert(pool.BalanceAsset.Equal(sdk.ZeroUint()), Equals, true, Commentf("%d", pool.BalanceAsset.Uint64()))

	// doing it a second time should keep it at zero
	c.Assert(refundTx(w.ctx, txin, w.txOutStore, w.keeper, true), IsNil)
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
