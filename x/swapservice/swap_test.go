package swapservice

import (
	"math"
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/jpthor/cosmos-swap/x/swapservice/mocks"
	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

func TestSwap(t *testing.T) {
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
		err          error
	}{
		{
			name:         "empty-source",
			source:       "",
			target:       "BNB",
			amount:       "100",
			requester:    "tester",
			destination:  "whatever",
			returnAmount: "0",
			err:          errors.New("source is empty"),
		},
		{
			name:         "empty-target",
			source:       "RUNE",
			target:       "",
			amount:       "100",
			requester:    "tester",
			destination:  "whatever",
			returnAmount: "0",
			err:          errors.New("target is empty"),
		},
		{
			name:         "empty-amount",
			source:       "RUNE",
			target:       "BNB",
			amount:       "",
			requester:    "tester",
			destination:  "whatever",
			returnAmount: "0",
			err:          errors.New("amount is empty"),
		},
		{
			name:         "empty-requester",
			source:       "RUNE",
			target:       "BNB",
			amount:       "100",
			requester:    "",
			destination:  "whatever",
			returnAmount: "0",
			err:          errors.New("requester is empty"),
		},
		{
			name:         "empty-destination",
			source:       "RUNE",
			target:       "BNB",
			amount:       "100",
			requester:    "tester",
			destination:  "",
			returnAmount: "0",
			err:          errors.New("destination is empty"),
		},
		{
			name:         "pool-not-exist",
			source:       "NOTEXIST",
			target:       "RUNE",
			amount:       "100",
			requester:    "tester",
			destination:  "don'tknow",
			returnAmount: "0",
			err:          errors.New("pool-NOTEXIST doesn't exist"),
		},
		{
			name:         "pool-not-exist-1",
			source:       "RUNE",
			target:       "NOTEXIST",
			amount:       "100",
			requester:    "tester",
			destination:  "don'tknow",
			returnAmount: "0",
			err:          errors.New("pool-NOTEXIST doesn't exist"),
		},
		{
			name:         "swap",
			source:       "RUNE",
			target:       "BNB",
			amount:       "5",
			requester:    "tester",
			destination:  "don'tknow",
			returnAmount: "4.53514739",
			err:          nil,
		},
		{
			name:         "swap",
			source:       "BTC",
			target:       "BNB",
			amount:       "5",
			requester:    "tester",
			destination:  "don'tknow",
			returnAmount: "4.15017810",
			err:          nil,
		},
	}
	for _, item := range inputs {
		t.Run(item.name, func(st *testing.T) {
			amount, err := swap(ctx, poolStorage, item.source, item.target, item.amount, item.requester, item.destination)
			if nil != item.err {
				if err == nil {
					t.Errorf("we expect err :%s , however we didn't get it", item.err)
					return
				}
				if err.Error() != item.err.Error() {
					t.Errorf("we expect err : %s , however we got %s ", item.err, err)
					return
				}
				return
			}
			if item.err == nil && err != nil {
				t.Errorf("we are not expecting err, however we got :%s ", err)
				return
			}
			if item.returnAmount != amount {
				t.Errorf("we expected the return amonut to be %s , however we got %s", item.amount, amount)
			}
		})

	}
}
func TestSwapCalculation(t *testing.T) {
	inputs := []struct {
		name              string
		source            string
		runeBalance       float64
		tokenBalance      float64
		amountToSwap      float64
		runeBalanceAfter  float64
		tokenBalanceAfter float64
		amountToReturn    float64
		err               error
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
			err:               errors.New("invalid balance"),
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
			err:               errors.New("invalid balance"),
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
			err:               errors.New("invalid balance"),
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
			err:               errors.New("invalid balance"),
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
			err:               errors.New("amount is invalid"),
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
			err:               errors.New("amount is invalid"),
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
		t.Run(item.name, func(st *testing.T) {
			r, t, a, err := calculateSwap(item.source, item.runeBalance, item.tokenBalance, item.amountToSwap)
			if item.err != nil {
				if err == nil {
					st.Errorf("expected err: %s, however we didn't get it", item.err)
					return
				}
				if item.err.Error() != err.Error() {
					st.Errorf("we expected err:%s ,however we got %s ", item.err, err)
					return
				}
				return
			}
			if item.err == nil && err != nil {
				st.Errorf("we are not expecting error , however we got :%s", err)
				return
			}
			if round(r) != item.runeBalanceAfter || round(t) != item.tokenBalanceAfter || round(a) != item.amountToReturn {
				st.Errorf("expected rune balance after: %f,token balance:%f ,amount:%f, however we got rune balance:%f,token balance:%f,amount:%f", item.runeBalanceAfter, item.tokenBalanceAfter, item.amountToReturn, r, t, a)
			}
		})
	}
}
func round(input float64) float64 {
	return math.Round(input*100) / 100
}
