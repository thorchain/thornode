package swapservice

import (
	. "gopkg.in/check.v1"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"gitlab.com/thorchain/bepswap/common"
	"gitlab.com/thorchain/bepswap/statechain/x/swapservice/mocks"
)

type SwapSuite struct{}

var _ = Suite(&SwapSuite{})

func GetCtx(key string) sdk.Context {
	keystore := sdk.NewKVStoreKey(key)
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(keystore, sdk.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	return sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())

}

func (s SwapSuite) TestSwap(c *C) {
	poolStorage := mocks.MockPoolStorage{}
	ctx := GetCtx("test")
	tradeSlipLimit := common.Amount("0.100000")
	globalSlipLimit := common.Amount("0.200000")
	inputs := []struct {
		name            string
		requestTxHash   common.TxID
		source          common.Ticker
		target          common.Ticker
		amount          common.Amount
		requester       common.BnbAddress
		destination     common.BnbAddress
		returnAmount    common.Amount
		tradeTarget     common.Amount
		tradeSlipLimit  common.Amount
		globalSlipLimit common.Amount
		expectedErr     error
	}{
		{
			name:          "empty-source",
			requestTxHash: "hash",
			source:        common.Ticker(""),
			target:        common.BNBTicker,
			amount:        "100",
			requester:     "tester",
			destination:   "whatever",
			returnAmount:  "0",
			expectedErr:   errors.New("source is empty"),
		},
		{
			name:          "empty-target",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        common.Ticker(""),
			amount:        "100",
			requester:     "tester",
			destination:   "whatever",
			returnAmount:  "0",
			expectedErr:   errors.New("target is empty"),
		},
		{
			name:          "empty-requestTxHash",
			requestTxHash: "",
			source:        common.RuneTicker,
			target:        common.BNBTicker,
			amount:        "100",
			requester:     "tester",
			destination:   "whatever",
			returnAmount:  "0",
			expectedErr:   errors.New("request tx hash is empty"),
		},
		{
			name:          "empty-amount",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        common.BNBTicker,
			amount:        "",
			requester:     "tester",
			destination:   "whatever",
			returnAmount:  "0",
			expectedErr:   errors.New("amount is empty"),
		},
		{
			name:          "empty-requester",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        common.BNBTicker,
			amount:        "100",
			requester:     "",
			destination:   "whatever",
			returnAmount:  "0",
			expectedErr:   errors.New("requester is empty"),
		},
		{
			name:          "empty-destination",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        common.BNBTicker,
			amount:        "100",
			requester:     "tester",
			destination:   "",
			returnAmount:  "0",
			expectedErr:   errors.New("destination is empty"),
		},
		{
			name:          "pool-not-exist",
			requestTxHash: "hash",
			source:        "NOTEXIST",
			target:        common.RuneTicker,
			amount:        "100",
			requester:     "tester",
			destination:   "don'tknow",
			tradeTarget:   "1.1",
			returnAmount:  "0",
			expectedErr:   errors.New("NOTEXIST doesn't exist"),
		},
		{
			name:          "pool-not-exist-1",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        "NOTEXIST",
			amount:        "100",
			requester:     "tester",
			destination:   "don'tknow",
			tradeTarget:   "1.2",
			returnAmount:  "0",
			expectedErr:   errors.New("NOTEXIST doesn't exist"),
		},
		{
			name:          "swap-over-global-sliplimit",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        common.BNBTicker,
			amount:        "50",
			requester:     "tester",
			destination:   "don't know",
			returnAmount:  "0",
			tradeTarget:   "0",
			expectedErr:   errors.Errorf("pool slip:0.928571 is over global pool slip limit :%s", globalSlipLimit),
		},
		{
			name:          "swap-over-trade-sliplimit",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        common.BNBTicker,
			amount:        "9",
			requester:     "tester",
			destination:   "don'tknow",
			returnAmount:  "0",
			tradeTarget:   "1.0",
			expectedErr:   errors.New("trade slip 1.188100 is more than 10.00 percent different than 1.000000"),
		},
		{
			name:          "swap-no-target-price-no-protection",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        common.BNBTicker,
			amount:        "8",
			requester:     "tester",
			destination:   "don'tknow",
			returnAmount:  "6.85871056",
			tradeTarget:   "0",
			expectedErr:   nil,
		},
		{
			name:          "swap",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        common.BNBTicker,
			amount:        "5",
			requester:     "tester",
			destination:   "don'tknow",
			returnAmount:  "4.53514739",
			tradeTarget:   "1.2",
			expectedErr:   nil,
		},
		{
			name:          "double-swap",
			requestTxHash: "hash",
			source:        "BTC",
			target:        common.BNBTicker,
			amount:        "5",
			requester:     "tester",
			destination:   "don'tknow",
			returnAmount:  "4.15017810",
			tradeTarget:   "1.1025",
			expectedErr:   nil,
		},
	}
	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	for _, item := range inputs {
		amount, err := swap(ctx, poolStorage, txID, item.source, item.target, item.amount, item.requester, item.destination, item.requestTxHash, item.tradeTarget, tradeSlipLimit, globalSlipLimit)
		if item.expectedErr == nil {
			c.Assert(err, IsNil)
		} else {
			c.Assert(err, NotNil, Commentf("Expected: %s, got nil", item.expectedErr.Error()))
			c.Assert(err.Error(), Equals, item.expectedErr.Error())
		}
		c.Check(item.returnAmount.Equals(amount), Equals, true)
	}
}

