package thorchain

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
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

	pk1, err := common.NewPoolPubKey(common.BNBChain, 0, GetRandomPubKey())
	c.Assert(err, IsNil)
	w.poolAddrMgr.GetCurrentPoolAddresses().Next = common.PoolPubKeys{pk1}
	// no asset get moved , because THORNode just opened window, however THORNode should instruct signer to kick off key sign process
	c.Assert(w.txOutStore.GetOutboundItems(), HasLen, 1)
	poolBNB := createTempNewPoolForTest(w.ctx, w.keeper, "BNB.BNB", c)
	poolTCan := createTempNewPoolForTest(w.ctx, w.keeper, "BNB.TCAN-014", c)
	poolLoki := createTempNewPoolForTest(w.ctx, w.keeper, "BNB.LOK-3C0", c)
	c.Assert(w.txOutStore.GetOutboundItems(), HasLen, 4)
	c.Assert(w.txOutStore.GetBlockOut().Valid(), IsNil)
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
	p.BalanceRune = sdk.NewUint(1535169738538008)
	p.BalanceAsset = sdk.NewUint(1535169738538008)
	c.Assert(k.SetPool(ctx, p), IsNil)
	k.SetChains(ctx, common.Chains{asset.Chain})
	return &p
}
