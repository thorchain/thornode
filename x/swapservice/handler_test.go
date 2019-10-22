package swapservice

import (
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
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/x/swapservice/types"
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
	multiPerm      = "multiple permissions account"
	randomPerm     = "random permission"
	holder         = "holder"
	keySwapService = sdk.NewKVStoreKey(StoreKey)
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
	ms.MountStoreWithDB(keySwapService, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	err := ms.LoadLatestVersion()
	c.Assert(err, IsNil)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "statechain"}, false, log.NewNopLogger())
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
	k := NewKeeper(bk, supplyKeeper, keySwapService, cdc)
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
		p := k.GetPool(ctx, common.BNBTicker)
		p.Ticker = common.BNBTicker
		p.Status = PoolEnabled
		p.BalanceRune = sdk.NewUint(100 * common.One)
		p.BalanceToken = sdk.NewUint(100 * common.One)
		k.SetPool(ctx, p)
	}
	poolAddrMgr := NewPoolAddressManager(k)
	validatorMgr := NewValidatorManager(k)
	poolAddrMgr.BeginBlock(ctx, height)
	validatorMgr.BeginBlock(ctx, height)
	txOutStore := NewTxOutStore(k)
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
	msgApply := NewMsgBond(w.activeNodeAccount.NodeAddress, bond, GetRandomTxHash(), bondAddr, w.activeNodeAccount.Accounts.ObserverBEPAddress)
	c.Assert(msgApply.ValidateBasic(), IsNil)
	result := handleMsgBond(w.ctx, w.keeper, msgApply)
	c.Assert(result.IsOK(), Equals, false)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	// nodeAccoutn already exist
	w = getHandlerTestWrapper(c, 1, true, false)
	msgApply = NewMsgBond(w.activeNodeAccount.NodeAddress, bond, GetRandomTxHash(), bondAddr, w.activeNodeAccount.Accounts.ObserverBEPAddress)
	result = handleMsgBond(w.ctx, w.keeper, msgApply)
	c.Assert(result.IsOK(), Equals, false)
	c.Assert(result.Code, Equals, sdk.CodeUnknownRequest)

	// invalid Msg
	invalidMsgApply := NewMsgBond(sdk.AccAddress{}, bond, GetRandomTxHash(), bondAddr, w.activeNodeAccount.Accounts.ObserverBEPAddress)
	invalidMsgApplyResult := handleMsgBond(w.ctx, w.keeper, invalidMsgApply)
	c.Assert(invalidMsgApplyResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Assert(invalidMsgApplyResult.IsOK(), Equals, false)

	newAcc := GetRandomNodeAccount(NodeWhiteListed)
	// less than minimum bond
	msgApplyLessThanMinimumBond := NewMsgBond(newAcc.NodeAddress, sdk.NewUint(1000), GetRandomTxHash(), bondAddr, w.activeNodeAccount.Accounts.ObserverBEPAddress)
	lessThanMinimumBondResult := handleMsgBond(w.ctx, w.keeper, msgApplyLessThanMinimumBond)
	c.Assert(lessThanMinimumBondResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Assert(lessThanMinimumBondResult.IsOK(), Equals, false)

	msgApply1 := NewMsgBond(newAcc.NodeAddress, sdk.NewUint(100*common.One), GetRandomTxHash(), bondAddr, w.activeNodeAccount.Accounts.ObserverBEPAddress)
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
	nodeAddr := GetRandomBech32Addr()
	signer := GetRandomBech32Addr()
	// add observer
	bnb, err := common.NewAddress("bnb1xlvns0n2mxh77mzaspn2hgav4rr4m8eerfju38")
	c.Assert(err, IsNil)
	bepConsPubKey := `bepcpub1zcjduepq4kn64fcjhf0fp20gp8var0rm25ca9jy6jz7acem8gckh0nkplznq85gdrg`

	bondAddr := GetRandomBNBAddress()
	trustAccount := NewTrustAccount(bnb, signer, bepConsPubKey)
	msgTrustAccount := types.NewMsgSetTrustAccount(trustAccount, signer)
	unAuthorizedResult := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(unAuthorizedResult.Code, Equals, sdk.CodeUnauthorized)
	c.Check(unAuthorizedResult.IsOK(), Equals, false)
	bond := sdk.NewUint(common.One * 100)
	nodeAccount := NewNodeAccount(signer, NodeActive, NewTrustAccount(common.NoAddress, sdk.AccAddress{}, ""), bond, bondAddr)
	k.SetNodeAccount(ctx, nodeAccount)

	activeFailResult := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(activeFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(activeFailResult.IsOK(), Equals, false)

	nodeAccount = NewNodeAccount(signer, NodeDisabled, NewTrustAccount(common.NoAddress, sdk.AccAddress{}, ""), bond, bondAddr)
	k.SetNodeAccount(ctx, nodeAccount)

	disabledFailResult := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(disabledFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(disabledFailResult.IsOK(), Equals, false)

	k.SetNodeAccount(ctx, NewNodeAccount(signer, NodeWhiteListed, NewTrustAccount(bnb, nodeAddr, bepConsPubKey), bond, bondAddr))

	notUniqueFailResult := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(notUniqueFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(notUniqueFailResult.IsOK(), Equals, false)

	nodeAccount = NewNodeAccount(signer, NodeWhiteListed, NewTrustAccount(common.NoAddress, sdk.AccAddress{}, ""), bond, bondAddr)
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
	ctx, k := setupKeeperForTest(c)
	txOutStore := NewTxOutStore(k)
	poolAddrMgr := NewPoolAddressManager(k)
	acc1 := GetRandomNodeAccount(NodeWhiteListed)
	bnbAddr := GetRandomBNBAddress()
	txHash := GetRandomTxHash()
	msgEndPool := NewMsgEndPool(common.BNBTicker, bnbAddr, txHash, acc1.NodeAddress)
	result := handleOperatorMsgEndPool(ctx, k, txOutStore, poolAddrMgr, msgEndPool)
	c.Assert(result.IsOK(), Equals, false)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	// do some stake
	acc1.UpdateStatus(NodeActive)
	k.SetNodeAccount(ctx, acc1)
	poolAddrMgr.BeginBlock(ctx, 1)
	stakeTxHash := GetRandomTxHash()
	msgSetStake := NewMsgSetStakeData(common.BNBTicker,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		bnbAddr,
		stakeTxHash,
		acc1.Accounts.ObserverBEPAddress)
	stakeResult := handleMsgSetStakeData(ctx, k, msgSetStake)
	c.Assert(stakeResult.Code, Equals, sdk.CodeOK)
	p := k.GetPool(ctx, common.BNBTicker)
	c.Assert(p.Empty(), Equals, false)
	c.Assert(p.BalanceRune.Uint64(), Equals, msgSetStake.RuneAmount.Uint64())
	c.Assert(p.BalanceToken.Uint64(), Equals, msgSetStake.TokenAmount.Uint64())
	c.Assert(p.Status, Equals, PoolEnabled)
	txOutStore.NewBlock(1)
	// EndPool again
	msgEndPool1 := NewMsgEndPool(common.BNBTicker, bnbAddr, txHash, acc1.NodeAddress)
	result1 := handleOperatorMsgEndPool(ctx, k, txOutStore, poolAddrMgr, msgEndPool1)
	c.Assert(result1.Code, Equals, sdk.CodeOK)
	p1 := k.GetPool(ctx, common.BNBTicker)
	c.Check(p1.Status, Equals, PoolSuspended)
	c.Check(p1.BalanceToken.Uint64(), Equals, uint64(0))
	c.Check(p1.BalanceRune.Uint64(), Equals, uint64(0))
	txOut := txOutStore.blockOut
	c.Check(txOut, NotNil)
	c.Check(len(txOut.TxArray) > 0, Equals, true)
	c.Check(txOut.Height, Equals, uint64(1))
	totalToken := sdk.ZeroUint()
	totalRune := sdk.ZeroUint()
	for _, item := range txOut.TxArray {
		c.Assert(item.Valid(), IsNil)
		c.Assert(item.ToAddress.Equals(bnbAddr), Equals, true)
		for _, co := range item.Coins {
			if common.IsRune(co.Denom) {
				totalRune = totalRune.Add(co.Amount)
			} else {
				totalToken = totalToken.Add(co.Amount)
			}
		}
	}
	c.Assert(totalToken.Equal(msgSetStake.TokenAmount.SubUint64(2*batchTransactionFee)), Equals, true)
	c.Assert(totalRune.Equal(msgSetStake.RuneAmount), Equals, true)
}

func (HandlerSuite) TestHandleMsgSetPoolData(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)

	msgSetPoolData := NewMsgSetPoolData(common.BNBTicker, PoolEnabled, w.notActiveNodeAccount.Accounts.ObserverBEPAddress)
	result := handleMsgSetPoolData(w.ctx, w.keeper, msgSetPoolData)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	msgSetPoolData = NewMsgSetPoolData(common.BNBTicker, PoolEnabled, w.activeNodeAccount.Accounts.ObserverBEPAddress)
	result1 := handleMsgSetPoolData(w.ctx, w.keeper, msgSetPoolData)
	c.Assert(result1.Code, Equals, sdk.CodeOK)

	msgSetPoolData1 := NewMsgSetPoolData("", PoolEnabled, w.activeNodeAccount.Accounts.ObserverBEPAddress)
	result2 := handleMsgSetPoolData(w.ctx, w.keeper, msgSetPoolData1)
	c.Assert(result2.Code, Equals, sdk.CodeUnknownRequest)
}

func (HandlerSuite) TestHandleMsgSetStakeData(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	bnbAddr := GetRandomBNBAddress()
	stakeTxHash := GetRandomTxHash()
	msgSetStake := NewMsgSetStakeData(common.BNBTicker,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		bnbAddr,
		stakeTxHash,
		w.notActiveNodeAccount.Accounts.ObserverBEPAddress)
	stakeResult := handleMsgSetStakeData(w.ctx, w.keeper, msgSetStake)
	c.Assert(stakeResult.Code, Equals, sdk.CodeUnauthorized)

	p := w.keeper.GetPool(w.ctx, common.BNBTicker)
	c.Assert(p.Empty(), Equals, true)
	msgSetStake = NewMsgSetStakeData(common.BNBTicker,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		bnbAddr,
		stakeTxHash,
		w.activeNodeAccount.Accounts.ObserverBEPAddress)
	stakeResult1 := handleMsgSetStakeData(w.ctx, w.keeper, msgSetStake)
	c.Assert(stakeResult1.Code, Equals, sdk.CodeOK)

	p = w.keeper.GetPool(w.ctx, common.BNBTicker)
	c.Assert(p.Empty(), Equals, false)
	c.Assert(p.BalanceRune.Uint64(), Equals, msgSetStake.RuneAmount.Uint64())
	c.Assert(p.BalanceToken.Uint64(), Equals, msgSetStake.TokenAmount.Uint64())
	e, err := w.keeper.GetCompletedEvent(w.ctx, 1)
	c.Assert(err, IsNil)
	c.Assert(e.Pool.Equals(common.BNBTicker), Equals, true)
	c.Assert(e.Status.Valid(), IsNil)
	c.Assert(e.InHash.Equals(stakeTxHash), Equals, true)

	// Suspended pool should not allow stake
	w.keeper.SetPoolData(w.ctx, common.BNBTicker, PoolSuspended)

	msgSetStake1 := NewMsgSetStakeData(common.BNBTicker,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		GetRandomBNBAddress(),
		GetRandomTxHash(),
		w.activeNodeAccount.Accounts.ObserverBEPAddress)
	stakeResult2 := handleMsgSetStakeData(w.ctx, w.keeper, msgSetStake1)
	c.Assert(stakeResult2.Code, Equals, sdk.CodeUnknownRequest)

	msgSetStake2 := NewMsgSetStakeData(common.BNBTicker,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		"",
		"",
		w.activeNodeAccount.Accounts.ObserverBEPAddress)
	stakeResult3 := handleMsgSetStakeData(w.ctx, w.keeper, msgSetStake2)
	c.Assert(stakeResult3.Code, Equals, sdk.CodeUnknownRequest)
}

func (HandlerSuite) TestHandleMsgConfirmNextPoolAddress(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	msgNextPoolAddr := NewMsgNextPoolAddress(
		GetRandomTxHash(),
		GetRandomBNBAddress(),
		GetRandomBNBAddress(),
		w.activeNodeAccount.Accounts.ObserverBEPAddress)
	result := handleMsgConfirmNextPoolAddress(w.ctx, w.keeper, w.validatorMgr, w.poolAddrMgr, msgNextPoolAddr)
	c.Assert(result.Code, Equals, sdk.CodeUnknownRequest)

	acc2 := GetRandomNodeAccount(NodeStandby)
	w.keeper.SetNodeAccount(w.ctx, acc2)
	updates := w.validatorMgr.EndBlock(w.ctx, w.validatorMgr.Meta.RotateWindowOpenAtBlockHeight)
	c.Assert(updates, IsNil)
	// we nominated a node account
	c.Assert(w.validatorMgr.Meta.Nominated.IsEmpty(), Equals, false)
	msgNextPoolAddr1 := NewMsgNextPoolAddress(
		GetRandomTxHash(),
		acc2.Accounts.SignerBNBAddress,
		w.poolAddrMgr.GetCurrentPoolAddresses().Current,
		acc2.Accounts.ObserverBEPAddress)
	result1 := handleMsgConfirmNextPoolAddress(w.ctx, w.keeper, w.validatorMgr, w.poolAddrMgr, msgNextPoolAddr1)
	c.Assert(result1.Code, Equals, sdk.CodeOK)
	acc2, err := w.keeper.GetNodeAccount(w.ctx, acc2.NodeAddress)
	c.Assert(err, IsNil)
	c.Assert(acc2.Status, Equals, NodeStandby)

}

func (HandlerSuite) TestHandleMsgSetTxIn(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	txIn := types.NewTxIn(
		common.Coins{
			common.NewCoin(common.BNBChain, common.BNBTicker, sdk.NewUint(100*common.One)),
			common.NewCoin(common.BNBChain, common.RuneA1FTicker, sdk.NewUint(100*common.One)),
		},
		"stake:BNB",
		GetRandomBNBAddress(),
		sdk.NewUint(1024),
		GetRandomBNBAddress())

	msgSetTxIn := types.NewMsgSetTxIn(
		[]TxInVoter{
			types.NewTxInVoter(GetRandomTxHash(), []TxIn{txIn}),
		},
		w.notActiveNodeAccount.Accounts.ObserverBEPAddress)
	result := handleMsgSetTxIn(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, msgSetTxIn)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	w = getHandlerTestWrapper(c, 1, true, true)

	msgSetTxIn = types.NewMsgSetTxIn(
		[]TxInVoter{
			types.NewTxInVoter(GetRandomTxHash(), []TxIn{txIn}),
		},
		w.activeNodeAccount.Accounts.ObserverBEPAddress)
	// send to wrong pool address, refund
	result1 := handleMsgSetTxIn(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, msgSetTxIn)
	c.Assert(result1.Code, Equals, sdk.CodeOK)
	c.Assert(w.txOutStore.blockOut, NotNil)
	c.Assert(w.txOutStore.blockOut.Valid(), IsNil)
	c.Assert(w.txOutStore.blockOut.IsEmpty(), Equals, false)
	c.Assert(len(w.txOutStore.blockOut.TxArray), Equals, 1)
	// expect to refund two coins
	c.Assert(len(w.txOutStore.blockOut.TxArray[0].Coins), Equals, 2)

	txIn1 := types.NewTxIn(
		common.Coins{
			common.NewCoin(common.BNBChain, common.BNBTicker, sdk.NewUint(100*common.One)),
			common.NewCoin(common.BNBChain, common.RuneA1FTicker, sdk.NewUint(100*common.One)),
		},
		"stake:BNB",
		GetRandomBNBAddress(),
		sdk.NewUint(1024),
		w.activeNodeAccount.Accounts.SignerBNBAddress)
	msgSetTxIn1 := types.NewMsgSetTxIn(
		[]TxInVoter{
			types.NewTxInVoter(GetRandomTxHash(), []TxIn{txIn1}),
		},
		w.activeNodeAccount.Accounts.ObserverBEPAddress)
	result2 := handleMsgSetTxIn(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, msgSetTxIn1)
	c.Assert(result2.Code, Equals, sdk.CodeOK)
	p1 := w.keeper.GetPool(w.ctx, common.BNBTicker)
	c.Assert(p1.BalanceRune.Uint64(), Equals, uint64(200*common.One))
	c.Assert(p1.BalanceRune.Uint64(), Equals, uint64(200*common.One))
	// pool staker
	ps, err := w.keeper.GetPoolStaker(w.ctx, common.BNBTicker)
	c.Assert(err, IsNil)
	c.Assert(ps.TotalUnits.GT(sdk.ZeroUint()), Equals, true)

}

func (HandlerSuite) TestHandleTxInCreateMemo(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	txIn := types.NewTxIn(
		common.Coins{
			common.NewCoin(common.BNBChain, common.RuneA1FTicker, sdk.NewUint(1*common.One)),
		},
		"create:BNB",
		GetRandomBNBAddress(),
		sdk.NewUint(1024),
		w.poolAddrMgr.GetCurrentPoolAddresses().Current)

	msgSetTxIn := types.NewMsgSetTxIn(
		[]TxInVoter{
			types.NewTxInVoter(GetRandomTxHash(), []TxIn{txIn}),
		},
		w.activeNodeAccount.Accounts.ObserverBEPAddress)

	result := handleMsgSetTxIn(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, msgSetTxIn)
	c.Assert(result.Code, Equals, sdk.CodeOK)

	pool := w.keeper.GetPool(w.ctx, common.BNBTicker)
	c.Assert(pool.Empty(), Equals, false)
	c.Assert(pool.Status, Equals, PoolEnabled)
	c.Assert(pool.PoolUnits.Uint64(), Equals, uint64(0))
	c.Assert(pool.BalanceRune.Uint64(), Equals, uint64(0))
	c.Assert(pool.BalanceToken.Uint64(), Equals, uint64(0))
}

func (HandlerSuite) TestHandleTxInWithdrawMemo(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	staker := GetRandomBNBAddress()
	// lets do a stake first, otherwise nothing to withdraw
	txStake := types.NewTxIn(
		common.Coins{
			common.NewCoin(common.BNBChain, common.BNBTicker, sdk.NewUint(100*common.One)),
			common.NewCoin(common.BNBChain, common.RuneA1FTicker, sdk.NewUint(100*common.One)),
		},
		"stake:BNB",
		staker,
		sdk.NewUint(1024),
		w.activeNodeAccount.Accounts.SignerBNBAddress)

	msgStake := types.NewMsgSetTxIn(
		[]TxInVoter{
			types.NewTxInVoter(GetRandomTxHash(), []TxIn{txStake}),
		},
		w.activeNodeAccount.Accounts.ObserverBEPAddress)
	result := handleMsgSetTxIn(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, msgStake)
	c.Assert(result.Code, Equals, sdk.CodeOK)
	w.txOutStore.CommitBlock(w.ctx)

	txIn := types.NewTxIn(
		common.Coins{
			common.NewCoin(common.BNBChain, common.RuneA1FTicker, sdk.NewUint(1*common.One)),
		},
		"withdraw:BNB",
		staker,
		sdk.NewUint(1025),
		w.poolAddrMgr.GetCurrentPoolAddresses().Current)

	msgSetTxIn := types.NewMsgSetTxIn(
		[]TxInVoter{
			types.NewTxInVoter(GetRandomTxHash(), []TxIn{txIn}),
		},
		w.activeNodeAccount.Accounts.ObserverBEPAddress)
	w.txOutStore.NewBlock(2)
	result1 := handleMsgSetTxIn(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, msgSetTxIn)
	c.Assert(result1.Code, Equals, sdk.CodeOK)
	pool := w.keeper.GetPool(w.ctx, common.BNBTicker)
	c.Assert(pool.Empty(), Equals, false)
	c.Assert(pool.Status, Equals, PoolEnabled)
	c.Assert(pool.PoolUnits.Uint64(), Equals, uint64(0))
	c.Assert(pool.BalanceRune.Uint64(), Equals, uint64(0))
	c.Assert(pool.BalanceToken.Uint64(), Equals, uint64(0))

}

func (HandlerSuite) TestHandleMsgLeave(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	txID := GetRandomTxHash()
	senderBNB := GetRandomBNBAddress()
	msgLeave := NewMsgLeave(txID, senderBNB, w.notActiveNodeAccount.Accounts.ObserverBEPAddress)
	c.Assert(msgLeave.ValidateBasic(), IsNil)
	result := handleMsgLeave(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, msgLeave)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	msgLeaveInvalidSender := NewMsgLeave(txID, senderBNB, w.activeNodeAccount.Accounts.ObserverBEPAddress)
	// try to leave, invalid sender
	result1 := handleMsgLeave(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, msgLeaveInvalidSender)
	c.Assert(result1.Code, Equals, sdk.CodeUnknownRequest)

	// active node can't leave
	msgLeaveActiveNode := NewMsgLeave(GetRandomTxHash(), w.activeNodeAccount.Accounts.SignerBNBAddress, w.activeNodeAccount.Accounts.ObserverBEPAddress)
	resultActiveNode := handleMsgLeave(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, msgLeaveActiveNode)
	c.Assert(resultActiveNode.Code, Equals, sdk.CodeUnknownRequest)

	acc2 := GetRandomNodeAccount(NodeStandby)
	acc2.Bond = sdk.NewUint(100 * common.One)
	w.keeper.SetNodeAccount(w.ctx, acc2)

	msgLeave1 := NewMsgLeave(GetRandomTxHash(), acc2.BondAddress, w.activeNodeAccount.Accounts.ObserverBEPAddress)
	result2 := handleMsgLeave(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, msgLeave1)
	c.Assert(result2.Code, Equals, sdk.CodeOK)
	c.Assert(w.txOutStore.blockOut.Valid(), IsNil)
	c.Assert(w.txOutStore.blockOut.IsEmpty(), Equals, false)
	c.Assert(len(w.txOutStore.blockOut.TxArray) > 0, Equals, true)

	invalidMsg := NewMsgLeave("", acc2.Accounts.SignerBNBAddress, w.activeNodeAccount.Accounts.ObserverBEPAddress)
	result3 := handleMsgLeave(w.ctx, w.keeper, w.txOutStore, w.poolAddrMgr, invalidMsg)
	c.Assert(result3.Code, Equals, sdk.CodeUnknownRequest)
}

func (HandlerSuite) TestHandleMsgOutboundTx(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	msgOutboundTx := NewMsgOutboundTx(GetRandomTxHash(), 1, w.notActiveNodeAccount.Accounts.SignerBNBAddress, w.notActiveNodeAccount.Accounts.ObserverBEPAddress)
	result := handleMsgOutboundTx(w.ctx, w.keeper, w.poolAddrMgr, msgOutboundTx)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	msgInvalidOutboundTx := NewMsgOutboundTx("", 1, w.activeNodeAccount.Accounts.SignerBNBAddress, w.activeNodeAccount.Accounts.ObserverBEPAddress)
	result1 := handleMsgOutboundTx(w.ctx, w.keeper, w.poolAddrMgr, msgInvalidOutboundTx)
	c.Assert(result1.Code, Equals, sdk.CodeUnknownRequest)

	msgInvalidPool := NewMsgOutboundTx(GetRandomTxHash(), 1, GetRandomBNBAddress(), w.activeNodeAccount.Accounts.ObserverBEPAddress)
	result2 := handleMsgOutboundTx(w.ctx, w.keeper, w.poolAddrMgr, msgInvalidPool)
	c.Assert(result2.Code, Equals, sdk.CodeUnauthorized)

	w = getHandlerTestWrapper(c, 1, true, true)
	msgOutboundTxNormal := NewMsgOutboundTx(GetRandomTxHash(), 1, w.activeNodeAccount.Accounts.SignerBNBAddress, w.activeNodeAccount.Accounts.ObserverBEPAddress)
	result3 := handleMsgOutboundTx(w.ctx, w.keeper, w.poolAddrMgr, msgOutboundTxNormal)
	c.Assert(result3.Code, Equals, sdk.CodeOK)

	w.txOutStore.NewBlock(2)
	// set a txin
	txIn1 := types.NewTxIn(
		common.Coins{
			common.NewCoin(common.BNBChain, common.RuneA1FTicker, sdk.NewUint(1*common.One)),
		},
		"swap:BNB",
		GetRandomBNBAddress(),
		sdk.NewUint(1024),
		w.activeNodeAccount.Accounts.SignerBNBAddress)
	msgSetTxIn1 := types.NewMsgSetTxIn(
		[]TxInVoter{
			types.NewTxInVoter(GetRandomTxHash(), []TxIn{txIn1}),
		},
		w.activeNodeAccount.Accounts.ObserverBEPAddress)
	ctx := w.ctx.WithBlockHeight(2)
	resultTxIn := handleMsgSetTxIn(ctx, w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, msgSetTxIn1)
	c.Assert(resultTxIn.Code, Equals, sdk.CodeOK)
	w.txOutStore.CommitBlock(ctx)
	msgOutboundTxNormal1 := NewMsgOutboundTx(GetRandomTxHash(), 2, w.activeNodeAccount.Accounts.SignerBNBAddress, w.activeNodeAccount.Accounts.ObserverBEPAddress)
	result4 := handleMsgOutboundTx(ctx, w.keeper, w.poolAddrMgr, msgOutboundTxNormal1)
	c.Assert(result4.Code, Equals, sdk.CodeOK)
}

func (HandlerSuite) TestHandleMsgSetAdminConfig(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)

	msgSetAdminCfg := NewMsgSetAdminConfig(GSLKey, "0.5", w.notActiveNodeAccount.Accounts.ObserverBEPAddress)
	result := handleMsgSetAdminConfig(w.ctx, w.keeper, msgSetAdminCfg)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	msgSetAdminCfg = NewMsgSetAdminConfig(GSLKey, "0.5", w.activeNodeAccount.Accounts.ObserverBEPAddress)
	result1 := handleMsgSetAdminConfig(w.ctx, w.keeper, msgSetAdminCfg)
	c.Assert(result1.Code, Equals, sdk.CodeOK)

	msgInvalidSetAdminCfg := NewMsgSetAdminConfig("Whatever", "blablab", w.activeNodeAccount.Accounts.ObserverBEPAddress)
	result2 := handleMsgSetAdminConfig(w.ctx, w.keeper, msgInvalidSetAdminCfg)
	c.Assert(result2.Code, Equals, sdk.CodeUnknownRequest)
}

func (HandlerSuite) TestHandleMsgAdd(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	msgAdd := NewMsgAdd(common.BNBTicker, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One), GetRandomTxHash(), w.notActiveNodeAccount.Accounts.ObserverBEPAddress)
	result := handleMsgAdd(w.ctx, w.keeper, msgAdd)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	msgInvalidAdd := NewMsgAdd("", sdk.NewUint(100*common.One), sdk.NewUint(100*common.One), GetRandomTxHash(), w.activeNodeAccount.Accounts.ObserverBEPAddress)
	result1 := handleMsgAdd(w.ctx, w.keeper, msgInvalidAdd)
	c.Assert(result1.Code, Equals, sdk.CodeUnknownRequest)

	msgAdd = NewMsgAdd(common.BNBTicker, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One), GetRandomTxHash(), w.activeNodeAccount.Accounts.ObserverBEPAddress)
	result2 := handleMsgAdd(w.ctx, w.keeper, msgAdd)
	c.Assert(result2.Code, Equals, sdk.CodeUnknownRequest)

	pool := w.keeper.GetPool(w.ctx, common.BNBTicker)
	pool.Ticker = common.BNBTicker
	pool.Status = PoolEnabled
	w.keeper.SetPool(w.ctx, pool)
	result3 := handleMsgAdd(w.ctx, w.keeper, msgAdd)
	c.Assert(result3.Code, Equals, sdk.CodeOK)
	pool = w.keeper.GetPool(w.ctx, common.BNBTicker)
	c.Assert(pool.Status, Equals, PoolEnabled)
	c.Assert(pool.BalanceToken.Uint64(), Equals, sdk.NewUint(100*common.One).Uint64())
	c.Assert(pool.BalanceRune.Uint64(), Equals, sdk.NewUint(100*common.One).Uint64())
	c.Assert(pool.PoolUnits.Uint64(), Equals, uint64(0))

}
func (HandlerSuite) TestHandleMsgAck(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	txID := GetRandomTxHash()
	sender := GetRandomBNBAddress()
	signer := GetRandomBech32Addr()
	msgAck := NewMsgAck(txID, sender, signer)
	result := handleMsgAck(w.ctx, w.keeper, w.validatorMgr, msgAck)
	c.Assert(result.Code, Equals, sdk.CodeUnknownRequest)
	// nominated a node
	node1 := GetRandomNodeAccount(NodeStandby)
	w.keeper.SetNodeAccount(w.ctx, node1)
	w.validatorMgr.BeginBlock(w.ctx, w.validatorMgr.Meta.RotateWindowOpenAtBlockHeight)
	updates := w.validatorMgr.EndBlock(w.ctx, w.validatorMgr.Meta.RotateWindowOpenAtBlockHeight)
	c.Assert(updates, IsNil)
	c.Assert(w.validatorMgr.Meta.Nominated.IsEmpty(), Equals, false)

	resultDifferentSignerAddr := handleMsgAck(w.ctx, w.keeper, w.validatorMgr, msgAck)
	c.Assert(resultDifferentSignerAddr.Code, Equals, sdk.CodeUnknownRequest)

	nominated := w.validatorMgr.Meta.Nominated
	msgAckNormal := NewMsgAck(GetRandomTxHash(), nominated.Accounts.SignerBNBAddress, w.activeNodeAccount.Accounts.ObserverBEPAddress)
	resultNormal := handleMsgAck(w.ctx, w.keeper, w.validatorMgr, msgAckNormal)
	c.Assert(resultNormal.Code, Equals, sdk.CodeOK)

	nominated.ObserverActive = true
	w.keeper.SetNodeAccount(w.ctx, nominated)
	msgAckNormal1 := NewMsgAck(GetRandomTxHash(), nominated.Accounts.SignerBNBAddress, w.activeNodeAccount.Accounts.ObserverBEPAddress)
	resultNormal1 := handleMsgAck(w.ctx, w.keeper, w.validatorMgr, msgAckNormal1)
	c.Assert(resultNormal1.Code, Equals, sdk.CodeOK)
	n, err := w.keeper.GetNodeAccount(w.ctx, nominated.NodeAddress)
	c.Assert(err, IsNil)
	c.Assert(n.Status, Equals, NodeReady)
}

