package swapservice

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/pkg/errors"
	. "gopkg.in/check.v1"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"gitlab.com/thorchain/bepswap/common"

	"gitlab.com/thorchain/bepswap/statechain/cmd"
	"gitlab.com/thorchain/bepswap/statechain/x/swapservice/mocks"
	"gitlab.com/thorchain/bepswap/statechain/x/swapservice/types"
)

type SwapSuite struct{}

var _ = Suite(&SwapSuite{})
var keyStore = sdk.NewKVStoreKey(StoreKey)

func (s *SwapSuite) SetUpSuite(c *C) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
}

func GetCtx() sdk.Context {

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(keyStore, sdk.StoreTypeIAVL, db)
	if err := cms.LoadLatestVersion(); nil != err {
		fmt.Printf("error load latest db version error: %s ", err)
	}
	return sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())

}

func (s SwapSuite) TestSwap(c *C) {
	poolStorage := mocks.MockPoolStorage{}
	ctx := GetCtx()
	tradeSlipLimit := common.Amount("0.100000")
	globalSlipLimit := common.Amount("0.200000")
	inputs := []struct {
		name            string
		requestTxHash   common.TxID
		source          common.Ticker
		target          common.Ticker
		amount          sdk.Uint
		requester       common.BnbAddress
		destination     common.BnbAddress
		returnAmount    sdk.Uint
		tradeTarget     sdk.Uint
		tradeSlipLimit  common.Amount
		globalSlipLimit common.Amount
		expectedErr     error
	}{
		{
			name:          "empty-source",
			requestTxHash: "hash",
			source:        common.Ticker(""),
			target:        common.BNBTicker,
			amount:        sdk.NewUint(100 * common.One),
			requester:     "tester",
			destination:   "whatever",
			returnAmount:  sdk.ZeroUint(),
			expectedErr:   errors.New("source is empty"),
		},
		{
			name:          "empty-target",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        common.Ticker(""),
			amount:        sdk.NewUint(100 * common.One),
			requester:     "tester",
			destination:   "whatever",
			returnAmount:  sdk.ZeroUint(),
			expectedErr:   errors.New("target is empty"),
		},
		{
			name:          "empty-requestTxHash",
			requestTxHash: "",
			source:        common.RuneTicker,
			target:        common.BNBTicker,
			amount:        sdk.NewUint(100 * common.One),
			requester:     "tester",
			destination:   "whatever",
			returnAmount:  sdk.ZeroUint(),
			expectedErr:   errors.New("request tx hash is empty"),
		},
		{
			name:          "empty-amount",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        common.BNBTicker,
			amount:        sdk.ZeroUint(),
			requester:     "tester",
			destination:   "whatever",
			returnAmount:  sdk.ZeroUint(),
			expectedErr:   errors.New("amount is zero"),
		},
		{
			name:          "empty-requester",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        common.BNBTicker,
			amount:        sdk.NewUint(100 * common.One),
			requester:     "",
			destination:   "whatever",
			returnAmount:  sdk.ZeroUint(),
			expectedErr:   errors.New("requester is empty"),
		},
		{
			name:          "empty-destination",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        common.BNBTicker,
			amount:        sdk.NewUint(100 * common.One),
			requester:     "tester",
			destination:   "",
			returnAmount:  sdk.ZeroUint(),
			expectedErr:   errors.New("destination is empty"),
		},
		{
			name:          "pool-not-exist",
			requestTxHash: "hash",
			source:        "NOTEXIST",
			target:        common.RuneTicker,
			amount:        sdk.NewUint(100 * common.One),
			requester:     "tester",
			destination:   "don'tknow",
			tradeTarget:   sdk.NewUint(110000000),
			returnAmount:  sdk.ZeroUint(),
			expectedErr:   errors.New("NOTEXIST doesn't exist"),
		},
		{
			name:          "pool-not-exist-1",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        "NOTEXIST",
			amount:        sdk.NewUint(100 * common.One),
			requester:     "tester",
			destination:   "don'tknow",
			tradeTarget:   sdk.NewUint(120000000),
			returnAmount:  sdk.ZeroUint(),
			expectedErr:   errors.New("NOTEXIST doesn't exist"),
		},
		{
			name:          "swap-over-global-sliplimit",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        common.BNBTicker,
			amount:        sdk.NewUint(50 * common.One),
			requester:     "tester",
			destination:   "don't know",
			returnAmount:  sdk.ZeroUint(),
			tradeTarget:   sdk.ZeroUint(),
			expectedErr:   errors.Errorf("pool slip:0.928571 is over global pool slip limit :%s", globalSlipLimit),
		},
		{
			name:          "swap-over-trade-sliplimit",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        common.BNBTicker,
			amount:        sdk.NewUint(9 * common.One),
			requester:     "tester",
			destination:   "don'tknow",
			returnAmount:  sdk.ZeroUint(),
			tradeTarget:   sdk.NewUint(common.One),
			expectedErr:   errors.New("trade slip 1.188100 is more than 10.00 percent different than 1.000000"),
		},
		{
			name:          "swap-no-target-price-no-protection",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        common.BNBTicker,
			amount:        sdk.NewUint(8 * common.One),
			requester:     "tester",
			destination:   "don'tknow",
			returnAmount:  sdk.NewUint(685871056),
			tradeTarget:   sdk.ZeroUint(),
			expectedErr:   nil,
		},
		{
			name:          "swap",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        common.BNBTicker,
			amount:        sdk.NewUint(5 * common.One),
			requester:     "tester",
			destination:   "don'tknow",
			returnAmount:  sdk.NewUint(453514739),
			tradeTarget:   sdk.NewUint(120000000),
			expectedErr:   nil,
		},
		{
			name:          "double-swap",
			requestTxHash: "hash",
			source:        "BTC",
			target:        common.BNBTicker,
			amount:        sdk.NewUint(5 * common.One),
			requester:     "tester",
			destination:   "don'tknow",
			returnAmount:  sdk.NewUint(415017809),
			tradeTarget:   sdk.NewUint(110250000),
			expectedErr:   nil,
		},
	}
	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	for _, item := range inputs {
		c.Logf("test name:%s", item.name)
		amount, err := swap(ctx, poolStorage, txID, item.source, item.target, item.amount, item.requester, item.destination, item.requestTxHash, item.tradeTarget, tradeSlipLimit, globalSlipLimit)
		if item.expectedErr == nil {
			c.Assert(err, IsNil)
		} else {
			c.Assert(err, NotNil, Commentf("Expected: %s, got nil", item.expectedErr.Error()))
			c.Assert(err.Error(), Equals, item.expectedErr.Error())
		}
		c.Logf("expected amount:%s,actual amount:%s", item.returnAmount, amount)
		c.Check(item.returnAmount.Uint64(), Equals, amount.Uint64())
	}
}

