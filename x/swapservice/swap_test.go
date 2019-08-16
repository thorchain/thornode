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
	"gitlab.com/thorchain/statechain/x/swapservice/mocks"
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
			target:        common.Ticker("BNB"),
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
			target:        common.Ticker("BNB"),
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
			target:        "BNB",
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
			target:        "BNB",
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
			target:        "BNB",
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
			target:        "BNB",
			amount:        "50",
			requester:     "tester",
			destination:   "don't know",
			returnAmount:  "0",
			tradeTarget:   "0",
			expectedErr:   errors.Errorf("pool slip:1.250000 is over global pool slip limit :%s", globalSlipLimit),
		},
		{
			name:          "swap-over-trade-sliplimit",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        "BNB",
			amount:        "9",
			requester:     "tester",
			destination:   "don'tknow",
			returnAmount:  "0",
			tradeTarget:   "1.0",
			expectedErr:   errors.New("user price 1.188100 is more than 10.00 percent different than 1.000000"),
		},
		{
			name:          "swap-no-target-price-no-protection",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        "BNB",
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
			target:        "BNB",
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
			target:        "BNB",
			amount:        "5",
			requester:     "tester",
			destination:   "don'tknow",
			returnAmount:  "4.15017810",
			tradeTarget:   "1.1025",
			expectedErr:   nil,
		},
	}
	for _, item := range inputs {
		amount, err := swap(ctx, poolStorage, item.source, item.target, item.amount, item.requester, item.destination, item.requestTxHash, item.tradeTarget, tradeSlipLimit, globalSlipLimit)
		if item.expectedErr == nil {
			c.Assert(err, IsNil)
		} else {
			c.Assert(err, NotNil, Commentf("Expected: %s, got nil", item.expectedErr.Error()))
			c.Assert(err.Error(), Equals, item.expectedErr.Error())
		}
		c.Check(item.returnAmount.Equals(amount), Equals, true)
	}
}

// TestCalculatePoolSlip the total pool slip
func (s SwapSuite) TestCalculatePoolSlip(c *C) {
	inputs := []struct {
		name             string
		source           common.Ticker
		runeBalance      float64
		tokenBalance     float64
		swapAmount       float64
		expectedPoolSlip float64
	}{
		{
			name:             "normal",
			source:           common.RuneTicker,
			runeBalance:      100.0,
			tokenBalance:     100.0,
			swapAmount:       5.0,
			expectedPoolSlip: 0.1025,
		},
		{
			name:             "normal-1",
			source:           common.RuneTicker,
			runeBalance:      50.0,
			tokenBalance:     200.0,
			swapAmount:       5.0,
			expectedPoolSlip: 0.21,
		},
		{
			name:             "normal-2",
			source:           "BNB",
			runeBalance:      100.0,
			tokenBalance:     100.0,
			swapAmount:       5.0,
			expectedPoolSlip: 0.1025,
		},
		{
			name:             "normal-3",
			source:           common.RuneTicker,
			runeBalance:      500.0,
			tokenBalance:     200.0,
			swapAmount:       5.0,
			expectedPoolSlip: 0.0201,
		},
	}
	for _, testCase := range inputs {
		result := calculatePoolSlip(testCase.source, testCase.runeBalance, testCase.tokenBalance, testCase.swapAmount)
		c.Check(round(result), Equals, round(testCase.expectedPoolSlip))
	}
}

// TestCalculateUserPrice ensure we calculate trade slip correctly
func (s SwapSuite) TestCalculateUserPrice(c *C) {
	inputs := []struct {
		name              string
		source            common.Ticker
		runeBalance       float64
		tokenBalance      float64
		swapAmount        float64
		expectedUserPrice float64
	}{
		{
			name:              "normal",
			source:            common.RuneTicker,
			runeBalance:       100.0,
			tokenBalance:      100.0,
			swapAmount:        5.0,
			expectedUserPrice: 1.1025,
		},
		{
			name:              "normal-1",
			source:            common.RuneTicker,
			runeBalance:       200.0,
			tokenBalance:      1000.0,
			swapAmount:        5,
			expectedUserPrice: 0.210125,
		},
		{
			name:              "normal-2",
			source:            common.Ticker("BNB"),
			runeBalance:       200.0,
			tokenBalance:      1000.0,
			swapAmount:        5,
			expectedUserPrice: 5.05,
		},
		{
			name:              "normal-3",
			source:            common.RuneTicker,
			runeBalance:       2000.0,
			tokenBalance:      1000.0,
			swapAmount:        50,
			expectedUserPrice: 2.10125,
		},
	}
	for _, testCase := range inputs {
		result := calculateUserPrice(testCase.source, testCase.runeBalance, testCase.tokenBalance, testCase.swapAmount)
		c.Check(round(result), Equals, round(testCase.expectedUserPrice))
	}
}

