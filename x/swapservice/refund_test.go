package swapservice

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/common"
	"gitlab.com/thorchain/bepswap/statechain/x/swapservice/mocks"
)

func Test(t *testing.T) { TestingT(t) }

type RefundSuite struct{}

var _ = Suite(&RefundSuite{})

func getTestContext() sdk.Context {
	key := sdk.NewKVStoreKey("test")
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	return sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())

}
func newPoolForTest(ticker common.Ticker, balanceRune, balanceToken common.Amount) Pool {
	ps := NewPool()
	ps.BalanceToken = balanceToken
	ps.BalanceRune = balanceRune
	ps.Ticker = ticker
	return ps
}
func (*RefundSuite) TestGetRefundCoin(c *C) {

	refundStoreAccessor := mocks.NewMockRefundStoreAccessor()
	bnbTicker, err := common.NewTicker("BNB")
	c.Assert(err, IsNil)
	inputs := []struct {
		name                string
		minimumRefundAmount common.Amount
		pool                Pool
		ticker              common.Ticker
		amount              common.Amount
		expectedCoin        common.Coin
	}{
		{
			name:                "invalid-MRRA",
			minimumRefundAmount: common.Amount("invalid"),
			pool:                newPoolForTest(common.RuneTicker, common.NewAmountFromFloat(100), common.NewAmountFromFloat(100)),
			ticker:              common.RuneTicker,
			amount:              common.NewAmountFromFloat(100),
			expectedCoin:        common.NewCoin(common.RuneTicker, common.NewAmountFromFloat(100)),
		},
		{
			name:                "OneRune-MRRA",
			minimumRefundAmount: common.NewAmountFromFloat(1.0),
			pool:                newPoolForTest(common.RuneTicker, common.NewAmountFromFloat(100), common.NewAmountFromFloat(100)),
			ticker:              common.RuneTicker,
			amount:              common.NewAmountFromFloat(100),
			expectedCoin:        common.NewCoin(common.RuneTicker, common.NewAmountFromFloat(99)),
		},
		{
			name:                "No-Refund",
			minimumRefundAmount: common.NewAmountFromFloat(1.0),
			pool:                newPoolForTest(common.RuneTicker, common.NewAmountFromFloat(100), common.NewAmountFromFloat(100)),
			ticker:              common.RuneTicker,
			amount:              common.NewAmountFromFloat(0.5),
			expectedCoin:        common.NewCoin(common.RuneTicker, common.ZeroAmount),
		},
		{
			name:                "invalid-MRRA-BNB-refund-all",
			minimumRefundAmount: common.Amount("invalid"),
			pool:                newPoolForTest(bnbTicker, common.NewAmountFromFloat(100), common.NewAmountFromFloat(100)),
			ticker:              bnbTicker,
			amount:              common.NewAmountFromFloat(5),
			expectedCoin:        common.NewCoin(bnbTicker, common.NewAmountFromFloat(5)),
		},
		{
			name:                "MRRA-BNB-refund-normal",
			minimumRefundAmount: common.NewAmountFromFloat(1.0),
			pool:                newPoolForTest(bnbTicker, common.NewAmountFromFloat(100), common.NewAmountFromFloat(100)),
			ticker:              bnbTicker,
			amount:              common.NewAmountFromFloat(5),
			expectedCoin:        common.NewCoin(bnbTicker, common.NewAmountFromFloat(4)),
		},
		{
			name:                "MRRA-BNB-no-refund",
			minimumRefundAmount: common.NewAmountFromFloat(1.0),
			pool:                newPoolForTest(bnbTicker, common.NewAmountFromFloat(100), common.NewAmountFromFloat(100)),
			ticker:              bnbTicker,
			amount:              common.NewAmountFromFloat(0.5),
			expectedCoin:        common.NewCoin(bnbTicker, common.ZeroAmount),
		},
	}
	for _, item := range inputs {
		ctx := getTestContext()
		ctx = ctx.WithValue(mocks.RefundAdminConfigKey, item.minimumRefundAmount).
			WithValue(mocks.RefundPoolKey, item.pool)
		coin := getRefundCoin(ctx, item.ticker, item.amount, refundStoreAccessor)
		c.Assert(coin, Equals, item.expectedCoin)
	}
}

// TestProcessRefund is to test the processRefund
func (*RefundSuite) TestProcessRefund(c *C) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("rune", "runepub")
	refundStoreAccessor := mocks.NewMockRefundStoreAccessor()
	bnbTicker, err := common.NewTicker("BNB")
	c.Assert(err, IsNil)
	accountAddress, err := sdk.AccAddressFromBech32("rune1lz8kde0dc5ru63et7kykzzc97jhu7rg3yp2qxd")
	c.Assert(err, IsNil)
	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	inputs := []struct {
		name                string
		minimumRefundAmount common.Amount
		pool                Pool
		result              sdk.Result
		msg                 sdk.Msg
		out                 *TxOutItem
	}{
		{
			name:                "result-ok",
			minimumRefundAmount: common.NewAmountFromFloat(1.0),
			pool:                newPoolForTest(bnbTicker, common.NewAmountFromFloat(100), common.NewAmountFromFloat(100)),
			result: sdk.Result{
				Code: sdk.CodeOK,
			},
			msg: nil,
			out: nil,
		},
		{
			name:                "msg-type-setpooldata",
			minimumRefundAmount: common.NewAmountFromFloat(1.0),
			pool:                newPoolForTest(bnbTicker, common.NewAmountFromFloat(100), common.NewAmountFromFloat(100)),
			result: sdk.Result{
				Code: sdk.CodeOK,
			},
			msg: NewMsgSetPoolData(bnbTicker, PoolEnabled, accountAddress),
			out: nil,
		},
		{
			name:                "msg-type-swap",
			minimumRefundAmount: common.NewAmountFromFloat(1.0),
			pool:                newPoolForTest(bnbTicker, common.NewAmountFromFloat(100), common.NewAmountFromFloat(100)),
			result:              sdk.ErrUnknownRequest("whatever").Result(),
			msg:                 NewMsgSwap(txID, common.RuneTicker, bnbTicker, common.NewAmountFromFloat(5.0), "asdf", "asdf", "1.0", accountAddress),
			out: &TxOutItem{
				ToAddress: "asdf",
				Coins: common.Coins{
					common.NewCoin(common.RuneTicker, common.NewAmountFromFloat(4.0)),
				},
			},
		},
	}
	for _, item := range inputs {
		ctx := getTestContext()
		ctx = ctx.WithValue(mocks.RefundAdminConfigKey, item.minimumRefundAmount).
			WithValue(mocks.RefundPoolKey, item.pool)
		txStore := &TxOutStore{
			blockOut: nil,
		}
		txStore.NewBlock(1)
		processRefund(ctx, &item.result, txStore, refundStoreAccessor, item.msg)
		if nil == item.out {
			c.Assert(txStore.blockOut.TxArray, IsNil)
		} else {
			if len(txStore.blockOut.TxArray) == 0 {
				c.FailNow()
			}
			c.Assert(item.out.String(), Equals, txStore.blockOut.TxArray[0].String())
		}
	}
}
