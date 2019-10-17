package swapservice

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type PoolAddressManagerSuite struct{}

var _ = Suite(&PoolAddressManagerSuite{})

func (ps *PoolAddressManagerSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

func (PoolAddressManagerSuite) TestSetupInitialPoolAddresses(c *C) {
	ctx, k := setupKeeperForTest(c)
	poolAddrMgr := NewPoolAddressManager(k)
	c.Assert(poolAddrMgr, NotNil)
	// incorrect block height
	pa, err := poolAddrMgr.setupInitialPoolAddresses(ctx, 0)
	c.Assert(err, NotNil)
	c.Assert(pa.IsEmpty(), Equals, true)

	pa1, err := poolAddrMgr.setupInitialPoolAddresses(ctx, 1)
	c.Assert(err, NotNil)
	c.Assert(pa1.IsEmpty(), Equals, true)

	bnb, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)
	addr, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	c.Check(addr.Empty(), Equals, false)
	bepConsPubKey := `bepcpub1zcjduepq4kn64fcjhf0fp20gp8var0rm25ca9jy6jz7acem8gckh0nkplznq85gdrg`
	trustAccount := NewTrustAccount(bnb, addr, bepConsPubKey)
	err = trustAccount.IsValid()
	c.Assert(err, IsNil)
	nodeAddress, err := sdk.AccAddressFromBech32("bep1rtgz3lcaw8vw0yfsc8ga0rdgwa3qh9ju7vfsnk")
	c.Assert(err, IsNil)
	na := NewNodeAccount(nodeAddress, NodeActive, trustAccount)
	k.SetNodeAccount(ctx, na)

	pa2, err := poolAddrMgr.setupInitialPoolAddresses(ctx, 1)
	c.Assert(err, IsNil)
	c.Assert(pa2.IsEmpty(), Equals, false)
	c.Assert(pa2.Current.String(), Equals, bnb.String())
	c.Assert(pa2.Next.String(), Equals, bnb.String())

	// Two nodes
	na1 := GetRandomNodeAccount(NodeActive)
	k.SetNodeAccount(ctx, na1)

	// with two active nodes
	pa3, err := poolAddrMgr.setupInitialPoolAddresses(ctx, 1)
	c.Assert(err, IsNil)
	c.Assert(pa3.IsEmpty(), Equals, false)
	c.Assert(pa3.Current.String(), Equals, bnb.String())
	c.Assert(pa3.Next.String(), Equals, na1.Accounts.SignerBNBAddress.String())

	nodeAccounts := NodeAccounts{na1}
	// with more than two  active nodes
	for i := 0; i < 10; i++ {
		na2 := GetRandomNodeAccount(NodeActive)
		k.SetNodeAccount(ctx, na2)
		nodeAccounts = append(nodeAccounts, na2)
	}

	sort.Sort(nodeAccounts)

	pa4, err := poolAddrMgr.setupInitialPoolAddresses(ctx, 1)
	c.Assert(err, IsNil)
	c.Assert(pa4.IsEmpty(), Equals, false)
	c.Assert(pa4.Current.String(), Equals, bnb.String())
	c.Assert(pa4.Next.String(), Equals, nodeAccounts[0].Accounts.SignerBNBAddress.String())
	c.Logf("%+v", pa4)
	rotatePerBlockHeight := k.GetAdminConfigRotatePerBlockHeight(ctx, sdk.AccAddress{})
	rotateAt := rotatePerBlockHeight + 1
	txOutStore := NewTxOutStore(&MockTxOutSetter{})
	txOutStore.NewBlock(uint64(rotateAt))
	newPa := poolAddrMgr.rotatePoolAddress(ctx, rotateAt, pa4, txOutStore)
	c.Assert(newPa.IsEmpty(), Equals, false)
	c.Assert(newPa.Previous.String(), Equals, pa4.Current.String())
	c.Assert(newPa.Current.String(), Equals, pa4.Next.String())
	c.Assert(newPa.Next.String(), Equals, nodeAccounts[1].Accounts.SignerBNBAddress.String())
	c.Assert(newPa.RotateAt, Equals, int64(rotatePerBlockHeight*2+1))
	poolBNB := createTempNewPoolForTest(ctx, k, "BNB", c)
	poolTCan := createTempNewPoolForTest(ctx, k, "TCAN-014", c)
	poolLoki := createTempNewPoolForTest(ctx, k, "LOK-3C0", c)

	newPa1 := poolAddrMgr.rotatePoolAddress(ctx, rotatePerBlockHeight*2+1, newPa, txOutStore)
	c.Logf("new pool addresses %+v", newPa1)
	c.Assert(newPa1.IsEmpty(), Equals, false)
	c.Assert(newPa1.Previous.String(), Equals, newPa.Current.String())
	c.Assert(newPa1.Current.String(), Equals, newPa.Next.String())
	c.Assert(newPa1.Next.String(), Equals, nodeAccounts[2].Accounts.SignerBNBAddress.String())
	c.Assert(newPa1.RotateAt, Equals, int64(rotatePerBlockHeight*3+1))
	c.Assert(len(txOutStore.blockOut.TxArray) > 0, Equals, true)
	c.Assert(txOutStore.blockOut.Valid(), IsNil)
	for _, item := range txOutStore.blockOut.TxArray {
		c.Assert(item.Valid(), IsNil)
		// make sure the fund is sending from previous pool address to current
		c.Assert(item.ToAddress.String(), Equals, newPa1.Current.String())
		c.Assert(len(item.Coins) > 0, Equals, true)
		if item.Coins[0].Denom == poolBNB.Ticker {
			c.Assert(item.Coins[0].Amount.Uint64(), Equals, poolBNB.BalanceToken.Uint64()-batchTransactionFee)
		}
		if item.Coins[0].Denom.String() == poolTCan.Ticker.String() {
			c.Assert(item.Coins[0].Amount.Uint64(), Equals, poolTCan.BalanceToken.Uint64()-batchTransactionFee)
		}
		if item.Coins[0].Denom.String() == poolLoki.Ticker.String() {
			c.Check(item.Coins[0].Amount.Uint64(), Equals, poolLoki.BalanceToken.Uint64()-batchTransactionFee)
		}
		if common.IsRune(item.Coins[0].Denom) {
			totalRune := poolBNB.BalanceRune.Add(poolLoki.BalanceRune).Add(poolTCan.BalanceRune)
			c.Assert(item.Coins[0].Amount.String(), Equals, totalRune.SubUint64(batchTransactionFee).String())
		}
	}

}

func createTempNewPoolForTest(ctx sdk.Context, k Keeper, ticker string, c *C) *Pool {
	p := NewPool()
	t, err := common.NewTicker(ticker)
	c.Assert(err, IsNil)
	p.Ticker = t
	// limiting balance to 59 bits, because the math done with floats looses
	// precision if the number is greater than 59 bits.
	// https://stackoverflow.com/questions/30897208/how-to-change-a-float64-number-to-uint64-in-a-right-way
	// https://github.com/golang/go/issues/29463
	p.BalanceRune = sdk.NewUint(1535169738538008)
	p.BalanceToken = sdk.NewUint(1535169738538008)
	k.SetPool(ctx, p)
	return &p
}
