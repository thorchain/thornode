package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type HelperSuite struct{}

var _ = Suite(&HelperSuite{})

type TestRefundBondKeeper struct {
	KVStoreDummy
	ygg    Vault
	pool   Pool
	na     NodeAccount
	vaults Vaults
}

func (k *TestRefundBondKeeper) GetAsgardVaultsByStatus(_ sdk.Context, _ VaultStatus) (Vaults, error) {
	return k.vaults, nil
}

func (k *TestRefundBondKeeper) GetVault(_ sdk.Context, pk common.PubKey) (Vault, error) {
	if k.ygg.PubKey.Equals(pk) {
		return k.ygg, nil
	}
	return Vault{}, kaboom
}

func (k *TestRefundBondKeeper) GetPool(_ sdk.Context, asset common.Asset) (Pool, error) {
	if k.pool.Asset.Equals(asset) {
		return k.pool, nil
	}
	return NewPool(), kaboom
}

func (k *TestRefundBondKeeper) SetNodeAccount(_ sdk.Context, na NodeAccount) error {
	k.na = na
	return nil
}

func (k *TestRefundBondKeeper) UpsertEvent(_ sdk.Context, e Event) error {
	return nil
}

func (k *TestRefundBondKeeper) SetPool(_ sdk.Context, p Pool) error {
	if k.pool.Asset.Equals(p.Asset) {
		k.pool = p
		return nil
	}
	return kaboom
}

func (k *TestRefundBondKeeper) DeleteVault(_ sdk.Context, key common.PubKey) error {
	if k.ygg.PubKey.Equals(key) {
		k.ygg = NewVault(1, InactiveVault, AsgardVault, GetRandomPubKey())
	}
	return nil
}

