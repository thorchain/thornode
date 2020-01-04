package thorchain

import (
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type SwapSuite struct{}

var _ = Suite(&SwapSuite{})

func (s *SwapSuite) SetUpSuite(c *C) {
	err := os.Setenv("NET", "other")
	c.Assert(err, IsNil)
	SetupConfigForTest()
}

func (s SwapSuite) TestSwap(c *C) {
	poolStorage := MockPoolStorage{}
	ctx, _ := setupKeeperForTest(c)
	inputs := []struct {
		name          string
		requestTxHash common.TxID
		source        common.Asset
		target        common.Asset
		amount        sdk.Uint
		requester     common.Address
		destination   common.Address
		returnAmount  sdk.Uint
		tradeTarget   sdk.Uint
		expectedErr   error
	}{
		{
			name:          "empty-source",
			requestTxHash: "hash",
			source:        common.Asset{},
			target:        common.BNBAsset,
			amount:        sdk.NewUint(100 * common.One),
			requester:     "tester",
			destination:   "whatever",
			returnAmount:  sdk.ZeroUint(),
			expectedErr:   errors.New("Denom cannot be empty"),
		},
		{
			name:          "empty-target",
			requestTxHash: "hash",
			source:        common.RuneAsset(),
			target:        common.Asset{},
			amount:        sdk.NewUint(100 * common.One),
			requester:     "tester",
			destination:   "whatever",
			returnAmount:  sdk.ZeroUint(),
			expectedErr:   errors.New("target is empty"),
		},
		{
			name:          "empty-requestTxHash",
			requestTxHash: "",
			source:        common.RuneAsset(),
			target:        common.BNBAsset,
			amount:        sdk.NewUint(100 * common.One),
			requester:     "tester",
			destination:   "whatever",
			returnAmount:  sdk.ZeroUint(),
			expectedErr:   errors.New("Tx ID cannot be empty"),
		},
		{
			name:          "empty-amount",
			requestTxHash: "hash",
			source:        common.RuneAsset(),
			target:        common.BNBAsset,
			amount:        sdk.ZeroUint(),
			requester:     "tester",
			destination:   "whatever",
			returnAmount:  sdk.ZeroUint(),
			expectedErr:   errors.New("Amount cannot be zero"),
		},
		{
			name:          "empty-requester",
			requestTxHash: "hash",
			source:        common.RuneAsset(),
			target:        common.BNBAsset,
			amount:        sdk.NewUint(100 * common.One),
			requester:     "",
			destination:   "whatever",
			returnAmount:  sdk.ZeroUint(),
			expectedErr:   errors.New("From address cannot be empty"),
		},
		{
			name:          "empty-destination",
			requestTxHash: "hash",
			source:        common.RuneAsset(),
			target:        common.BNBAsset,
			amount:        sdk.NewUint(100 * common.One),
			requester:     GetRandomBNBAddress(),
			destination:   "",
			returnAmount:  sdk.ZeroUint(),
			expectedErr:   errors.New("To address cannot be empty"),
		},
		{
			name:          "pool-not-exist",
			requestTxHash: "hash",
			source:        common.Asset{Chain: common.BNBChain, Ticker: "NOTEXIST", Symbol: "NOTEXIST"},
			target:        common.RuneAsset(),
			amount:        sdk.NewUint(100 * common.One),
			requester:     GetRandomBNBAddress(),
			destination:   GetRandomBNBAddress(),
			tradeTarget:   sdk.NewUint(110000000),
			returnAmount:  sdk.ZeroUint(),
			expectedErr:   errors.New("BNB.NOTEXIST doesn't exist"),
		},
		{
			name:          "pool-not-exist-1",
			requestTxHash: "hash",
			source:        common.RuneAsset(),
			target:        common.Asset{Chain: common.BNBChain, Ticker: "NOTEXIST", Symbol: "NOTEXIST"},
			amount:        sdk.NewUint(100 * common.One),
			requester:     "tester",
			destination:   "don'tknow",
			tradeTarget:   sdk.NewUint(120000000),
			returnAmount:  sdk.ZeroUint(),
			expectedErr:   errors.New("BNB.NOTEXIST doesn't exist"),
		},
		{
			name:          "swap-no-global-sliplimit",
			requestTxHash: "hash",
			source:        common.RuneAsset(),
			target:        common.BNBAsset,
			amount:        sdk.NewUint(50 * common.One),
			requester:     "tester",
			destination:   "don't know",
			returnAmount:  sdk.NewUint(2222222222),
			tradeTarget:   sdk.ZeroUint(),
			expectedErr:   nil,
		},
		{
			name:          "swap-over-trade-sliplimit",
			requestTxHash: "hash",
			source:        common.RuneAsset(),
			target:        common.BNBAsset,
			amount:        sdk.NewUint(9 * common.One),
			requester:     "tester",
			destination:   "don'tknow",
			returnAmount:  sdk.ZeroUint(),
			tradeTarget:   sdk.NewUint(9 * common.One),
			expectedErr:   errors.New("emit asset 757511993 less than price limit 900000000"),
		},
		{
			name:          "swap-no-target-price-no-protection",
			requestTxHash: "hash",
			source:        common.RuneAsset(),
			target:        common.BNBAsset,
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
			source:        common.RuneAsset(),
			target:        common.BNBAsset,
			amount:        sdk.NewUint(5 * common.One),
			requester:     "tester",
			destination:   "don'tknow",
			returnAmount:  sdk.NewUint(453514739),
			tradeTarget:   sdk.NewUint(453514738),
			expectedErr:   nil,
		},
		{
			name:          "double-swap",
			requestTxHash: "hash",
			source:        common.Asset{Chain: common.BTCChain, Ticker: "BTC", Symbol: "BTC"},
			target:        common.BNBAsset,
			amount:        sdk.NewUint(5 * common.One),
			requester:     "tester",
			destination:   "don'tknow",
			returnAmount:  sdk.NewUint(415017809),
			tradeTarget:   sdk.NewUint(415017809),
			expectedErr:   nil,
		},
	}
	for _, item := range inputs {
		c.Logf("test name:%s", item.name)
		tx := common.NewTx(
			item.requestTxHash,
			item.requester,
			item.destination,
			common.Coins{
				common.NewCoin(item.source, item.amount),
			},
			common.BNBGasFeeSingleton,
			"",
		)
		tx.Chain = common.BNBChain
		amount, swapEvents, err := swap(ctx, poolStorage, tx, item.target, item.destination, item.tradeTarget, sdk.NewUint(1000_000))
		if item.expectedErr == nil {
			c.Assert(err, IsNil)
			c.Assert(len(swapEvents) > 0, Equals, true)
		} else {
			c.Assert(err, NotNil, Commentf("Expected: %s, got nil", item.expectedErr.Error()))
			c.Assert(err.Error(), Equals, item.expectedErr.Error())
		}

		c.Logf("expected amount:%s, actual amount:%s", item.returnAmount, amount)
		c.Check(item.returnAmount.Uint64(), Equals, amount.Uint64())

	}
}

func (s SwapSuite) TestValidatePools(c *C) {
	keeper := MockPoolStorage{}
	ctx, _ := setupKeeperForTest(c)
	c.Check(validatePools(ctx, keeper, common.RuneAsset()), IsNil)
	c.Check(validatePools(ctx, keeper, common.Asset{Chain: common.BNBChain, Ticker: "NOTEXIST", Symbol: "NOTEXIST"}), NotNil)
}

func (s SwapSuite) TestValidateMessage(c *C) {
	c.Check(validateMessage(
		common.NewTx(
			GetRandomTxHash(),
			GetRandomBNBAddress(),
			GetRandomBNBAddress(),
			common.Coins{
				common.NewCoin(common.RuneAsset(), sdk.NewUint(3429850000)),
			},
			common.BNBGasFeeSingleton,
			"",
		),
		common.BNBAsset,
		"bnbYYY",
	), IsNil)
	c.Check(validateMessage(
		common.NewTx(
			"",
			GetRandomBNBAddress(),
			GetRandomBNBAddress(),
			common.Coins{
				common.NewCoin(common.RuneAsset(), sdk.NewUint(3429850000)),
			},
			common.BNBGasFeeSingleton,
			"",
		),
		common.BNBAsset,
		"bnbYYY",
	), NotNil)
	c.Check(validateMessage(
		common.NewTx(
			GetRandomTxHash(),
			GetRandomBNBAddress(),
			GetRandomBNBAddress(),
			common.Coins{
				common.NewCoin(common.Asset{}, sdk.NewUint(3429850000)),
			},
			common.BNBGasFeeSingleton,
			"",
		),
		common.BNBAsset,
		"bnbYYY",
	), NotNil)
	c.Check(validateMessage(
		common.NewTx(
			GetRandomTxHash(),
			GetRandomBNBAddress(),
			GetRandomBNBAddress(),
			common.Coins{
				common.NewCoin(common.RuneAsset(), sdk.NewUint(3429850000)),
			},
			common.BNBGasFeeSingleton,
			"",
		),
		common.Asset{},
		"bnbYYY",
	), NotNil)
	c.Check(validateMessage(
		common.NewTx(
			GetRandomTxHash(),
			GetRandomBNBAddress(),
			GetRandomBNBAddress(),
			common.Coins{
				common.NewCoin(common.RuneAsset(), sdk.ZeroUint()),
			},
			common.BNBGasFeeSingleton,
			"",
		),
		common.BNBAsset,
		"bnbYYY",
	), NotNil)
	c.Check(validateMessage(
		common.NewTx(
			GetRandomTxHash(),
			"",
			GetRandomBNBAddress(),
			common.Coins{
				common.NewCoin(common.RuneAsset(), sdk.NewUint(3429850000)),
			},
			common.BNBGasFeeSingleton,
			"",
		),
		common.BNBAsset,
		"bnbYYY",
	), NotNil)
	c.Check(validateMessage(
		common.NewTx(
			GetRandomTxHash(),
			GetRandomBNBAddress(),
			GetRandomBNBAddress(),
			common.Coins{
				common.NewCoin(common.RuneAsset(), sdk.NewUint(3429850000)),
			},
			common.BNBGasFeeSingleton,
			"",
		),
		common.BNBAsset,
		"",
	), NotNil)
}

func (s SwapSuite) TestCalculators(c *C) {
	X := sdk.NewUint(100 * common.One)
	x := sdk.NewUint(10 * common.One)
	Y := sdk.NewUint(100 * common.One)

	// These calculations are verified by using the spreadsheet
	// https://docs.google.com/spreadsheets/d/1wJHYBRKBdw_WP7nUyVnkySPkOmPUNoiRGsEqgBVVXKU/edit#gid=0
	c.Check(calcAssetEmission(X, x, Y).Uint64(), Equals, uint64(826446280))
	c.Check(calcLiquidityFee(X, x, Y).Uint64(), Equals, uint64(82644628))
	c.Check(calcTradeSlip(X, x).Uint64(), Equals, uint64(2100))
}