func (s SwapSuite) TestValidatePools(c *C) {
	keeper := mocks.MockPoolStorage{}
	ctx := GetCtx("test")
	c.Check(validatePools(ctx, keeper, common.RuneTicker), IsNil)
	c.Check(validatePools(ctx, keeper, "NOTEXIST"), NotNil)
}

func (s SwapSuite) TestValidateMessage(c *C) {
	c.Check(validateMessage("txHASH", common.RuneTicker, "BNB", "34.2985", "bnbXXXX", "bnbYYY"), IsNil)
	c.Check(validateMessage("", common.RuneTicker, "BNB", "34.2985", "bnbXXXX", "bnbYYY"), NotNil)
	c.Check(validateMessage("txHASH", "", "BNB", "34.2985", "bnbXXXX", "bnbYYY"), NotNil)
	c.Check(validateMessage("txHASH", common.RuneTicker, "", "34.2985", "bnbXXXX", "bnbYYY"), NotNil)
	c.Check(validateMessage("txHASH", common.RuneTicker, "BNB", "", "bnbXXXX", "bnbYYY"), NotNil)
	c.Check(validateMessage("txHASH", common.RuneTicker, "BNB", "34.2985", "", "bnbYYY"), NotNil)
	c.Check(validateMessage("txHASH", common.RuneTicker, "BNB", "34.2985", "bnbXXXX", ""), NotNil)
}

func (s SwapSuite) TestCalculators(c *C) {
	X := 100.0
	x := 10.0
	Y := 100.0

	// These calculations are verified by using the spreadsheet
	// https://docs.google.com/spreadsheets/d/1wJHYBRKBdw_WP7nUyVnkySPkOmPUNoiRGsEqgBVVXKU/edit#gid=0
	c.Check(calcTokenEmission(X, x, Y), Equals, 8.264462809917354)
	c.Check(calcLiquitityFee(X, x, Y), Equals, 0.8264462809917356)
	c.Check(calcPoolSlip(X, x), Equals, 0.1990990990990991)
	c.Check(calcTradeSlip(X, x), Equals, 0.21)
	c.Check(calcPriceSlip(X, x, Y), Equals, 1.2100000000000002)
	c.Check(calcOutputSlip(X, x), Equals, 0.09090909090909091)
}
