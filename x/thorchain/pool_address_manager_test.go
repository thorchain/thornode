package thorchain

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type PoolAddressManagerSuite struct{}

var _ = Suite(&PoolAddressManagerSuite{})

func (ps *PoolAddressManagerSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

func (PoolAddressManagerSuite) TestPoolAddressManager(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	c.Assert(w.poolAddrMgr.GetCurrentPoolAddresses().IsEmpty(), Equals, false)
	c.Assert(w.poolAddrMgr.GetCurrentPoolAddresses().IsEmpty(), Equals, false)

	rotateWindowOpenHeight := w.poolAddrMgr.GetCurrentPoolAddresses().RotateWindowOpenAt
	w.ctx = w.ctx.WithBlockHeight(rotateWindowOpenHeight)
	c.Assert(w.poolAddrMgr.BeginBlock(w.ctx), IsNil)
	w.txOutStore.NewBlock(uint64(rotateWindowOpenHeight))
	c.Assert(w.poolAddrMgr.IsRotateWindowOpen(), Equals, true)

	pk1, err := common.NewPoolPubKey(common.BNBChain, 0, GetRandomPubKey())
	c.Assert(err, IsNil)
	w.poolAddrMgr.GetCurrentPoolAddresses().Next = common.PoolPubKeys{pk1}
	w.poolAddrMgr.EndBlock(w.ctx, w.txOutStore)
	// no asset get moved , because THORNode just opened window, however THORNode should instruct signer to kick off key sign process
	c.Assert(w.txOutStore.GetOutboundItems(), HasLen, 1)
	poolBNB := createTempNewPoolForTest(w.ctx, w.keeper, "BNB.BNB", c)
	poolTCan := createTempNewPoolForTest(w.ctx, w.keeper, "BNB.TCAN-014", c)
	poolLoki := createTempNewPoolForTest(w.ctx, w.keeper, "BNB.LOK-3C0", c)
	rotatePoolHeight := w.poolAddrMgr.GetCurrentPoolAddresses().RotateAt
	w.ctx = w.ctx.WithBlockHeight(rotatePoolHeight)
	w.txOutStore.NewBlock(uint64(rotatePoolHeight))
	c.Assert(w.poolAddrMgr.BeginBlock(w.ctx), IsNil)
	w.poolAddrMgr.EndBlock(w.ctx, w.txOutStore)
	windowOpen := int64(constants.ValidatorsChangeWindow)
	rotatePerBlockHeight := int64(constants.RotatePerBlockHeight)
	c.Assert(w.poolAddrMgr.GetCurrentPoolAddresses().RotateAt, Equals, 100+rotatePerBlockHeight)
	c.Assert(w.poolAddrMgr.GetCurrentPoolAddresses().RotateWindowOpenAt, Equals, 100+rotatePerBlockHeight-windowOpen)
	c.Assert(w.txOutStore.GetOutboundItems(), HasLen, 4)
	c.Assert(w.txOutStore.getBlockOut().Valid(), IsNil)
	totalBond := sdk.ZeroUint()
	nodeAccounts, err := w.keeper.ListNodeAccounts(w.ctx)
	c.Assert(err, IsNil)
	for _, item := range nodeAccounts {
		totalBond = totalBond.Add(item.Bond)
	}
	defaultPoolGas := PoolRefundGasKey.Default()
	poolGas, err := strconv.Atoi(defaultPoolGas)

	c.Assert(err, IsNil)
	for _, item := range w.txOutStore.GetOutboundItems() {
		c.Assert(item.Valid(), IsNil)
		// make sure the fund is sending from previous pool address to current
		c.Assert(item.Coin.IsValid(), IsNil)
		chain := item.Coin.Asset.Chain
		newChainPoolAddr := w.poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(chain)
		c.Assert(newChainPoolAddr, NotNil)
		newPoolAddr, err := newChainPoolAddr.GetAddress()
		c.Assert(err, IsNil)
		c.Assert(item.ToAddress.String(), Equals, newPoolAddr.String())
		// given THORNode on
		if item.Coin.Asset.Equals(poolBNB.Asset) {
			// there are four coins , BNB,TCAN-014,LOK-3C0 and RUNE
			c.Assert(item.Coin.Amount.Uint64(), Equals, poolBNB.BalanceAsset.Uint64()-uint64(poolGas))
		}
		if item.Coin.Asset.Equals(poolTCan.Asset) {
			c.Assert(item.Coin.Amount.Uint64(), Equals, uint64(1535169738538008), Commentf("%d", item.Coin.Amount.Uint64()))
		}
		if item.Coin.Asset.Equals(poolLoki.Asset) {
			c.Check(item.Coin.Amount.Uint64(), Equals, uint64(1535169738538008), Commentf("%d", item.Coin.Amount.Uint64()))
		}
		if item.Coin.Asset.IsRune() {
			c.Assert(item.Coin.Amount.String(), Equals, "4605519215614024")
		}
	}
	w.txOutStore.CommitBlock(w.ctx)
}

func createTempNewPoolForTest(ctx sdk.Context, k Keeper, input string, c *C) *Pool {
	p := NewPool()
	asset, err := common.NewAsset(input)
	c.Assert(err, IsNil)
	p.Asset = asset
	// limiting balance to 59 bits, because the math done with floats looses
	// precision if the number is greater than 59 bits.
	// https://stackoverflow.com/questions/30897208/how-to-change-a-float64-number-to-uint64-in-a-right-way
	// https://github.com/golang/go/issues/29463
	p.BalanceRune = sdk.NewUint(1535169738538008)
	p.BalanceAsset = sdk.NewUint(1535169738538008)
	c.Assert(k.SetPool(ctx, p), IsNil)
	k.SetChains(ctx, common.Chains{asset.Chain})
	return &p
}