func (HandlerSuite) TestRefund(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)

	pool := Pool{
		Ticker:       common.BNBTicker,
		BalanceRune:  sdk.NewUint(100 * common.One),
		BalanceToken: sdk.NewUint(100 * common.One),
	}
	w.keeper.SetPool(w.ctx, pool)

	// test we create a refund transaction
	txin := TxIn{
		Sender: GetRandomBNBAddress(),
		Coins: common.Coins{
			common.NewCoin(common.BNBChain, common.BNBTicker, sdk.NewUint(100*common.One)),
		},
	}
	currentPoolAddr := w.poolAddrMgr.GetCurrentPoolAddresses().Current
	refundTx(w.ctx, txin, w.txOutStore, w.keeper, currentPoolAddr, true)
	c.Assert(w.txOutStore.GetOutboundItems(), HasLen, 1)

	// check we DONT create a refund transaction when we don't have a pool for
	// the asset sent.
	txin = TxIn{
		Sender: GetRandomBNBAddress(),
		Coins: common.Coins{
			common.NewCoin(common.BNBChain, "LOKI", sdk.NewUint(100*common.One)),
		},
	}

	refundTx(w.ctx, txin, w.txOutStore, w.keeper, currentPoolAddr, true)
	c.Assert(w.txOutStore.GetOutboundItems(), HasLen, 1)
	pool = w.keeper.GetPool(w.ctx, "LOKI")
	c.Assert(pool.BalanceToken.Equal(sdk.NewUint(100*common.One)), Equals, true)

	// doing it a second time should add the tokens again.
	refundTx(w.ctx, txin, w.txOutStore, w.keeper, currentPoolAddr, true)
	c.Assert(w.txOutStore.GetOutboundItems(), HasLen, 1)
	pool = w.keeper.GetPool(w.ctx, "LOKI")
	c.Assert(pool.BalanceToken.Equal(sdk.NewUint(200*common.One)), Equals, true)
}