func (s *HelperSuite) TestSubsidizePoolWithSlashBond(c *C) {
	ctx, k := setupKeeperForTest(c)
	ygg := GetRandomVault()
	c.Assert(subsidizePoolWithSlashBond(ctx, k, ygg, sdk.NewUint(100*common.One), sdk.ZeroUint()), IsNil)
	poolBNB := NewPool()
	poolBNB.Asset = common.BNBAsset
	poolBNB.BalanceRune = sdk.NewUint(100 * common.One)
	poolBNB.BalanceAsset = sdk.NewUint(100 * common.One)
	poolBNB.Status = PoolEnabled
	c.Assert(k.SetPool(ctx, poolBNB), IsNil)

	poolTCAN := NewPool()
	tCanAsset, err := common.NewAsset("BNB.TCAN-014")
	c.Assert(err, IsNil)
	poolTCAN.Asset = tCanAsset
	poolTCAN.BalanceRune = sdk.NewUint(200 * common.One)
	poolTCAN.BalanceAsset = sdk.NewUint(200 * common.One)
	poolTCAN.Status = PoolEnabled
	c.Assert(k.SetPool(ctx, poolTCAN), IsNil)

	poolBTC := NewPool()
	poolBTC.Asset = common.BTCAsset
	poolBTC.BalanceAsset = sdk.NewUint(300 * common.One)
	poolBTC.BalanceRune = sdk.NewUint(300 * common.One)
	poolBTC.Status = PoolEnabled
	c.Assert(k.SetPool(ctx, poolBTC), IsNil)
	ygg.Type = YggdrasilVault
	ygg.Coins = common.Coins{
		common.NewCoin(common.RuneAsset(), sdk.NewUint(1*common.One)),
		common.NewCoin(common.BNBAsset, sdk.NewUint(1*common.One)),            // 1
		common.NewCoin(tCanAsset, sdk.NewUint(common.One).QuoUint64(2)),       // 0.5 TCAN
		common.NewCoin(common.BTCAsset, sdk.NewUint(common.One).QuoUint64(4)), // 0.25 BTC
	}
	totalRuneLeft, err := getTotalYggValueInRune(ctx, k, ygg)
	c.Assert(err, IsNil)

	totalRuneStolen := ygg.GetCoin(common.RuneAsset()).Amount
	slashAmt := totalRuneLeft.MulUint64(3).QuoUint64(2)
	c.Assert(subsidizePoolWithSlashBond(ctx, k, ygg, totalRuneLeft, slashAmt), IsNil)

	slashAmt = common.SafeSub(slashAmt, totalRuneStolen)
	totalRuneLeft = common.SafeSub(totalRuneLeft, totalRuneStolen)

	amountBNBForBNBPool := slashAmt.Mul(poolBNB.AssetValueInRune(sdk.NewUint(common.One))).Quo(totalRuneLeft)
	runeBNB := poolBNB.BalanceRune.Add(amountBNBForBNBPool)
	bnbPoolAsset := poolBNB.BalanceAsset.Sub(sdk.NewUint(common.One))
	poolBNB, err = k.GetPool(ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	c.Assert(poolBNB.BalanceRune.Equal(runeBNB), Equals, true)
	c.Assert(poolBNB.BalanceAsset.Equal(bnbPoolAsset), Equals, true)
	amountRuneForTCANPool := slashAmt.Mul(poolTCAN.AssetValueInRune(sdk.NewUint(common.One).QuoUint64(2))).Quo(totalRuneLeft)
	runeTCAN := poolTCAN.BalanceRune.Add(amountRuneForTCANPool)
	tcanPoolAsset := poolTCAN.BalanceAsset.Sub(sdk.NewUint(common.One).QuoUint64(2))
	poolTCAN, err = k.GetPool(ctx, tCanAsset)
	c.Assert(err, IsNil)
	c.Assert(poolTCAN.BalanceRune.Equal(runeTCAN), Equals, true)
	c.Assert(poolTCAN.BalanceAsset.Equal(tcanPoolAsset), Equals, true)
	amountRuneForBTCPool := slashAmt.Mul(poolBTC.AssetValueInRune(sdk.NewUint(common.One).QuoUint64(4))).Quo(totalRuneLeft)
	runeBTC := poolBTC.BalanceRune.Add(amountRuneForBTCPool)
	btcPoolAsset := poolBTC.BalanceAsset.Sub(sdk.NewUint(common.One).QuoUint64(4))
	poolBTC, err = k.GetPool(ctx, common.BTCAsset)
	c.Assert(err, IsNil)
	c.Assert(poolBTC.BalanceRune.Equal(runeBTC), Equals, true)
	c.Assert(poolBTC.BalanceAsset.Equal(btcPoolAsset), Equals, true)

	ygg1 := GetRandomVault()
	ygg1.Type = YggdrasilVault
	ygg1.Coins = common.Coins{
		common.NewCoin(tCanAsset, sdk.NewUint(common.One*2)),       // 2 TCAN
		common.NewCoin(common.BTCAsset, sdk.NewUint(common.One*4)), // 4 BTC
	}
	totalRuneLeft, err = getTotalYggValueInRune(ctx, k, ygg1)
	c.Assert(err, IsNil)
	slashAmt = sdk.NewUint(100 * common.One)
	c.Assert(subsidizePoolWithSlashBond(ctx, k, ygg1, totalRuneLeft, slashAmt), IsNil)
	amountRuneForTCANPool = slashAmt.Mul(poolTCAN.AssetValueInRune(sdk.NewUint(common.One * 2))).Quo(totalRuneLeft)
	runeTCAN = poolTCAN.BalanceRune.Add(amountRuneForTCANPool)
	poolTCAN, err = k.GetPool(ctx, tCanAsset)
	c.Assert(err, IsNil)
	c.Assert(poolTCAN.BalanceRune.Equal(runeTCAN), Equals, true)
	amountRuneForBTCPool = slashAmt.Mul(poolBTC.AssetValueInRune(sdk.NewUint(common.One * 4))).Quo(totalRuneLeft)
	runeBTC = poolBTC.BalanceRune.Add(amountRuneForBTCPool)
	poolBTC, err = k.GetPool(ctx, common.BTCAsset)
	c.Assert(err, IsNil)
	c.Assert(poolBTC.BalanceRune.Equal(runeBTC), Equals, true)
}

func (s *HelperSuite) TestRefundBondError(c *C) {
	ctx, _ := setupKeeperForTest(c)
	// active node should not refund bond
	pk := GetRandomPubKey()
	na := GetRandomNodeAccount(NodeActive)
	na.PubKeySet.Secp256k1 = pk
	na.Bond = sdk.NewUint(100 * common.One)
	txOut := NewTxStoreDummy()
	tx := GetRandomTx()
	keeper1 := &TestRefundBondKeeper{}
	c.Assert(refundBond(ctx, tx, na, keeper1, txOut), IsNil)

	// fail to get vault should return an error
	na.UpdateStatus(NodeStandby, ctx.BlockHeight())
	keeper1.na = na
	c.Assert(refundBond(ctx, tx, na, keeper1, txOut), NotNil)

	// if the vault is not a yggdrasil pool , it should return an error
	ygg := NewVault(ctx.BlockHeight(), ActiveVault, AsgardVault, pk)
	ygg.Coins = common.Coins{}
	keeper1.ygg = ygg
	c.Assert(refundBond(ctx, tx, na, keeper1, txOut), NotNil)

	// fail to get pool should fail
	ygg = NewVault(ctx.BlockHeight(), ActiveVault, YggdrasilVault, pk)
	ygg.Coins = common.Coins{
		common.NewCoin(common.RuneAsset(), sdk.NewUint(27*common.One)),
		common.NewCoin(common.BNBAsset, sdk.NewUint(27*common.One)),
	}
	keeper1.ygg = ygg
	c.Assert(refundBond(ctx, tx, na, keeper1, txOut), NotNil)

	// when ygg asset in RUNE is more then bond , thorchain should slash the node account with all their bond
	keeper1.pool = Pool{
		Asset:        common.BNBAsset,
		BalanceRune:  sdk.NewUint(1024 * common.One),
		BalanceAsset: sdk.NewUint(167 * common.One),
	}
	c.Assert(refundBond(ctx, tx, na, keeper1, txOut), IsNil)
	// make sure no tx has been generated for refund
	c.Assert(txOut.GetOutboundItems(), HasLen, 0)
}

func (s *HelperSuite) TestRefundBondHappyPath(c *C) {
	ctx, _ := setupKeeperForTest(c)
	na := GetRandomNodeAccount(NodeActive)
	na.Bond = sdk.NewUint(12098 * common.One)
	txOut := NewTxStoreDummy()
	pk := GetRandomPubKey()
	na.PubKeySet.Secp256k1 = pk
	ygg := NewVault(ctx.BlockHeight(), ActiveVault, YggdrasilVault, pk)

	ygg.Coins = common.Coins{
		common.NewCoin(common.RuneAsset(), sdk.NewUint(3946*common.One)),
		common.NewCoin(common.BNBAsset, sdk.NewUint(27*common.One)),
	}
	keeper := &TestRefundBondKeeper{
		pool: Pool{
			Asset:        common.BNBAsset,
			BalanceRune:  sdk.NewUint(23789 * common.One),
			BalanceAsset: sdk.NewUint(167 * common.One),
		},
		ygg:    ygg,
		vaults: Vaults{GetRandomVault()},
	}
	na.Status = NodeStandby
	tx := GetRandomTx()
	yggAssetInRune, err := getTotalYggValueInRune(ctx, keeper, ygg)
	c.Assert(err, IsNil)
	err = refundBond(ctx, tx, na, keeper, txOut)
	slashAmt := yggAssetInRune.MulUint64(3).QuoUint64(2)
	c.Assert(err, IsNil)
	c.Assert(txOut.GetOutboundItems(), HasLen, 1)
	outCoin := txOut.GetOutboundItems()[0].Coin
	c.Check(outCoin.Amount.Equal(sdk.NewUint(40981137725)), Equals, true)
	p, err := keeper.GetPool(ctx, common.BNBAsset)
	c.Assert(err, IsNil)
	expectedPoolRune := sdk.NewUint(23789 * common.One).Sub(sdk.NewUint(3946 * common.One)).Add(slashAmt)
	c.Assert(p.BalanceRune.Equal(expectedPoolRune), Equals, true, Commentf("expect %s however we got %s", expectedPoolRune, p.BalanceRune))
	expectedPoolBNB := sdk.NewUint(167 * common.One).Sub(sdk.NewUint(27 * common.One))
	c.Assert(p.BalanceAsset.Equal(expectedPoolBNB), Equals, true, Commentf("expected BNB in pool %s , however we got %s", expectedPoolBNB, p.BalanceAsset))
}

func (s *HelperSuite) TestEnableNextPool(c *C) {
	var err error
	ctx, k := setupKeeperForTest(c)

	pool := NewPool()
	pool.Asset = common.BNBAsset
	pool.Status = PoolEnabled
	pool.BalanceRune = sdk.NewUint(100 * common.One)
	pool.BalanceAsset = sdk.NewUint(100 * common.One)
	c.Assert(k.SetPool(ctx, pool), IsNil)

	pool = NewPool()
	pool.Asset = common.BTCAsset
	pool.Status = PoolBootstrap
	pool.BalanceRune = sdk.NewUint(50 * common.One)
	pool.BalanceAsset = sdk.NewUint(50 * common.One)
	c.Assert(k.SetPool(ctx, pool), IsNil)

	ethAsset, err := common.NewAsset("ETH.ETH")
	c.Assert(err, IsNil)
	pool = NewPool()
	pool.Asset = ethAsset
	pool.Status = PoolBootstrap
	pool.BalanceRune = sdk.NewUint(40 * common.One)
	pool.BalanceAsset = sdk.NewUint(40 * common.One)
	c.Assert(k.SetPool(ctx, pool), IsNil)

	xmrAsset, err := common.NewAsset("XMR.XMR")
	c.Assert(err, IsNil)
	pool = NewPool()
	pool.Asset = xmrAsset
	pool.Status = PoolBootstrap
	pool.BalanceRune = sdk.NewUint(40 * common.One)
	pool.BalanceAsset = sdk.NewUint(0 * common.One)
	c.Assert(k.SetPool(ctx, pool), IsNil)

	// usdAsset
	usdAsset, err := common.NewAsset("BNB.TUSDB")
	c.Assert(err, IsNil)
	pool = NewPool()
	pool.Asset = usdAsset
	pool.Status = PoolBootstrap
	pool.BalanceRune = sdk.NewUint(140 * common.One)
	pool.BalanceAsset = sdk.NewUint(0 * common.One)
	c.Assert(k.SetPool(ctx, pool), IsNil)
	// should enable BTC
	c.Assert(enableNextPool(ctx, k), IsNil)
	pool, err = k.GetPool(ctx, common.BTCAsset)
	c.Check(pool.Status, Equals, PoolEnabled)

	// should enable ETH
	c.Assert(enableNextPool(ctx, k), IsNil)
	pool, err = k.GetPool(ctx, ethAsset)
	c.Check(pool.Status, Equals, PoolEnabled)

	// should NOT enable XMR, since it has no assets
	c.Assert(enableNextPool(ctx, k), IsNil)
	pool, err = k.GetPool(ctx, xmrAsset)
	c.Assert(pool.Empty(), Equals, false)
	c.Check(pool.Status, Equals, PoolBootstrap)
}
