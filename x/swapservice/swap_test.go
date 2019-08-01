package swapservice

import (
	. "gopkg.in/check.v1"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/jpthor/cosmos-swap/x/swapservice/mocks"
	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

type SwapSuite struct{}

var _ = Suite(&SwapSuite{})

func (s SwapSuite) TestSwap(c *C) {
	poolStorage := mocks.MockPoolStorage{}
	key := sdk.NewKVStoreKey("test")
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
	inputs := []struct {
		name         string
		source       string
		target       string
		amount       string
		requester    string
		destination  string
		returnAmount string
		expectedErr  error
	}{
		{
			name:         "empty-source",
			source:       "",
			target:       "BNB",
			amount:       "100",
			requester:    "tester",
			destination:  "whatever",
			returnAmount: "0",
			expectedErr:  errors.New("source is empty"),
		},
		{
			name:         "empty-target",
			source:       "RUNE",
			target:       "",
			amount:       "100",
			requester:    "tester",
			destination:  "whatever",
			returnAmount: "0",
			expectedErr:  errors.New("target is empty"),
		},
		{
			name:         "empty-amount",
			source:       "RUNE",
			target:       "BNB",
			amount:       "",
			requester:    "tester",
			destination:  "whatever",
			returnAmount: "0",
			expectedErr:  errors.New("amount is empty"),
		},
		{
			name:         "empty-requester",
			source:       "RUNE",
			target:       "BNB",
			amount:       "100",
			requester:    "",
			destination:  "whatever",
			returnAmount: "0",
			expectedErr:  errors.New("requester is empty"),
		},
		{
			name:         "empty-destination",
			source:       "RUNE",
			target:       "BNB",
			amount:       "100",
			requester:    "tester",
			destination:  "",
			returnAmount: "0",
			expectedErr:  errors.New("destination is empty"),
		},
		{
			name:         "pool-not-exist",
			source:       "NOTEXIST",
			target:       "RUNE",
			amount:       "100",
			requester:    "tester",
			destination:  "don'tknow",
			returnAmount: "0",
			expectedErr:  errors.New("pool-NOTEXIST doesn't exist"),
		},
		{
			name:         "pool-not-exist-1",
			source:       "RUNE",
			target:       "NOTEXIST",
			amount:       "100",
			requester:    "tester",
			destination:  "don'tknow",
			returnAmount: "0",
			expectedErr:  errors.New("pool-NOTEXIST doesn't exist"),
		},
		{
			name:         "swap",
			source:       "RUNE",
			target:       "BNB",
			amount:       "5",
			requester:    "tester",
			destination:  "don'tknow",
			returnAmount: "4.53514739",
			expectedErr:  nil,
		},
		{
			name:         "swap",
			source:       "BTC",
			target:       "BNB",
			amount:       "5",
			requester:    "tester",
			destination:  "don'tknow",
			returnAmount: "4.15017810",
			expectedErr:  nil,
		},
	}
	for _, item := range inputs {
		amount, err := swap(ctx, poolStorage, item.source, item.target, item.amount, item.requester, item.destination)
		if item.expectedErr == nil {
			c.Assert(err, IsNil)
		} else {
			c.Assert(err.Error(), Equals, item.expectedErr.Error())
		}
		c.Check(item.returnAmount, Equals, amount)
	}
}

func (s SwapSuite) TestSwapCalculation(c *C) {
	inputs := []struct {
		name              string
		source            string
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
			source:            types.RuneTicker,
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
			source:            types.RuneTicker,
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
			source:            types.RuneTicker,
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
			source:            types.RuneTicker,
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
			source:            types.RuneTicker,
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
			source:            types.RuneTicker,
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
			source:            types.RuneTicker,
			runeBalance:       100.0,
			tokenBalance:      100.0,
			amountToSwap:      5.0,
			runeBalanceAfter:  105.0,
			tokenBalanceAfter: 95.46,
			amountToReturn:    4.54,
		},
		{
			name:              "normal-rune-1",
			source:            types.RuneTicker,
			runeBalance:       1000.0,
			tokenBalance:      1000.0,
			amountToSwap:      20.0,
			runeBalanceAfter:  1020.0,
			tokenBalanceAfter: 980.78,
			amountToReturn:    19.22,
		},
		{
			name:              "normal-rune-2",
			source:            types.RuneTicker,
			runeBalance:       10000.0,
			tokenBalance:      10000.0,
			amountToSwap:      20.0,
			runeBalanceAfter:  10020.0,
			tokenBalanceAfter: 9980.08,
			amountToReturn:    19.92,
		},
		{
			name:              "normal-token",
			source:            "BNB",
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
