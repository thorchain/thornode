package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
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

func (k *TestRefundBondKeeper) VaultExists(_ sdk.Context, pk common.PubKey) bool {
	return true
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
		k.ygg = NewVault(1, InactiveVault, AsgardVault, GetRandomPubKey(), common.Chains{common.BNBChain})
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
	eventMgr := NewEventMgr()
	keeper1 := &TestRefundBondKeeper{}
	c.Assert(refundBond(ctx, tx, na, keeper1, txOut, eventMgr), IsNil)

	// fail to get vault should return an error
	na.UpdateStatus(NodeStandby, ctx.BlockHeight())
	keeper1.na = na
	c.Assert(refundBond(ctx, tx, na, keeper1, txOut, eventMgr), NotNil)

	// if the vault is not a yggdrasil pool , it should return an error
	ygg := NewVault(ctx.BlockHeight(), ActiveVault, AsgardVault, pk, common.Chains{common.BNBChain})
	ygg.Coins = common.Coins{}
	keeper1.ygg = ygg
	c.Assert(refundBond(ctx, tx, na, keeper1, txOut, eventMgr), NotNil)

	// fail to get pool should fail
	ygg = NewVault(ctx.BlockHeight(), ActiveVault, YggdrasilVault, pk, common.Chains{common.BNBChain})
	ygg.Coins = common.Coins{
		common.NewCoin(common.RuneAsset(), sdk.NewUint(27*common.One)),
		common.NewCoin(common.BNBAsset, sdk.NewUint(27*common.One)),
	}
	keeper1.ygg = ygg
	c.Assert(refundBond(ctx, tx, na, keeper1, txOut, eventMgr), NotNil)

	// when ygg asset in RUNE is more then bond , thorchain should slash the node account with all their bond
	keeper1.pool = Pool{
		Asset:        common.BNBAsset,
		BalanceRune:  sdk.NewUint(1024 * common.One),
		BalanceAsset: sdk.NewUint(167 * common.One),
	}
	c.Assert(refundBond(ctx, tx, na, keeper1, txOut, eventMgr), IsNil)
	// make sure no tx has been generated for refund
	items, err := txOut.GetOutboundItems(ctx)
	c.Assert(err, IsNil)
	c.Check(items, HasLen, 0)
}

func (s *HelperSuite) TestRefundBondHappyPath(c *C) {
	ctx, _ := setupKeeperForTest(c)
	na := GetRandomNodeAccount(NodeActive)
	na.Bond = sdk.NewUint(12098 * common.One)
	txOut := NewTxStoreDummy()
	eventMgr := NewEventMgr()
	pk := GetRandomPubKey()
	na.PubKeySet.Secp256k1 = pk
	ygg := NewVault(ctx.BlockHeight(), ActiveVault, YggdrasilVault, pk, common.Chains{common.BNBChain})

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
	err = refundBond(ctx, tx, na, keeper, txOut, eventMgr)
	slashAmt := yggAssetInRune.MulUint64(3).QuoUint64(2)
	c.Assert(err, IsNil)
	items, err := txOut.GetOutboundItems(ctx)
	c.Assert(err, IsNil)
	c.Check(items, HasLen, 1)
	outCoin := items[0].Coin
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
	versionedEventManagerDummy := NewDummyVersionedEventMgr()
	eventMgr, err := versionedEventManagerDummy.GetEventManager(ctx, semver.MustParse("0.1.0"))
	c.Assert(err, IsNil)
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
	c.Assert(enableNextPool(ctx, k, eventMgr), IsNil)
	pool, err = k.GetPool(ctx, common.BTCAsset)
	c.Check(pool.Status, Equals, PoolEnabled)

	// should enable ETH
	c.Assert(enableNextPool(ctx, k, eventMgr), IsNil)
	pool, err = k.GetPool(ctx, ethAsset)
	c.Check(pool.Status, Equals, PoolEnabled)

	// should NOT enable XMR, since it has no assets
	c.Assert(enableNextPool(ctx, k, eventMgr), IsNil)
	pool, err = k.GetPool(ctx, xmrAsset)
	c.Assert(pool.Empty(), Equals, false)
	c.Check(pool.Status, Equals, PoolBootstrap)
}

