package swapservice

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/statechain/cmd"
	"gitlab.com/thorchain/bepswap/statechain/x/swapservice/mocks"
	"gitlab.com/thorchain/bepswap/statechain/x/swapservice/types"
)

type SwapSuite struct{}

var _ = Suite(&SwapSuite{})
var keyStore = sdk.NewKVStoreKey(StoreKey)

func (s *SwapSuite) SetUpSuite(c *C) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	config.SetBech32PrefixForConsensusNode(cmd.Bech32PrefixConsAddr, cmd.Bech32PrefixConsPub)
}

func (s SwapSuite) TestSwap(c *C) {
	poolStorage := mocks.MockPoolStorage{}
	ctx, _ := setupKeeperForTest(c)
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
		globalSlipLimit common.Amount
		expectedErr     error
	}{
		{
			name:          "empty-source",
			requestTxHash: "hash",
			source:        common.Ticker(""),
			target:        common.BNBTicker,
			amount:        sdk.NewUint(100 * common.One),
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
			amount:        sdk.NewUint(100 * common.One),
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
			amount:        sdk.NewUint(100 * common.One),
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
			amount:        sdk.NewUint(100 * common.One),
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
			amount:        sdk.NewUint(100 * common.One),
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
			amount:        sdk.NewUint(100 * common.One),
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
			amount:        sdk.NewUint(100 * common.One),
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
			amount:        sdk.NewUint(50 * common.One),
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
			amount:        sdk.NewUint(9 * common.One),
			requester:     "tester",
			destination:   "don'tknow",
			returnAmount:  sdk.ZeroUint(),
			tradeTarget:   sdk.NewUint(9 * common.One),
			expectedErr:   errors.New("emit token 757511993 less than price limit 900000000"),
		},
		{
			name:          "swap-no-target-price-no-protection",
			requestTxHash: "hash",
			source:        common.RuneTicker,
			target:        common.BNBTicker,
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
			source:        common.RuneTicker,
			target:        common.BNBTicker,
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
			source:        "BTC",
			target:        common.BNBTicker,
			amount:        sdk.NewUint(5 * common.One),
			requester:     "tester",
			destination:   "don'tknow",
			returnAmount:  sdk.NewUint(415017809),
			tradeTarget:   sdk.NewUint(415017800),
			expectedErr:   nil,
		},
	}
	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
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
	X := sdk.NewUint(100 * common.One)
	x := sdk.NewUint(10 * common.One)
	Y := sdk.NewUint(100 * common.One)

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

func (s SwapSuite) TestHandleMsgSwap(c *C) {
	ctx, k := setupKeeperForTest(c)
	txOutStore := NewTxOutStore(k)
	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	addr, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)
	signerAddr, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	observerAddr, err := sdk.AccAddressFromBech32("bep1rtgz3lcaw8vw0yfsc8ga0rdgwa3qh9ju7vfsnk")
	bepConsPubKey := `bepcpub1zcjduepq4kn64fcjhf0fp20gp8var0rm25ca9jy6jz7acem8gckh0nkplznq85gdrg`
	ta := types.NewTrustAccount(addr, observerAddr, bepConsPubKey)
	k.SetNodeAccount(ctx, types.NewNodeAccount(signerAddr, NodeActive, ta))
	txOutStore.NewBlock(1)
	// no pool
	msg := NewMsgSwap(txID, common.RuneA1FTicker, common.BNBTicker, sdk.NewUint(common.One), addr, addr, sdk.ZeroUint(), observerAddr)
	res := handleMsgSwap(ctx, k, txOutStore, msg)
	c.Assert(res.Code, Equals, sdk.CodeInternal)
	pool := NewPool()
	pool.Ticker = common.BNBTicker
	pool.BalanceToken = sdk.NewUint(100 * common.One)
	pool.BalanceRune = sdk.NewUint(100 * common.One)
	k.SetPool(ctx, pool)

	res = handleMsgSwap(ctx, k, txOutStore, msg)
	c.Assert(res.IsOK(), Equals, true)

	msgSwapPriceProtection := NewMsgSwap(txID, common.RuneA1FTicker, common.BNBTicker, sdk.NewUint(common.One), addr, addr, sdk.NewUint(2*common.One), observerAddr)
	res1 := handleMsgSwap(ctx, k, txOutStore, msgSwapPriceProtection)
	c.Assert(res1.IsOK(), Equals, false)
	c.Assert(res1.Code, Equals, sdk.CodeInternal)

	poolTCAN := NewPool()
	tCanTicker, err := common.NewTicker("TCAN-014")
	c.Assert(err, IsNil)
	poolTCAN.Ticker = tCanTicker
	poolTCAN.BalanceToken = sdk.NewUint(334850000)
	poolTCAN.BalanceRune = sdk.NewUint(2349500000)
	k.SetPool(ctx, poolTCAN)

	txID1, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD211")
	m, err := ParseMemo("swap:RUNE-B1A:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlXXX:124958592")

	txIn := types.NewTxIn(common.Coins{
		common.NewCoin(tCanTicker, sdk.NewUint(20000000)),
	}, "swap:RUNE-B1A:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlXXX:124958592", addr, sdk.NewUint(1))
	msgSwapFromTxIn, err := getMsgSwapFromMemo(m.(SwapMemo), txID1, txIn, observerAddr)

	//msgSwapRune := NewMsgSwap(txID1, tCanTicker, common.RuneA1FTicker, sdk.NewUint(20000000), addr, addr, sdk.NewUint(124958593), observerAddr)
	res2 := handleMsgSwap(ctx, k, txOutStore, msgSwapFromTxIn.(MsgSwap))

	c.Assert(res2.IsOK(), Equals, true)
	c.Assert(res2.Code, Equals, sdk.CodeOK)

}
