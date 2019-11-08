package thorchain

import (
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/common"

	"gitlab.com/thorchain/bepswap/thornode/x/thorchain/mocks"
	"gitlab.com/thorchain/bepswap/thornode/x/thorchain/types"
)

type SwapSuite struct{}

var _ = Suite(&SwapSuite{})

func (s *SwapSuite) SetUpSuite(c *C) {
	err := os.Setenv("NET", "other")
	c.Assert(err, IsNil)
	SetupConfigForTest()
}

func (s SwapSuite) TestSwap(c *C) {
	poolStorage := mocks.MockPoolStorage{}
	ctx, _ := setupKeeperForTest(c)
	globalSlipLimit := common.Amount("0.200000")
	inputs := []struct {
		name            string
		requestTxHash   common.TxID
		source          common.Asset
		target          common.Asset
		amount          sdk.Uint
		requester       common.Address
		destination     common.Address
		returnAmount    sdk.Uint
		tradeTarget     sdk.Uint
		globalSlipLimit common.Amount
		expectedErr     error
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
			expectedErr:   errors.New("source is empty"),
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
			expectedErr:   errors.New("request tx hash is empty"),
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
			expectedErr:   errors.New("amount is zero"),
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
			expectedErr:   errors.New("requester is empty"),
		},
		{
			name:          "empty-destination",
			requestTxHash: "hash",
			source:        common.RuneAsset(),
			target:        common.BNBAsset,
			amount:        sdk.NewUint(100 * common.One),
			requester:     "tester",
			destination:   "",
			returnAmount:  sdk.ZeroUint(),
			expectedErr:   errors.New("destination is empty"),
		},
		{
			name:          "pool-not-exist",
			requestTxHash: "hash",
			source:        common.Asset{Chain: common.BNBChain, Ticker: "NOTEXIST", Symbol: "NOTEXIST"},
			target:        common.RuneAsset(),
			amount:        sdk.NewUint(100 * common.One),
			requester:     "tester",
			destination:   "don'tknow",
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
			name:          "swap-over-global-sliplimit",
			requestTxHash: "hash",
			source:        common.RuneAsset(),
			target:        common.BNBAsset,
			amount:        sdk.NewUint(50 * common.One),
			requester:     "tester",
			destination:   "don't know",
			returnAmount:  sdk.ZeroUint(),
			tradeTarget:   sdk.ZeroUint(),
			expectedErr:   errors.Errorf("fail to swap from %s to BNB.BNB: pool slip:0.928571 is over global pool slip limit :%s", common.RuneAsset(), globalSlipLimit),
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
			tradeTarget:   sdk.NewUint(415017800),
			expectedErr:   nil,
		},
	}
	txID := GetRandomTxHash()
	for _, item := range inputs {
		c.Logf("test name:%s", item.name)
		amount, err := swap(ctx, poolStorage, txID, item.source, item.target, item.amount, item.requester, item.destination, item.requestTxHash, item.tradeTarget, globalSlipLimit)
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
	ctx, _ := setupKeeperForTest(c)
	c.Check(validatePools(ctx, keeper, common.RuneAsset()), IsNil)
	c.Check(validatePools(ctx, keeper, common.Asset{Chain: common.BNBChain, Ticker: "NOTEXIST", Symbol: "NOTEXIST"}), NotNil)
}

func (s SwapSuite) TestValidateMessage(c *C) {
	c.Check(validateMessage(common.RuneAsset(), common.BNBAsset, sdk.NewUint(3429850000), "bnbXXXX", "bnbYYY", "txHASH"), IsNil)
	c.Check(validateMessage(common.RuneAsset(), common.BNBAsset, sdk.NewUint(3429850000), "bnbXXXX", "bnbYYY", ""), NotNil)
	c.Check(validateMessage(common.Asset{}, common.BNBAsset, sdk.NewUint(3429850000), "bnbXXXX", "bnbYYY", "txHASH"), NotNil)
	c.Check(validateMessage(common.RuneAsset(), common.Asset{}, sdk.NewUint(3429850000), "bnbXXXX", "bnbYYY", "txHASH"), NotNil)
	c.Check(validateMessage(common.RuneAsset(), common.BNBAsset, sdk.ZeroUint(), "bnbXXXX", "bnbYYY", "txHASH"), NotNil)
	c.Check(validateMessage(common.RuneAsset(), common.BNBAsset, sdk.NewUint(3429850000), "", "bnbYYY", "txHASH"), NotNil)
	c.Check(validateMessage(common.RuneAsset(), common.BNBAsset, sdk.NewUint(3429850000), "bnbXXXX", "", "txHASH"), NotNil)
}