func (s SwapSuite) TestSwapCalculation(c *C) {
	inputs := []struct {
		name              string
		source            common.Ticker
		runeBalance       float64
		tokenBalance      float64
		amountToSwap      float64
		runeBalanceAfter  float64
		tokenBalanceAfter float64
		amountToReturn    float64
		expectedErr       error
	}{
		{
			name:              "negative-balance-rune",
			source:            common.RuneTicker,
			runeBalance:       -1.0,
			tokenBalance:      100.0,
			amountToSwap:      5.0,
			runeBalanceAfter:  105.0,
			tokenBalanceAfter: 95.46,
			amountToReturn:    4.54,
			expectedErr:       errors.New("invalid balance"),
		},
		{
			name:              "zero-balance-rune",
			source:            common.RuneTicker,
			runeBalance:       0.0,
			tokenBalance:      100.0,
			amountToSwap:      5.0,
			runeBalanceAfter:  105.0,
			tokenBalanceAfter: 95.46,
			amountToReturn:    4.54,
			expectedErr:       errors.New("invalid balance"),
		},
		{
			name:              "negative-balance-token",
			source:            common.RuneTicker,
			runeBalance:       100.0,
			tokenBalance:      -100.0,
			amountToSwap:      5.0,
			runeBalanceAfter:  105.0,
			tokenBalanceAfter: 95.46,
			amountToReturn:    4.54,
			expectedErr:       errors.New("invalid balance"),
		},
		{
			name:              "zero-balance-token",
			source:            common.RuneTicker,
			runeBalance:       100.0,
			tokenBalance:      0.0,
			amountToSwap:      5.0,
			runeBalanceAfter:  105.0,
			tokenBalanceAfter: 95.46,
			amountToReturn:    4.54,
			expectedErr:       errors.New("invalid balance"),
		},
		{
			name:              "negative-amount",
			source:            common.RuneTicker,
			runeBalance:       100.0,
			tokenBalance:      100.0,
			amountToSwap:      -5.0,
			runeBalanceAfter:  105.0,
			tokenBalanceAfter: 95.46,
			amountToReturn:    4.54,
			expectedErr:       errors.New("amount is invalid"),
		},
		{
			name:              "invalid-amount-0",
			source:            common.RuneTicker,
			runeBalance:       100.0,
			tokenBalance:      100.0,
			amountToSwap:      0.0,
			runeBalanceAfter:  105.0,
			tokenBalanceAfter: 95.46,
			amountToReturn:    4.54,
			expectedErr:       errors.New("amount is invalid"),
		},
		{
			name:              "normal-rune",
			source:            common.RuneTicker,
			runeBalance:       100.0,
			tokenBalance:      100.0,
			amountToSwap:      5.0,
			runeBalanceAfter:  105.0,
			tokenBalanceAfter: 95.46,
			amountToReturn:    4.54,
		},
		{
			name:              "normal-rune-1",
			source:            common.RuneTicker,
			runeBalance:       1000.0,
			tokenBalance:      1000.0,
			amountToSwap:      20.0,
			runeBalanceAfter:  1020.0,
			tokenBalanceAfter: 980.78,
			amountToReturn:    19.22,
		},
		{
			name:              "normal-rune-2",
			source:            common.RuneTicker,
			runeBalance:       10000.0,
			tokenBalance:      10000.0,
			amountToSwap:      20.0,
			runeBalanceAfter:  10020.0,
			tokenBalanceAfter: 9980.08,
			amountToReturn:    19.92,
		},
		{
			name:              "normal-token",
			source:            common.Ticker("BNB"),
			runeBalance:       100.0,
			tokenBalance:      100.0,
			amountToSwap:      5.0,
			runeBalanceAfter:  95.46,
			tokenBalanceAfter: 105.0,
			amountToReturn:    4.54,
		},
	}

	for _, item := range inputs {
		r, t, a, err := calculateSwap(item.source, item.runeBalance, item.tokenBalance, item.amountToSwap)
		if item.expectedErr == nil {
			c.Assert(err, IsNil)
			c.Check(round(r), Equals, item.runeBalanceAfter)
			c.Check(round(t), Equals, item.tokenBalanceAfter)
			c.Check(round(a), Equals, item.amountToReturn)
		} else {
			c.Assert(err.Error(), Equals, item.expectedErr.Error())
		}
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
