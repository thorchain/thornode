package thorchain

import (
	"errors"
	"os"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

type SwapSuite struct{}

var _ = Suite(&SwapSuite{})

func (s *SwapSuite) SetUpSuite(c *C) {
	err := os.Setenv("NET", "other")
	c.Assert(err, IsNil)
	SetupConfigForTest()
}

type TestSwapKeeper struct {
	KVStoreDummy
}

func (k *TestSwapKeeper) PoolExist(ctx sdk.Context, asset common.Asset) bool {
	if asset.Equals(common.Asset{Chain: common.BNBChain, Symbol: "NOTEXIST", Ticker: "NOTEXIST"}) {
		return false
	}
	return true
}

func (k *TestSwapKeeper) GetPool(ctx sdk.Context, asset common.Asset) (types.Pool, error) {
	if asset.Equals(common.Asset{Chain: common.BNBChain, Symbol: "NOTEXIST", Ticker: "NOTEXIST"}) {
		return types.Pool{}, nil
	} else {
		return types.Pool{
			BalanceRune:  sdk.NewUint(100).MulUint64(common.One),
			BalanceAsset: sdk.NewUint(100).MulUint64(common.One),
			PoolUnits:    sdk.NewUint(100).MulUint64(common.One),
			Status:       types.Enabled,
			Asset:        asset,
		}, nil
	}
}
func (k *TestSwapKeeper) SetPool(ctx sdk.Context, ps types.Pool) error { return nil }

func (k *TestSwapKeeper) GetStaker(ctx sdk.Context, asset common.Asset, addr common.Address) (types.Staker, error) {
	if asset.Equals(common.Asset{Chain: common.BNBChain, Symbol: "NOTEXISTSTICKER", Ticker: "NOTEXISTSTICKER"}) {
		return types.Staker{}, errors.New("you asked for it")
	}
	return Staker{
		Asset:        asset,
		RuneAddress:  addr,
		AssetAddress: addr,
		Units:        sdk.NewUint(100),
		PendingRune:  sdk.ZeroUint(),
	}, nil
}

func (k *TestSwapKeeper) SetStaker(ctx sdk.Context, ps types.Staker) {}

func (k *TestSwapKeeper) AddToLiquidityFees(ctx sdk.Context, asset common.Asset, fs sdk.Uint) error {
	return nil
}

func (k *TestSwapKeeper) GetLowestActiveVersion(ctx sdk.Context) semver.Version {
	return constants.SWVersion
}

func (k *TestSwapKeeper) AddFeeToReserve(ctx sdk.Context, fee sdk.Uint) error { return nil }
func (k *TestSwapKeeper) UpsertEvent(ctx sdk.Context, event Event) error {
	return nil
}

func (k *TestSwapKeeper) GetGas(ctx sdk.Context, _ common.Asset) ([]sdk.Uint, error) {
	return []sdk.Uint{sdk.NewUint(37500), sdk.NewUint(30000)}, nil
}

func (s *SwapSuite) TestSwap(c *C) {
	poolStorage := &TestSwapKeeper{}
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
		expectedErr   sdk.Error
		events        []Event
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
			expectedErr:   sdk.NewError(DefaultCodespace, CodeValidationError, "Denom cannot be empty"),
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
			expectedErr:   sdk.NewError(DefaultCodespace, CodeValidationError, "target is empty"),
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
			expectedErr:   sdk.NewError(DefaultCodespace, CodeValidationError, "Tx ID cannot be empty"),
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
			expectedErr:   sdk.NewError(DefaultCodespace, CodeValidationError, "Amount cannot be zero"),
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
			expectedErr:   sdk.NewError(DefaultCodespace, CodeValidationError, "From address cannot be empty"),
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
			expectedErr:   sdk.NewError(DefaultCodespace, CodeValidationError, "To address cannot be empty"),
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
			expectedErr:   sdk.NewError(DefaultCodespace, CodeSwapFailPoolNotExist, "BNB.NOTEXIST pool doesn't exist"),
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
			expectedErr:   sdk.NewError(DefaultCodespace, CodeSwapFailPoolNotExist, "BNB.NOTEXIST pool doesn't exist"),
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
			events: []Event{
				Event{ID: 0, Height: 18, Type: "swap", InTx: common.Tx{ID: "hash", Chain: "BNB", FromAddress: "tester", ToAddress: "don't know", Coins: common.Coins{common.NewCoin(common.RuneAsset(), sdk.NewUint(5000000000))}, Gas: common.Gas{common.NewCoin(common.BNBAsset, sdk.NewUint(37500))}}},
			},
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
			expectedErr:   sdk.NewError(DefaultCodespace, CodeSwapFailTradeTarget, "emit asset 757511993 less than price limit 900000000"),
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
			events: []Event{
				Event{ID: 0, Height: 18, Type: "swap", InTx: common.Tx{ID: "hash", Chain: "BNB", FromAddress: "tester", ToAddress: "don'tknow", Coins: common.Coins{common.NewCoin(common.RuneAsset(), sdk.NewUint(800000000))}, Gas: common.Gas{common.NewCoin(common.BNBAsset, sdk.NewUint(37500))}}},
			},
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
			events: []Event{
				Event{ID: 0, Height: 18, Type: "swap", InTx: common.Tx{ID: "hash", Chain: "BNB", FromAddress: "tester", ToAddress: "don'tknow", Coins: common.Coins{common.NewCoin(common.RuneAsset(), sdk.NewUint(500000000))}, Gas: common.Gas{common.NewCoin(common.BNBAsset, sdk.NewUint(37500))}}},
			},
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
			events: []Event{
				Event{ID: 0, Height: 18, Type: "swap", InTx: common.Tx{ID: "hash", Chain: "BNB", FromAddress: "tester", ToAddress: "don'tknow", Coins: common.Coins{common.NewCoin(common.BTCAsset, sdk.NewUint(5*common.One))}, Gas: common.Gas{common.NewCoin(common.BNBAsset, sdk.NewUint(37500))}}},
				Event{ID: 0, Height: 18, Type: "swap", InTx: common.Tx{ID: "hash", Chain: "BNB", FromAddress: "tester", ToAddress: "don'tknow", Coins: common.Coins{common.NewCoin(common.RuneAsset(), sdk.NewUint(453514739))}, Gas: nil}},
			},
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
			BNBGasFeeSingleton,
			"",
		)
		tx.Chain = common.BNBChain
		amount, evts, err := swap(ctx, poolStorage, tx, item.target, item.destination, item.tradeTarget, sdk.NewUint(1000_000))
		if item.expectedErr == nil {
			c.Assert(err, IsNil)
			c.Assert(evts, HasLen, len(item.events))
			for i := range evts {
				c.Assert(item.events[i].Type, Equals, evts[i].Type())
				c.Assert(item.events[i].InTx.Equals(evts[i].InTx), Equals, true, Commentf("%+v\n%+v", item.events[i].InTx, evts[i].InTx))
				// TODO: test for price target, trade slip, and liquidity fee
			}
		} else {
			c.Assert(err, NotNil, Commentf("Expected: %s, got nil", item.expectedErr.Error()))
			c.Assert(err.Error(), Equals, item.expectedErr.Error())
		}

		c.Logf("expected amount:%s, actual amount:%s", item.returnAmount, amount)
		c.Check(item.returnAmount.Uint64(), Equals, amount.Uint64())

	}
}

func (s SwapSuite) TestValidatePools(c *C) {
	keeper := &TestSwapKeeper{}
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
			BNBGasFeeSingleton,
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
			BNBGasFeeSingleton,
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
			BNBGasFeeSingleton,
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
			BNBGasFeeSingleton,
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
			BNBGasFeeSingleton,
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
			BNBGasFeeSingleton,
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
			BNBGasFeeSingleton,
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