func (s SwapSuite) TestCalculators(c *C) {
	X := sdk.NewUint(100 * common.One)
	x := sdk.NewUint(10 * common.One)
	Y := sdk.NewUint(100 * common.One)

	// These calculations are verified by using the spreadsheet
	// https://docs.google.com/spreadsheets/d/1wJHYBRKBdw_WP7nUyVnkySPkOmPUNoiRGsEqgBVVXKU/edit#gid=0
	c.Check(calcAssetEmission(X, x, Y).Uint64(), Equals, uint64(826446280))
	c.Check(calcLiquitityFee(X, x, Y).Uint64(), Equals, uint64(82644628))
	c.Check(calcPoolSlip(X, x), Equals, 0.1990990990990991)
	c.Check(calcTradeSlip(X, x), Equals, 0.21)
	c.Check(calcPriceSlip(X, x, Y), Equals, 1.210000001452)
	c.Check(calcOutputSlip(X, x), Equals, 0.09090909090909091)
}

func (s SwapSuite) TestHandleMsgSwap(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	txOutStore := NewTxOutStore(w.keeper, w.poolAddrMgr)
	txID := GetRandomTxHash()
	signerBNBAddr := GetRandomBNBAddress()
	observerAddr := w.activeNodeAccount.NodeAddress
	txOutStore.NewBlock(1)
	// no pool
	msg := NewMsgSwap(txID, common.RuneAsset(), common.BNBAsset, sdk.NewUint(common.One), signerBNBAddr, signerBNBAddr, sdk.ZeroUint(), observerAddr)
	res := handleMsgSwap(w.ctx, w.keeper, txOutStore, w.poolAddrMgr, msg)
	c.Assert(res.Code, Equals, sdk.CodeInternal)
	pool := NewPool()
	pool.Asset = common.BNBAsset
	pool.BalanceAsset = sdk.NewUint(100 * common.One)
	pool.BalanceRune = sdk.NewUint(100 * common.One)
	w.keeper.SetPool(w.ctx, pool)

	res = handleMsgSwap(w.ctx, w.keeper, txOutStore, w.poolAddrMgr, msg)
	c.Assert(res.IsOK(), Equals, true)

	msgSwapPriceProtection := NewMsgSwap(txID, common.RuneAsset(), common.BNBAsset, sdk.NewUint(common.One), signerBNBAddr, signerBNBAddr, sdk.NewUint(2*common.One), observerAddr)
	res1 := handleMsgSwap(w.ctx, w.keeper, txOutStore, w.poolAddrMgr, msgSwapPriceProtection)
	c.Assert(res1.IsOK(), Equals, false)
	c.Assert(res1.Code, Equals, sdk.CodeInternal)

	poolTCAN := NewPool()
	tCanAsset, err := common.NewAsset("BNB.TCAN-014")
	c.Assert(err, IsNil)
	poolTCAN.Asset = tCanAsset
	poolTCAN.BalanceAsset = sdk.NewUint(334850000)
	poolTCAN.BalanceRune = sdk.NewUint(2349500000)
	w.keeper.SetPool(w.ctx, poolTCAN)

	txID1, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD211")
	m, err := ParseMemo("swap:RUNE-B1A:bnb18jtza8j86hfyuj2f90zec0g5gvjh823e5psn2u:124958592")
	currentChainPoolAddr := w.poolAddrMgr.currentPoolAddresses.Current.GetByChain(common.BNBChain)
	c.Assert(currentChainPoolAddr, NotNil)
	txIn := types.NewTxIn(
		common.Coins{
			common.NewCoin(tCanAsset, sdk.NewUint(20000000)),
		},
		"swap:RUNE-B1A:bnb18jtza8j86hfyuj2f90zec0g5gvjh823e5psn2u:124958592",
		signerBNBAddr,
		GetRandomBNBAddress(),
		sdk.NewUint(1),
		currentChainPoolAddr.PubKey,
	)
	msgSwapFromTxIn, err := getMsgSwapFromMemo(m.(SwapMemo), txID1, txIn, observerAddr)
	c.Assert(err, IsNil)

	res2 := handleMsgSwap(w.ctx, w.keeper, txOutStore, w.poolAddrMgr, msgSwapFromTxIn.(MsgSwap))

	c.Assert(res2.IsOK(), Equals, true)
	c.Assert(res2.Code, Equals, sdk.CodeOK)

}
