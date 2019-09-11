package swapservice

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"

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
func newPoolForTest(ticker common.Ticker, balanceRune, balanceToken sdk.Uint) Pool {
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
		minimumRefundAmount sdk.Uint
		pool                Pool
		ticker              common.Ticker
		amount              sdk.Uint
		expectedCoin        common.Coin
	}{
		{
			name:                "invalid-MRRA",
			minimumRefundAmount: sdk.ZeroUint(),
			pool:                newPoolForTest(common.RuneTicker, sdk.NewUint(100*One), sdk.NewUint(100*One)),
			ticker:              common.RuneTicker,
			amount:              sdk.NewUint(100 * One),
			expectedCoin:        common.NewCoin(common.RuneTicker, common.NewAmountFromFloat(100*float64(One))),
		},
		{
			name:                "OneRune-MRRA",
			minimumRefundAmount: sdk.NewUint(One),
			pool:                newPoolForTest(common.RuneTicker, sdk.NewUint(100*One), sdk.NewUint(100*One)),
			ticker:              common.RuneTicker,
			amount:              sdk.NewUint(100 * One),
			expectedCoin:        common.NewCoin(common.RuneTicker, common.NewAmountFromFloat(99*float64(One))),
		},
		{
			name:                "No-Refund",
			minimumRefundAmount: sdk.NewUint(One),
			pool:                newPoolForTest(common.RuneTicker, sdk.NewUint(100*One), sdk.NewUint(100*One)),
			ticker:              common.RuneTicker,
			amount:              sdk.NewUint(One / 2),
			expectedCoin:        common.NewCoin(common.RuneTicker, common.ZeroAmount),
		},
		{
			name:                "invalid-MRRA-BNB-refund-all",
			minimumRefundAmount: sdk.ZeroUint(),
			pool:                newPoolForTest(bnbTicker, sdk.NewUint(100*One), sdk.NewUint(100*One)),
			ticker:              bnbTicker,
			amount:              sdk.NewUint(5 * One),
			expectedCoin:        common.NewCoin(bnbTicker, common.NewAmountFromFloat(5*float64(One))),
		},
		{
			name:                "MRRA-BNB-refund-normal",
			minimumRefundAmount: sdk.NewUint(One),
			pool:                newPoolForTest(bnbTicker, sdk.NewUint(100*One), sdk.NewUint(100*One)),
			ticker:              bnbTicker,
			amount:              sdk.NewUint(5 * One),
			expectedCoin:        common.NewCoin(bnbTicker, common.NewAmountFromFloat(4*float64(One))),
		},
		{
			name:                "MRRA-BNB-refund-normal-1",
			minimumRefundAmount: sdk.NewUint(One),
			pool:                newPoolForTest(bnbTicker, sdk.NewUint(100*One), sdk.NewUint(One)),
			ticker:              bnbTicker,
			amount:              sdk.NewUint(5 * One),
			expectedCoin:        common.NewCoin(bnbTicker, common.NewAmountFromFloat(4.99*float64(One))),
		},
		{
			name:                "MRRA-BNB-no-refund",
			minimumRefundAmount: sdk.NewUint(One),
			pool:                newPoolForTest(bnbTicker, sdk.NewUint(100*One), sdk.NewUint(100*One)),
			ticker:              bnbTicker,
			amount:              sdk.NewUint(One / 2),
			expectedCoin:        common.NewCoin(bnbTicker, common.ZeroAmount),
		},
	}
	for _, item := range inputs {
		ctx := getTestContext()
		ctx = ctx.WithValue(mocks.RefundAdminConfigKeyMRRA, item.minimumRefundAmount).
			WithValue(mocks.RefundPoolKey, item.pool)
		c.Log(item.name)
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
		minimumRefundAmount sdk.Uint
		pool                Pool
		result              sdk.Result
		msg                 sdk.Msg
		out                 *TxOutItem
	}{
		{
			name:                "result-ok",
			minimumRefundAmount: sdk.NewUint(One),
			pool:                newPoolForTest(bnbTicker, sdk.NewUint(100*One), sdk.NewUint(100*One)),
			result: sdk.Result{
				Code: sdk.CodeOK,
			},
			msg: nil,
			out: nil,
		},
		{
			name:                "msg-type-setpooldata",
			minimumRefundAmount: sdk.NewUint(One),
			pool:                newPoolForTest(bnbTicker, sdk.NewUint(100*One), sdk.NewUint(100*One)),
			result: sdk.Result{
				Code: sdk.CodeOK,
			},
			msg: NewMsgSetPoolData(bnbTicker, PoolEnabled, accountAddress),
			out: nil,
		},
		{
			name:                "msg-type-swap",
			minimumRefundAmount: sdk.NewUint(One),
			pool:                newPoolForTest(bnbTicker, sdk.NewUint(100*One), sdk.NewUint(100*One)),
			result:              sdk.ErrUnknownRequest("whatever").Result(),
			msg:                 NewMsgSwap(txID, common.RuneTicker, bnbTicker, sdk.NewUint(5*One), "asdf", "asdf", sdk.NewUint(One), accountAddress),
			out: &TxOutItem{
				ToAddress: "asdf",
				Coins: common.Coins{
					common.NewCoin(common.RuneTicker, common.NewAmountFromFloat(4.0*float64(One))),
				},
			},
		},
	}
	for _, item := range inputs {
		ctx := getTestContext()
		ctx = ctx.WithValue(mocks.RefundAdminConfigKeyMRRA, item.minimumRefundAmount).
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

func (RefundSuite) TestProcessRefund1(c *C) {
	ctx := getTestContext()
	refundStoreAccessor := mocks.NewMockRefundStoreAccessor()
	addr := sdk.AccAddress("rune1gqva7eh03jkz39tk8m3tlw7ch558dz0ncdag0j")
	s := NewTxOutStore(MockTxOutSetter{})
	s.NewBlock(1)
	processRefund(ctx, &sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}, s, refundStoreAccessor, NewMsgNoOp(addr))
	c.Assert(len(s.blockOut.TxArray), Equals, 0)
	s.CommitBlock(ctx)
	txId, err := common.NewTxID("4D60A73FEBD42592DB697EF1DA020A214EC3102355D0E1DD07B18557321B106X")
	if nil != err {
		c.Errorf("fail to create tx id,%s", err)
	}
	bnbAddress, err := common.NewBnbAddress("tbnb1c2yvdphs674vlkp2s2e68cw89garykgau2c8vx")
	if nil != err {
		c.Errorf("fail to create bnb address,%s", err)
	}
	ctx = ctx.WithValue(mocks.RefundAdminConfigKeyMRRA, sdk.NewUint(2*One))
	ctx = ctx.WithValue(mocks.RefundPoolKey, newPoolForTest(common.BNBTicker, sdk.NewUint(100*One), sdk.NewUint(100*One)))
	// stake refund test
	stakeMsg := NewMsgSetStakeData(common.BNBTicker, sdk.NewUint(100*One), sdk.NewUint(100*One), bnbAddress, txId, addr)
	result := sdk.ErrUnknownRequest("invalid").Result()
	s.NewBlock(2)
	processRefund(ctx, &result, s, refundStoreAccessor, stakeMsg)
	s.CommitBlock(ctx)
	c.Assert(len(s.blockOut.TxArray) > 0, Equals, true)

	//stake refund test
	stakeMsg1 := NewMsgSetStakeData(common.BNBTicker, sdk.NewUint(One/2), sdk.NewUint(One/2), bnbAddress, txId, addr)
	result1 := sdk.ErrUnknownRequest("invalid").Result()
	s.NewBlock(2)
	processRefund(ctx, &result1, s, refundStoreAccessor, stakeMsg1)
	s.CommitBlock(ctx)
	c.Assert(len(result1.Events) > 0, Equals, true)
	c.Assert(len(s.blockOut.TxArray) > 0, Equals, false)

	//swap refund test
	swapMsg := NewMsgSwap(txId, common.RuneTicker, common.BNBTicker, sdk.NewUint(One*2/3), bnbAddress, bnbAddress, sdk.NewUint(One*2), addr)
	resultMsg := sdk.ErrUnknownRequest("invalid").Result()
	s.NewBlock(3)
	processRefund(ctx, &resultMsg, s, refundStoreAccessor, swapMsg)
	s.CommitBlock(ctx)
	c.Assert(len(resultMsg.Events) > 0, Equals, true)

	swapNoop := NewMsgNoOp(addr)
	resultNoop := sdk.ErrUnknownRequest("invalid").Result()
	s.NewBlock(3)
	processRefund(ctx, &resultNoop, s, refundStoreAccessor, swapNoop)
	s.CommitBlock(ctx)
	c.Assert(len(s.blockOut.TxArray), Equals, 0)

}
