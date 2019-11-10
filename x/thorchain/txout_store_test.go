package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

type TxOutStoreSuite struct{}

var _ = Suite(&TxOutStoreSuite{})

func (s TxOutStoreSuite) TestMinusGas(c *C) {
	ctx, k := setupKeeperForTest(c)

	p := NewPool()
	p.Asset = common.BNBAsset
	p.BalanceRune = sdk.NewUint(100 * common.One)
	p.BalanceAsset = sdk.NewUint(100 * common.One)
	k.SetPool(ctx, p)

	loki := NewPool()
	loki.Asset, _ = common.NewAsset("BNB.LOKI")
	loki.BalanceRune = sdk.NewUint(100 * common.One)
	loki.BalanceAsset = sdk.NewUint(100 * common.One)
	k.SetPool(ctx, loki)
	poolAddrMgr := NewPoolAddressManager(k)
	ctx = ctx.WithBlockHeight(1)
	poolAddrMgr.BeginBlock(ctx)

	txOutStore := NewTxOutStore(&MockTxOutSetter{}, poolAddrMgr)
	txOutStore.NewBlock(uint64(1))

	bnbAddress, err := common.NewAddress("bnb1xlvns0n2mxh77mzaspn2hgav4rr4m8eerfju38")
	c.Assert(err, IsNil)

	item := &TxOutItem{
		Chain:       common.BNBChain,
		PoolAddress: GetRandomPubKey(),
		ToAddress:   bnbAddress,
		Coin:        common.NewCoin(common.BNBAsset, sdk.NewUint(3980500*common.One)),
	}

	txOutStore.AddTxOutItem(ctx, k, item, true)
	bnbPool := k.GetPool(ctx, common.BNBAsset)
	// happy path
	c.Assert(bnbPool.BalanceAsset.String(), Equals, "10000000000")
	c.Assert(item.Coin.Amount.String(), Equals, "398049999962500")

	item.Coin = common.NewCoin(common.BNBAsset, sdk.NewUint(1))

	txOutStore.AddTxOutItem(ctx, k, item, true)
	// test not enough coins to pay for gas
	c.Assert(item.Coin.Amount.String(), Equals, "0")

	item.Coin = common.NewCoin(common.RuneAsset(), sdk.NewUint(20*common.One))

	txOutStore.AddTxOutItem(ctx, k, item, true)
	bnbPool = k.GetPool(ctx, common.BNBAsset)
	// test takes gas out of rune
	c.Assert(bnbPool.BalanceAsset.String(), Equals, "9999962500")
	c.Assert(bnbPool.BalanceRune.String(), Equals, "10000037500")
	c.Assert(item.Coin.Amount.String(), Equals, "1999962500")

	item.Coin = common.NewCoin(loki.Asset, sdk.NewUint(20*common.One))
	txOutStore.AddTxOutItem(ctx, k, item, true)
	lokiPool := k.GetPool(ctx, loki.Asset)
	bnbPool = k.GetPool(ctx, common.BNBAsset)
	// test takes gas out of loki pool
	c.Assert(bnbPool.BalanceAsset.String(), Equals, "9999925000")
	c.Assert(bnbPool.BalanceRune.String(), Equals, "10000075000")
	c.Assert(lokiPool.BalanceRune.String(), Equals, "9999962500", Commentf("%+v\n", lokiPool))
	c.Assert(item.Coin.Amount.String(), Equals, "1999962500")

	bnbPool = k.GetPool(ctx, common.BNBAsset)
	bnbPool.BalanceAsset = sdk.NewUint(1 * common.One)
	bnbPool.BalanceRune = sdk.NewUint(1000 * common.One)
	k.SetPool(ctx, bnbPool)
	item.Coin = common.NewCoin(common.RuneAsset(), sdk.NewUint(10000*common.One))
	txOutStore.AddTxOutItem(ctx, k, item, true)
	bnbPool = k.GetPool(ctx, common.BNBAsset)
	// test takes gas out of loki pool
	c.Assert(bnbPool.BalanceRune.String(), Equals, "100037500000")
}
