package swapservice

import (
	"fmt"

	"github.com/pkg/errors"
	. "gopkg.in/check.v1"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
			amount:        sdk.NewUint(100 * One),
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
			amount:        sdk.NewUint(100 * One),
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
			amount:        sdk.NewUint(100 * One),
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
			amount:        sdk.NewUint(100 * One),
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
			amount:        sdk.NewUint(100 * One),
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
			amount:        sdk.NewUint(100 * One),
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
			amount:        sdk.NewUint(100 * One),
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
			amount:        sdk.NewUint(50 * One),
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
			amount:        sdk.NewUint(9 * One),
			requester:     "tester",
			destination:   "don'tknow",
			returnAmount:  sdk.ZeroUint(),
			tradeTarget:   sdk.NewUint(One),
			expectedErr:   errors.New("trade slip 1.188100 is more than 10.00 percent different than 1.000000"),
		},
		{
			name:          "swap-no-target-price-no-protection",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        common.BNBTicker,
			amount:        sdk.NewUint(8 * One),
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
			amount:        sdk.NewUint(5 * One),
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
			amount:        sdk.NewUint(5 * One),
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
	ctx := GetCtx("test")
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
	X := sdk.NewUint(100 * One)
	x := sdk.NewUint(10 * One)
	Y := sdk.NewUint(100 * One)

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