type addGasFeesKeeperHelper struct {
	Keeper
	errGetVaultData bool
	errSetVaultData bool
	errGetPool      bool
	errSetPool      bool
	errSetEvent     bool
}

func newAddGasFeesKeeperHelper(keeper Keeper) *addGasFeesKeeperHelper {
	return &addGasFeesKeeperHelper{
		Keeper: keeper,
	}
}

func (h *addGasFeesKeeperHelper) GetVaultData(ctx sdk.Context) (VaultData, error) {
	if h.errGetVaultData {
		return VaultData{}, kaboom
	}
	return h.Keeper.GetVaultData(ctx)
}

func (h *addGasFeesKeeperHelper) SetVaultData(ctx sdk.Context, data VaultData) error {
	if h.errSetVaultData {
		return kaboom
	}
	return h.Keeper.SetVaultData(ctx, data)
}

func (h *addGasFeesKeeperHelper) SetPool(ctx sdk.Context, pool Pool) error {
	if h.errSetPool {
		return kaboom
	}
	return h.Keeper.SetPool(ctx, pool)
}

func (h *addGasFeesKeeperHelper) GetPool(ctx sdk.Context, asset common.Asset) (Pool, error) {
	if h.errGetPool {
		return Pool{}, kaboom
	}
	return h.Keeper.GetPool(ctx, asset)
}

func (h *addGasFeesKeeperHelper) UpsertEvent(ctx sdk.Context, event Event) error {
	if h.errSetEvent {
		return kaboom
	}
	return h.Keeper.UpsertEvent(ctx, event)
}

type addGasFeeTestHelper struct {
	ctx        sdk.Context
	k          *addGasFeesKeeperHelper
	na         NodeAccount
	gasManager GasManager
}

func newAddGasFeeTestHelper(c *C) addGasFeeTestHelper {
	ctx, k := setupKeeperForTest(c)
	keeper := newAddGasFeesKeeperHelper(k)
	pool := NewPool()
	pool.Asset = common.BNBAsset
	pool.BalanceAsset = sdk.NewUint(100 * common.One)
	pool.BalanceRune = sdk.NewUint(100 * common.One)
	pool.Status = PoolEnabled
	c.Assert(k.SetPool(ctx, pool), IsNil)

	poolBTC := NewPool()
	poolBTC.Asset = common.BTCAsset
	poolBTC.BalanceAsset = sdk.NewUint(100 * common.One)
	poolBTC.BalanceRune = sdk.NewUint(100 * common.One)
	poolBTC.Status = PoolEnabled
	c.Assert(k.SetPool(ctx, poolBTC), IsNil)

	na := GetRandomNodeAccount(NodeActive)
	c.Assert(k.SetNodeAccount(ctx, na), IsNil)
	yggVault := NewVault(ctx.BlockHeight(), ActiveVault, YggdrasilVault, na.PubKeySet.Secp256k1, common.Chains{common.BNBChain})
	c.Assert(k.SetVault(ctx, yggVault), IsNil)
	return addGasFeeTestHelper{
		ctx:        ctx,
		k:          keeper,
		na:         na,
		gasManager: NewGasMgr(),
	}
}