func (s SwapSuite) TestValidatePools(c *C) {
	keeper := mocks.MockPoolStorage{}
	ctx := GetCtx()
	c.Check(validatePools(ctx, keeper, common.RuneTicker), IsNil)
	c.Check(validatePools(ctx, keeper, "NOTEXIST"), NotNil)
}

func (s SwapSuite) TestValidateMessage(c *C) {
	c.Check(validateMessage(common.RuneTicker, "BNB", sdk.NewUint(3429850000), "bnbXXXX", "bnbYYY", "txHASH"), IsNil)
	c.Check(validateMessage(common.RuneTicker, "BNB", sdk.NewUint(3429850000), "bnbXXXX", "bnbYYY", ""), NotNil)
	c.Check(validateMessage("", "BNB", sdk.NewUint(3429850000), "bnbXXXX", "bnbYYY", "txHASH"), NotNil)
	c.Check(validateMessage(common.RuneTicker, "", sdk.NewUint(3429850000), "bnbXXXX", "bnbYYY", "txHASH"), NotNil)
	c.Check(validateMessage(common.RuneTicker, "BNB", sdk.ZeroUint(), "bnbXXXX", "bnbYYY", "txHASH"), NotNil)
	c.Check(validateMessage(common.RuneTicker, "BNB", sdk.NewUint(3429850000), "", "bnbYYY", "txHASH"), NotNil)
	c.Check(validateMessage(common.RuneTicker, "BNB", sdk.NewUint(3429850000), "bnbXXXX", "", "txHASH"), NotNil)
}

func (s SwapSuite) TestCalculators(c *C) {
	X := sdk.NewUint(100 * common.One)
	x := sdk.NewUint(10 * common.One)
	Y := sdk.NewUint(100 * common.One)

	// These calculations are verified by using the spreadsheet
	// https://docs.google.com/spreadsheets/d/1wJHYBRKBdw_WP7nUyVnkySPkOmPUNoiRGsEqgBVVXKU/edit#gid=0
	fmt.Println("poolslip", calcPoolSlip(X, x))
	c.Check(calcTokenEmission(X, x, Y).Uint64(), Equals, uint64(826446280))
	c.Check(calcLiquitityFee(X, x, Y).Uint64(), Equals, uint64(82644628))
	c.Check(calcPoolSlip(X, x), Equals, 0.1990990990990991)
	c.Check(calcTradeSlip(X, x), Equals, 0.21)
	c.Check(calcPriceSlip(X, x, Y), Equals, 1.210000001452)
	c.Check(calcOutputSlip(X, x), Equals, 0.09090909090909091)
}

func (s SwapSuite) TestHandleMsgSwap(c *C) {
	ctx := GetCtx()
	var cdc = codec.New()
	RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	tkeys := sdk.NewTransientStoreKeys(staking.TStoreKey, params.TStoreKey)
	paramsKeeper := params.NewKeeper(cdc, keyStore, tkeys[params.TStoreKey], params.DefaultCodespace)
	// Set specific supspaces
	authSubspace := paramsKeeper.Subspace(auth.DefaultParamspace)
	bankSupspace := paramsKeeper.Subspace(bank.DefaultParamspace)

	// The AccountKeeper handles address -> account lookups
	accountKeeper := auth.NewAccountKeeper(
		cdc,
		keyStore,
		authSubspace,
		auth.ProtoBaseAccount,
	)

	// The BankKeeper allows you perform sdk.Coins interactions
	bankKeeper := bank.NewBaseKeeper(
		accountKeeper,
		bankSupspace,
		bank.DefaultCodespace,
		nil, // app.ModuleAccountAddrs(),
	)

	k := NewKeeper(bankKeeper, keyStore, cdc)

	txOutStore := NewTxOutStore(k)
	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	addr, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)
	signerAddr, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	k.SetTrustAccount(ctx, types.NewTrustAccount(addr, addr, signerAddr))
	txOutStore.NewBlock(1)
	// no pool
	msg := NewMsgSwap(txID, common.RuneA1FTicker, common.BNBTicker, sdk.NewUint(common.One), addr, addr, sdk.ZeroUint(), signerAddr)
	res := handleMsgSwap(ctx, k, txOutStore, msg)
	c.Assert(res.Code, Equals, sdk.CodeInternal)
	pool := NewPool()
	pool.Ticker = common.BNBTicker
	pool.BalanceToken = sdk.NewUint(100 * common.One)
	pool.BalanceRune = sdk.NewUint(100 * common.One)
	k.SetPool(ctx, pool)

	res = handleMsgSwap(ctx, k, txOutStore, msg)
	c.Assert(res.IsOK(), Equals, true)
}
