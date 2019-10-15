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
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/statechain/cmd"
	"gitlab.com/thorchain/bepswap/statechain/x/swapservice/types"
)

type HandlerSuite struct{}

var _ = Suite(&HandlerSuite{})

func (s *HandlerSuite) SetUpSuite(c *C) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
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
		supply.Minter:         []string{supply.Minter},
		supply.Burner:         []string{supply.Burner},
		multiPerm:             []string{supply.Minter, supply.Burner, supply.Staking},
		randomPerm:            []string{"random"},
		ModuleName:            {supply.Minter},
	}
	supplyKeeper := supply.NewKeeper(cdc, keySupply, ak, bk, maccPerms)
	totalSupply := sdk.NewCoins(sdk.NewCoin("bep", sdk.NewInt(1000*common.One)))
	supplyKeeper.SetSupply(ctx, supply.NewSupply(totalSupply))
	k := NewKeeper(bk, supplyKeeper, keySwapService, cdc)
	return ctx, k
}

func (HandlerSuite) TestHandleMsgApply(c *C) {
	ctx, k := setupKeeperForTest(c)
	nodeAddr, err := sdk.AccAddressFromBech32("bep1munaxncdrr305vf0ljzydyqryncxx8cp06v65u")
	c.Check(err, IsNil)
	bond := sdk.NewUint(100)
	txID, err := common.NewTxID("33913271624A28F9EF4F53CA90BEABADF196E791EA2C281740159FEBE161B620")
	c.Check(err, IsNil)
	signer, err := sdk.AccAddressFromBech32("bep1a8amdesvfjtxamagh6hmrkcrej73rzas2jny7z")
	c.Check(err, IsNil)

	// Not Authorized
	msgApply := NewMsgApply(nodeAddr, bond, txID, signer)
	c.Assert(msgApply.ValidateBasic(), IsNil)
	result := handleMsgApply(ctx, k, msgApply)
	c.Assert(result.IsOK(), Equals, false)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	// add observer
	bnb, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)
	bepConsPubKey := `bepcpub1zcjduepq4kn64fcjhf0fp20gp8var0rm25ca9jy6jz7acem8gckh0nkplznq85gdrg`
	nodeAccount := NewNodeAccount(signer, NodeActive, NewTrustAccount(bnb, signer, bepConsPubKey))
	k.SetNodeAccount(ctx, nodeAccount)
	// nodeAccoutn already exist
	result = handleMsgApply(ctx, k, msgApply)
	c.Assert(result.IsOK(), Equals, false)
	c.Assert(result.Code, Equals, sdk.CodeUnknownRequest)

	// invalid Msg
	invalidMsgApply := NewMsgApply(sdk.AccAddress{}, bond, txID, signer)
	invalidMsgApplyResult := handleMsgApply(ctx, k, invalidMsgApply)
	c.Assert(invalidMsgApplyResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Assert(invalidMsgApplyResult.IsOK(), Equals, false)

	// less than minimum bond
	msgApplyLessThanMinimumBond := NewMsgApply(nodeAddr, sdk.NewUint(1000), txID, signer)
	lessThanMinimumBondResult := handleMsgApply(ctx, k, msgApplyLessThanMinimumBond)
	c.Assert(lessThanMinimumBondResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Assert(lessThanMinimumBondResult.IsOK(), Equals, false)

	msgApply1 := NewMsgApply(nodeAddr, sdk.NewUint(100*common.One), txID, signer)
	result = handleMsgApply(ctx, k, msgApply1)
	c.Assert(result.IsOK(), Equals, true)
	c.Assert(result.Code, Equals, sdk.CodeOK)
	coins := k.coinKeeper.GetCoins(ctx, nodeAddr)
	c.Assert(coins.AmountOf("bep").Int64(), Equals, int64(1000))

	// apply again shohuld fail
	resultExist := handleMsgApply(ctx, k, msgApply1)
	c.Assert(resultExist.IsOK(), Equals, false)
	c.Assert(resultExist.Code, Equals, sdk.CodeUnknownRequest)

}

func (HandlerSuite) TestHandleMsgSetTrustAccount(c *C) {
	ctx, k := setupKeeperForTest(c)
	nodeAddr, err := sdk.AccAddressFromBech32("bep1munaxncdrr305vf0ljzydyqryncxx8cp06v65u")
	c.Check(err, IsNil)

	signer, err := sdk.AccAddressFromBech32("bep1a8amdesvfjtxamagh6hmrkcrej73rzas2jny7z")
	c.Check(err, IsNil)
	// add observer
	bnb, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)
	bepConsPubKey := `bepcpub1zcjduepq4kn64fcjhf0fp20gp8var0rm25ca9jy6jz7acem8gckh0nkplznq85gdrg`

	trustAccount := NewTrustAccount(bnb, signer, bepConsPubKey)
	msgTrustAccount := types.NewMsgSetTrustAccount(trustAccount, signer)
	unAuthorizedResult := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(unAuthorizedResult.Code, Equals, sdk.CodeUnauthorized)
	c.Check(unAuthorizedResult.IsOK(), Equals, false)
	nodeAccount := NewNodeAccount(signer, NodeActive, NewTrustAccount(common.NoBnbAddress, sdk.AccAddress{}, ""))
	k.SetNodeAccount(ctx, nodeAccount)

	activeFailResult := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(activeFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(activeFailResult.IsOK(), Equals, false)

	nodeAccount = NewNodeAccount(signer, NodeDisabled, NewTrustAccount(common.NoBnbAddress, sdk.AccAddress{}, ""))
	k.SetNodeAccount(ctx, nodeAccount)

	disabledFailResult := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(disabledFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(disabledFailResult.IsOK(), Equals, false)

	k.SetNodeAccount(ctx, NewNodeAccount(signer, NodeWhiteListed, NewTrustAccount(bnb, nodeAddr, bepConsPubKey)))

	notUniqueFailResult := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(notUniqueFailResult.Code, Equals, sdk.CodeUnknownRequest)
	c.Check(notUniqueFailResult.IsOK(), Equals, false)

	nodeAccount = NewNodeAccount(signer, NodeWhiteListed, NewTrustAccount(common.NoBnbAddress, sdk.AccAddress{}, ""))
	k.SetNodeAccount(ctx, nodeAccount)

	success := handleMsgSetTrustAccount(ctx, k, msgTrustAccount)
	c.Check(success.Code, Equals, sdk.CodeOK)
	c.Check(success.IsOK(), Equals, true)

}

func (HandlerSuite) TestIsSignedByActiveObserver(c *C) {
	ctx, k := setupKeeperForTest(c)
	nodeAddr, err := sdk.AccAddressFromBech32("bep1munaxncdrr305vf0ljzydyqryncxx8cp06v65u")
	c.Check(err, IsNil)
	c.Check(isSignedByActiveObserver(ctx, k, []sdk.AccAddress{nodeAddr}), Equals, false)
	c.Check(isSignedByActiveObserver(ctx, k, []sdk.AccAddress{}), Equals, false)
}

func (HandlerSuite) TestIsSignedByActiveNodeAccounts(c *C) {
	ctx, k := setupKeeperForTest(c)
	nodeAddr, err := sdk.AccAddressFromBech32("bep1munaxncdrr305vf0ljzydyqryncxx8cp06v65u")
	c.Check(err, IsNil)
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
	ctx, k := setupKeeperForTest(c)
	acc1 := GetRandomNodeAccount(NodeWhiteListed)
	msgSetPoolData := NewMsgSetPoolData(common.BNBTicker, PoolEnabled, acc1.NodeAddress)
	result := handleMsgSetPoolData(ctx, k, msgSetPoolData)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)
	acc1.UpdateStatus(NodeActive)
	k.SetNodeAccount(ctx, acc1)

	result1 := handleMsgSetPoolData(ctx, k, msgSetPoolData)
	c.Assert(result1.Code, Equals, sdk.CodeOK)

	msgSetPoolData1 := NewMsgSetPoolData("", PoolEnabled, acc1.NodeAddress)
	result2 := handleMsgSetPoolData(ctx, k, msgSetPoolData1)
	c.Assert(result2.Code, Equals, sdk.CodeUnknownRequest)
}

func (HandlerSuite) TestHandleMsgSetStakeData(c *C) {
	ctx, k := setupKeeperForTest(c)
	acc1 := GetRandomNodeAccount(NodeWhiteListed)
	bnbAddr := GetRandomBNBAddress()
	stakeTxHash := GetRandomTxHash()
	msgSetStake := NewMsgSetStakeData(common.BNBTicker,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		bnbAddr,
		stakeTxHash,
		acc1.Accounts.ObserverBEPAddress)
	stakeResult := handleMsgSetStakeData(ctx, k, msgSetStake)
	c.Assert(stakeResult.Code, Equals, sdk.CodeUnauthorized)
	p := k.GetPool(ctx, common.BNBTicker)
	c.Assert(p.Empty(), Equals, true)
	acc1.UpdateStatus(NodeActive)
	k.SetNodeAccount(ctx, acc1)
	stakeResult1 := handleMsgSetStakeData(ctx, k, msgSetStake)
	c.Assert(stakeResult1.Code, Equals, sdk.CodeOK)
	p = k.GetPool(ctx, common.BNBTicker)
	c.Assert(p.Empty(), Equals, false)
	c.Assert(p.BalanceRune.Uint64(), Equals, msgSetStake.RuneAmount.Uint64())
	c.Assert(p.BalanceToken.Uint64(), Equals, msgSetStake.TokenAmount.Uint64())
	e, err := k.GetCompletedEvent(ctx, 1)
	c.Assert(err, IsNil)
	c.Assert(e.Pool.Equals(common.BNBTicker), Equals, true)
	c.Assert(e.Status.Valid(), IsNil)
	c.Assert(e.InHash.Equals(stakeTxHash), Equals, true)

	// Suspended pool should not allow stake
	k.SetPoolData(ctx, common.BNBTicker, PoolSuspended)

	msgSetStake1 := NewMsgSetStakeData(common.BNBTicker,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		GetRandomBNBAddress(),
		GetRandomTxHash(),
		acc1.Accounts.ObserverBEPAddress)
	stakeResult2 := handleMsgSetStakeData(ctx, k, msgSetStake1)
	c.Assert(stakeResult2.Code, Equals, sdk.CodeUnknownRequest)

	msgSetStake2 := NewMsgSetStakeData(common.BNBTicker,
		sdk.NewUint(100*common.One),
		sdk.NewUint(100*common.One),
		"",
		"",
		acc1.Accounts.ObserverBEPAddress)
	stakeResult3 := handleMsgSetStakeData(ctx, k, msgSetStake2)
	c.Assert(stakeResult3.Code, Equals, sdk.CodeUnknownRequest)
}

func (HandlerSuite) TestHandleMsgConfirmNextPoolAddress(c *C) {
	ctx, k := setupKeeperForTest(c)
	acc1 := GetRandomNodeAccount(NodeActive)
	k.SetNodeAccount(ctx, acc1)
	validatorMgr := NewValidatorManager(k)
	validatorMgr.BeginBlock(ctx, 1)
	msgNextPoolAddr := NewMsgNextPoolAddress(
		GetRandomTxHash(),
		GetRandomBNBAddress(),
		acc1.Accounts.ObserverBEPAddress)
	result := handleMsgConfirmNextPoolAddress(ctx, k, validatorMgr, msgNextPoolAddr)
	c.Assert(result.Code, Equals, sdk.CodeUnknownRequest)

	acc2 := GetRandomNodeAccount(NodeStandby)
	k.SetNodeAccount(ctx, acc2)
	updates := validatorMgr.EndBlock(ctx, validatorMgr.Meta.RotateWindowOpenAtBlockHeight)
	c.Assert(updates, IsNil)
	// we nominated a node account
	c.Assert(validatorMgr.Meta.Nominated.IsEmpty(), Equals, false)
	msgNextPoolAddr1 := NewMsgNextPoolAddress(
		GetRandomTxHash(),
		acc2.Accounts.SignerBNBAddress,
		acc2.Accounts.ObserverBEPAddress)
	result1 := handleMsgConfirmNextPoolAddress(ctx, k, validatorMgr, msgNextPoolAddr1)
	c.Assert(result1.Code, Equals, sdk.CodeOK)
	acc2, err := k.GetNodeAccount(ctx, acc2.NodeAddress)
	c.Assert(err, IsNil)
	c.Assert(acc2.Status, Equals, NodeStandby)

}

func (HandlerSuite) TestHandleMsgSetTxIn(c *C) {
	ctx, k := setupKeeperForTest(c)
	acc1 := GetRandomNodeAccount(NodeActive)
	txOutStore := NewTxOutStore(k)
	validatorMgr := NewValidatorManager(k)
	validatorMgr.BeginBlock(ctx, 1)
	poolAddrMgr := NewPoolAddressManager(k)
	poolAddrMgr.BeginBlock(ctx, 1)
	txIn := types.NewTxIn(
		common.Coins{
			common.NewCoin(common.BNBTicker, sdk.NewUint(100*common.One)),
			common.NewCoin(common.RuneA1FTicker, sdk.NewUint(100*common.One)),
		},
		"stake:BNB",
		GetRandomBNBAddress(),
		sdk.NewUint(1024),
		GetRandomBNBAddress())

	msgSetTxIn := types.NewMsgSetTxIn(
		[]TxInVoter{
			types.NewTxInVoter(GetRandomTxHash(), []TxIn{txIn}),
		},
		acc1.NodeAddress)
	txOutStore.NewBlock(1)
	result := handleMsgSetTxIn(ctx, k, txOutStore, poolAddrMgr, validatorMgr, msgSetTxIn)
	c.Assert(result.Code, Equals, sdk.CodeUnauthorized)

	// set one active account
	k.SetNodeAccount(ctx, acc1)

	// create a pool so we have pool price
	p := k.GetPool(ctx, common.BNBTicker)
	p.BalanceToken = sdk.NewUint(100 * common.One)
	p.BalanceRune = sdk.NewUint(100 * common.One)
	p.Ticker = common.BNBTicker
	p.Status = PoolEnabled
	k.SetPool(ctx, p)

	poolAddrMgr.BeginBlock(ctx, 1)
	validatorMgr.BeginBlock(ctx, 1)
	txOutStore.NewBlock(1)
	// send to wrong pool address, refund
	result1 := handleMsgSetTxIn(ctx, k, txOutStore, poolAddrMgr, validatorMgr, msgSetTxIn)
	c.Assert(result1.Code, Equals, sdk.CodeOK)
	c.Assert(txOutStore.blockOut, NotNil)
	c.Assert(txOutStore.blockOut.Valid(), IsNil)
	c.Assert(txOutStore.blockOut.IsEmpty(), Equals, false)
	c.Assert(len(txOutStore.blockOut.TxArray), Equals, 1)
	// expect to refund two coins
	c.Assert(len(txOutStore.blockOut.TxArray[0].Coins), Equals, 2)

	txIn1 := types.NewTxIn(
		common.Coins{
			common.NewCoin(common.BNBTicker, sdk.NewUint(100*common.One)),
			common.NewCoin(common.RuneA1FTicker, sdk.NewUint(100*common.One)),
		},
		"stake:BNB",
		GetRandomBNBAddress(),
		sdk.NewUint(1024),
		acc1.Accounts.SignerBNBAddress)
	msgSetTxIn1 := types.NewMsgSetTxIn(
		[]TxInVoter{
			types.NewTxInVoter(GetRandomTxHash(), []TxIn{txIn1}),
		},
		acc1.NodeAddress)
	result2 := handleMsgSetTxIn(ctx, k, txOutStore, poolAddrMgr, validatorMgr, msgSetTxIn1)
	c.Assert(result2.Code, Equals, sdk.CodeOK)
	p1 := k.GetPool(ctx, common.BNBTicker)
	c.Assert(p1.BalanceRune.Uint64(), Equals, uint64(200*common.One))
	c.Assert(p1.BalanceRune.Uint64(), Equals, uint64(200*common.One))
	// pool staker
	ps, err := k.GetPoolStaker(ctx, common.BNBTicker)
	c.Assert(err, IsNil)
	c.Assert(ps.TotalUnits.GT(sdk.ZeroUint()), Equals, true)

}
