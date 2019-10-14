package swapservice

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type TxOutStoreSuite struct{}

var _ = Suite(&TxOutStoreSuite{})

func (s TxOutStoreSuite) TestMinusGas(c *C) {
	fmt.Println("Testing tx out store")
	ctx, k := setupKeeperForTest(c)

	p := NewPool()
	p.Ticker = common.BNBTicker
	p.BalanceRune = sdk.NewUint(100 * common.One)
	p.BalanceToken = sdk.NewUint(100 * common.One)
	k.SetPool(ctx, p)

	loki := NewPool()
	loki.Ticker = common.Ticker("LOKI")
	loki.BalanceRune = sdk.NewUint(100 * common.One)
	loki.BalanceToken = sdk.NewUint(100 * common.One)
	k.SetPool(ctx, loki)

	txOutStore := NewTxOutStore(&MockTxOutSetter{})
	txOutStore.NewBlock(uint64(1))

	bnbAddress, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)

	item := &TxOutItem{
		PoolAddress: bnbAddress,
		ToAddress:   bnbAddress,
		Coins: common.Coins{
			common.NewCoin(common.BNBTicker, sdk.NewUint(3980500*common.One)),
		},
	}

	txOutStore.AddTxOutItem(ctx, k, item)
	bnbPool := k.GetPool(ctx, common.BNBTicker)
	// happy path
	c.Assert(bnbPool.BalanceToken.String(), Equals, "10000000000")
	c.Assert(item.Coins[0].Amount.String(), Equals, "398049999970000")

	item.Coins = common.Coins{
		common.NewCoin(common.BNBTicker, sdk.NewUint(1)),
	}

	txOutStore.AddTxOutItem(ctx, k, item)
	// test not enough coins to pay for gas
	c.Assert(item.Coins[0].Amount.String(), Equals, "0")

	item.Coins = common.Coins{
		common.NewCoin(common.RuneTicker, sdk.NewUint(20*common.One)),
	}

	txOutStore.AddTxOutItem(ctx, k, item)
	bnbPool = k.GetPool(ctx, common.BNBTicker)
	// test takes gas out of rune
	c.Assert(bnbPool.BalanceToken.String(), Equals, "9999970000")
	c.Assert(bnbPool.BalanceRune.String(), Equals, "10000030000")
	c.Assert(item.Coins[0].Amount.String(), Equals, "1999970000")

	item.Coins = common.Coins{
		common.NewCoin(loki.Ticker, sdk.NewUint(20*common.One)),
	}
	txOutStore.AddTxOutItem(ctx, k, item)
	lokiPool := k.GetPool(ctx, loki.Ticker)
	bnbPool = k.GetPool(ctx, common.BNBTicker)
	// test takes gas out of loki pool
	c.Assert(bnbPool.BalanceToken.String(), Equals, "9999940000")
	c.Assert(bnbPool.BalanceRune.String(), Equals, "10000060000")
	c.Assert(lokiPool.BalanceRune.String(), Equals, "9999970000", Commentf("%+v\n", lokiPool))
	c.Assert(item.Coins[0].Amount.String(), Equals, "1999970000")

}