func (s *HelperSuite) TestAddGasFees(c *C) {
	testCases := []struct {
		name        string
		txCreator   func(helper addGasFeeTestHelper) ObservedTx
		runner      func(helper addGasFeeTestHelper, tx ObservedTx) error
		expectError bool
		validator   func(helper addGasFeeTestHelper, c *C)
	}{
		{
			name: "empty Gas should just return nil",
			txCreator: func(helper addGasFeeTestHelper) ObservedTx {
				return GetRandomObservedTx()
			},

			expectError: false,
		},
		{
			name: "normal BNB gas",
			txCreator: func(helper addGasFeeTestHelper) ObservedTx {
				tx := ObservedTx{
					Tx: common.Tx{
						ID:          GetRandomTxHash(),
						Chain:       common.BNBChain,
						FromAddress: GetRandomBNBAddress(),
						ToAddress:   GetRandomBNBAddress(),
						Coins: common.Coins{
							common.NewCoin(common.BNBAsset, sdk.NewUint(5*common.One)),
							common.NewCoin(common.RuneAsset(), sdk.NewUint(8*common.One)),
						},
						Gas: common.Gas{
							common.NewCoin(common.BNBAsset, BNBGasFeeSingleton[0].Amount),
						},
						Memo: "",
					},
					Status:         types.Done,
					OutHashes:      nil,
					BlockHeight:    helper.ctx.BlockHeight(),
					Signers:        []sdk.AccAddress{helper.na.NodeAddress},
					ObservedPubKey: helper.na.PubKeySet.Secp256k1,
				}
				return tx
			},
			runner: func(helper addGasFeeTestHelper, tx ObservedTx) error {
				return AddGasFees(helper.ctx, helper.k, tx, helper.gasManager)
			},
			expectError: false,
			validator: func(helper addGasFeeTestHelper, c *C) {
				expected := common.NewCoin(common.BNBAsset, BNBGasFeeSingleton[0].Amount)
				c.Assert(helper.gasManager.GetGas(), HasLen, 1)
				c.Assert(helper.gasManager.GetGas()[0].Equals(expected), Equals, true)
			},
		},
		{
			name: "normal BTC gas",
			txCreator: func(helper addGasFeeTestHelper) ObservedTx {
				tx := ObservedTx{
					Tx: common.Tx{
						ID:          GetRandomTxHash(),
						Chain:       common.BTCChain,
						FromAddress: GetRandomBTCAddress(),
						ToAddress:   GetRandomBTCAddress(),
						Coins: common.Coins{
							common.NewCoin(common.BTCAsset, sdk.NewUint(5*common.One)),
						},
						Gas: common.Gas{
							common.NewCoin(common.BTCAsset, sdk.NewUint(2000)),
						},
						Memo: "",
					},
					Status:         types.Done,
					OutHashes:      nil,
					BlockHeight:    helper.ctx.BlockHeight(),
					Signers:        []sdk.AccAddress{helper.na.NodeAddress},
					ObservedPubKey: helper.na.PubKeySet.Secp256k1,
				}
				return tx
			},
			runner: func(helper addGasFeeTestHelper, tx ObservedTx) error {
				return AddGasFees(helper.ctx, helper.k, tx, helper.gasManager)
			},
			expectError: false,
			validator: func(helper addGasFeeTestHelper, c *C) {
				expected := common.NewCoin(common.BTCAsset, sdk.NewUint(2000))
				c.Assert(helper.gasManager.GetGas(), HasLen, 1)
				c.Assert(helper.gasManager.GetGas()[0].Equals(expected), Equals, true)
			},
		},
	}
	for _, tc := range testCases {
		helper := newAddGasFeeTestHelper(c)
		tx := tc.txCreator(helper)
		var err error
		if tc.runner == nil {
			err = AddGasFees(helper.ctx, helper.k, tx, helper.gasManager)
		} else {
			err = tc.runner(helper, tx)
		}

		if err != nil && !tc.expectError {
			c.Errorf("test case: %s,didn't expect error however it got : %s", tc.name, err)
			c.FailNow()
		}
		if err == nil && tc.expectError {
			c.Errorf("test case: %s, expect error however it didn't", tc.name)
			c.FailNow()
		}
		if !tc.expectError && tc.validator != nil {
			tc.validator(helper, c)
			continue
		}
	}
}
