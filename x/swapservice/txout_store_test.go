package swapservice

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"
)

type TxOutStoreSuite struct{}

var _ = Suite(&TxOutStoreSuite{})

func (s TxOutStoreSuite) TestMinusGas(c *C) {
	fmt.Println("Testing tx out store")
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

	txOutStore := NewTxOutStore(&MockTxOutSetter{})
	txOutStore.NewBlock(uint64(1))

	bnbAddress, err := common.NewAddress("bnb1xlvns0n2mxh77mzaspn2hgav4rr4m8eerfju38")
	c.Assert(err, IsNil)

	item := &TxOutItem{
		PoolAddress: bnbAddress,
		ToAddress:   bnbAddress,
		Coins: common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(3980500*common.One)),
		},
	}

	txOutStore.AddTxOutItem(ctx, k, item, true)
	bnbPool := k.GetPool(ctx, common.BNBAsset)
	// happy path
	c.Assert(bnbPool.BalanceAsset.String(), Equals, "10000000000")
	c.Assert(item.Coins[0].Amount.String(), Equals, "398049999970000")

	item.Coins = common.Coins{
		common.NewCoin(common.BNBAsset, sdk.NewUint(1)),
	}

	txOutStore.AddTxOutItem(ctx, k, item, true)
	// test not enough coins to pay for gas
	c.Assert(item.Coins[0].Amount.String(), Equals, "0")

	item.Coins = common.Coins{
		common.NewCoin(common.RuneA1FAsset, sdk.NewUint(20*common.One)),
	}

	txOutStore.AddTxOutItem(ctx, k, item, true)
	bnbPool = k.GetPool(ctx, common.BNBAsset)
	// test takes gas out of rune
	c.Assert(bnbPool.BalanceAsset.String(), Equals, "9999970000")
	c.Assert(bnbPool.BalanceRune.String(), Equals, "10000030000")
	c.Assert(item.Coins[0].Amount.String(), Equals, "1999970000")

	item.Coins = common.Coins{
		common.NewCoin(loki.Asset, sdk.NewUint(20*common.One)),
	}
	txOutStore.AddTxOutItem(ctx, k, item, true)
	lokiPool := k.GetPool(ctx, loki.Asset)
	bnbPool = k.GetPool(ctx, common.BNBAsset)
	// test takes gas out of loki pool
	c.Assert(bnbPool.BalanceAsset.String(), Equals, "9999940000")
	c.Assert(bnbPool.BalanceRune.String(), Equals, "10000060000")
	c.Assert(lokiPool.BalanceRune.String(), Equals, "9999970000", Commentf("%+v\n", lokiPool))
	c.Assert(item.Coins[0].Amount.String(), Equals, "1999970000")

}
