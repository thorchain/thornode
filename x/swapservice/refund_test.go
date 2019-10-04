package swapservice

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
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
			pool:                newPoolForTest(common.RuneTicker, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One)),
			ticker:              common.RuneTicker,
			amount:              sdk.NewUint(100 * common.One),
			expectedCoin:        common.NewCoin(common.RuneTicker, sdk.NewUint(100*common.One)),
		},
		{
			name:                "OneRune-MRRA",
			minimumRefundAmount: sdk.NewUint(common.One),
			pool:                newPoolForTest(common.RuneTicker, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One)),
			ticker:              common.RuneTicker,
			amount:              sdk.NewUint(100 * common.One),
			expectedCoin:        common.NewCoin(common.RuneTicker, sdk.NewUint(99*common.One)),
		},
		{
			name:                "No-Refund",
			minimumRefundAmount: sdk.NewUint(common.One),
			pool:                newPoolForTest(common.RuneTicker, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One)),
			ticker:              common.RuneTicker,
			amount:              sdk.NewUint(common.One / 2),
			expectedCoin:        common.NewCoin(common.RuneTicker, sdk.ZeroUint()),
		},
		{
			name:                "invalid-MRRA-BNB-refund-all",
			minimumRefundAmount: sdk.ZeroUint(),
			pool:                newPoolForTest(bnbTicker, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One)),
			ticker:              bnbTicker,
			amount:              sdk.NewUint(5 * common.One),
			expectedCoin:        common.NewCoin(bnbTicker, sdk.NewUint(5*common.One)),
		},
		{
			name:                "MRRA-BNB-refund-normal",
			minimumRefundAmount: sdk.NewUint(common.One),
			pool:                newPoolForTest(bnbTicker, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One)),
			ticker:              bnbTicker,
			amount:              sdk.NewUint(5 * common.One),
			expectedCoin:        common.NewCoin(bnbTicker, sdk.NewUint(4*common.One)),
		},
		{
			name:                "MRRA-BNB-refund-normal-1",
			minimumRefundAmount: sdk.NewUint(common.One),
			pool:                newPoolForTest(bnbTicker, sdk.NewUint(100*common.One), sdk.NewUint(common.One)),
			ticker:              bnbTicker,
			amount:              sdk.NewUint(5 * common.One),
			expectedCoin:        common.NewCoin(bnbTicker, sdk.NewUint(499000000)),
		},
		{
			name:                "MRRA-BNB-no-refund",
			minimumRefundAmount: sdk.NewUint(common.One),
			pool:                newPoolForTest(bnbTicker, sdk.NewUint(100*common.One), sdk.NewUint(100*common.One)),
			ticker:              bnbTicker,
			amount:              sdk.NewUint(common.One / 2),
			expectedCoin:        common.NewCoin(bnbTicker, sdk.ZeroUint()),
		},
	}
	for _, item := range inputs {
		ctx := getTestContext()
		ctx = ctx.WithValue(mocks.RefundAdminConfigKeyMRRA, item.minimumRefundAmount).
			WithValue(mocks.RefundPoolKey, item.pool)
		c.Log(item.name)
		coin := getRefundCoin(ctx, item.ticker, item.amount, refundStoreAccessor)
		c.Assert(coin.Denom, Equals, item.expectedCoin.Denom)
		c.Assert(coin.Amount.Uint64(), Equals, item.expectedCoin.Amount.Uint64())
	}
}
