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

	txOutStore.AddTxOutItem(ctx, k, item, true, false)
	bnbPool := k.GetPool(ctx, common.BNBAsset)
	// happy path
	c.Assert(bnbPool.BalanceAsset.String(), Equals, "10000000000")
	c.Assert(item.Coin.Amount.String(), Equals, "398049999962500")

	item.Coin = common.NewCoin(common.BNBAsset, sdk.NewUint(1))

	txOutStore.AddTxOutItem(ctx, k, item, true, false)
	// test not enough coins to pay for gas
	c.Assert(item.Coin.Amount.String(), Equals, "0")

	item.Coin = common.NewCoin(common.RuneAsset(), sdk.NewUint(20*common.One))

	txOutStore.AddTxOutItem(ctx, k, item, true, false)
	bnbPool = k.GetPool(ctx, common.BNBAsset)
	// test takes gas out of rune
	c.Assert(bnbPool.BalanceAsset.String(), Equals, "9999962500")
	c.Assert(bnbPool.BalanceRune.String(), Equals, "10000037500")
	c.Assert(item.Coin.Amount.String(), Equals, "1999962500")

	item.Coin = common.NewCoin(loki.Asset, sdk.NewUint(20*common.One))
	txOutStore.AddTxOutItem(ctx, k, item, true, false)
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
	txOutStore.AddTxOutItem(ctx, k, item, true, false)
	bnbPool = k.GetPool(ctx, common.BNBAsset)
	// test takes gas out of loki pool
	c.Assert(bnbPool.BalanceRune.String(), Equals, "100037500000")
}

func (s TxOutStoreSuite) TestAddOutTxItem(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	pk1, err := common.NewPoolPubKey(common.BNBChain, 0, GetRandomPubKey())
	c.Assert(err, IsNil)
	w.poolAddrMgr.currentPoolAddresses.Current = common.PoolPubKeys{pk1}

	acc1 := GetRandomNodeAccount(NodeActive)
	acc2 := GetRandomNodeAccount(NodeActive)
	acc3 := GetRandomNodeAccount(NodeActive)
	w.keeper.SetNodeAccount(w.ctx, acc1)
	w.keeper.SetNodeAccount(w.ctx, acc2)
	w.keeper.SetNodeAccount(w.ctx, acc3)

	ygg := NewYggdrasil(acc1.NodePubKey.Secp256k1)
	ygg.AddFunds(
		common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(40*common.One)),
		},
	)
	w.keeper.SetYggdrasil(w.ctx, ygg)

	ygg = NewYggdrasil(acc2.NodePubKey.Secp256k1)
	ygg.AddFunds(
		common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(50*common.One)),
		},
	)
	w.keeper.SetYggdrasil(w.ctx, ygg)

	ygg = NewYggdrasil(acc3.NodePubKey.Secp256k1)
	ygg.AddFunds(
		common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One)),
		},
	)
	w.keeper.SetYggdrasil(w.ctx, ygg)

	// Create voter
	inTxID := GetRandomTxHash()
	voter := NewTxInVoter(inTxID, []TxIn{
		TxIn{
			Signers: []sdk.AccAddress{w.activeNodeAccount.NodeAddress, acc1.NodeAddress, acc2.NodeAddress},
		},
	})
	w.keeper.SetTxInVoter(w.ctx, voter)

	// Should get acc2. Acc3 hasn't signed and acc2 is the highest value
	item := &TxOutItem{
		Chain:     common.BNBChain,
		ToAddress: GetRandomBNBAddress(),
		InHash:    inTxID,
		Coin:      common.NewCoin(common.BNBAsset, sdk.NewUint(20*common.One)),
	}

	w.txOutStore.AddTxOutItem(w.ctx, w.keeper, item, false, false)
	msgs := w.txOutStore.GetOutboundItems()
	c.Assert(msgs, HasLen, 1)
	c.Assert(msgs[0].PoolAddress.String(), Equals, acc2.NodePubKey.Secp256k1.String())

	// Should get acc1. Acc3 hasn't signed and acc1 now has the highest amount
	// of coin.
	item = &TxOutItem{
		Chain:     common.BNBChain,
		ToAddress: GetRandomBNBAddress(),
		InHash:    inTxID,
		Coin:      common.NewCoin(common.BNBAsset, sdk.NewUint(20*common.One)),
	}

	w.txOutStore.AddTxOutItem(w.ctx, w.keeper, item, false, false)
	msgs = w.txOutStore.GetOutboundItems()
	c.Assert(msgs, HasLen, 2)
	c.Assert(msgs[1].PoolAddress.String(), Equals, acc1.NodePubKey.Secp256k1.String())

	item = &TxOutItem{
		Chain:     common.BNBChain,
		ToAddress: GetRandomBNBAddress(),
		InHash:    inTxID,
		Coin:      common.NewCoin(common.BNBAsset, sdk.NewUint(1000*common.One)),
	}
	w.txOutStore.AddTxOutItem(w.ctx, w.keeper, item, false, false)
	msgs = w.txOutStore.GetOutboundItems()
	c.Assert(msgs, HasLen, 3)
	c.Assert(msgs[2].PoolAddress.String(), Equals, w.poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(common.BNBChain).PubKey.String())

}
