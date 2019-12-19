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
	ygg  Vault
	pool Pool
	na   NodeAccount
}

func (k *TestRefundBondKeeper) GetVault(_ sdk.Context, _ common.PubKey) (Vault, error) {
	return k.ygg, nil
}

func (k *TestRefundBondKeeper) GetPool(_ sdk.Context, _ common.Asset) (Pool, error) {
	return k.pool, nil
}

func (k *TestRefundBondKeeper) SetNodeAccount(_ sdk.Context, na NodeAccount) error {
	k.na = na
	return nil
}

func (s *HelperSuite) TestRefundBond(c *C) {
	ctx, _ := setupKeeperForTest(c)
	txID := GetRandomTxHash()
	na := GetRandomNodeAccount(NodeActive)
	na.Bond = sdk.NewUint(12098 * common.One)
	txOut := NewTxStoreDummy()

	pk := GetRandomPubKey()
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
		ygg: ygg,
	}

	err := refundBond(ctx, txID, na, keeper, txOut)
	c.Assert(err, IsNil)
	c.Assert(txOut.GetOutboundItems(), HasLen, 1)
	outCoin := txOut.GetOutboundItems()[0].Coin
	c.Check(outCoin.Amount.Equal(sdk.NewUint(430587425150)), Equals, true)
}

func (s *HelperSuite) TestEnableNextPool(c *C) {
	var err error
	ctx, k := setupKeeperForTest(c)

	pool := NewPool()
	pool.Asset = common.BNBAsset
	pool.Status = PoolEnabled
	pool.BalanceRune = sdk.NewUint(100 * common.One)
	pool.BalanceAsset = sdk.NewUint(100 * common.One)
	k.SetPool(ctx, pool)

	pool = NewPool()
	pool.Asset = common.BTCAsset
	pool.Status = PoolBootstrap
	pool.BalanceRune = sdk.NewUint(50 * common.One)
	pool.BalanceAsset = sdk.NewUint(50 * common.One)
	k.SetPool(ctx, pool)

	ethAsset, err := common.NewAsset("ETH.ETH")
	c.Assert(err, IsNil)
	pool = NewPool()
	pool.Asset = ethAsset
	pool.Status = PoolBootstrap
	pool.BalanceRune = sdk.NewUint(40 * common.One)
	pool.BalanceAsset = sdk.NewUint(40 * common.One)
	k.SetPool(ctx, pool)

	xmrAsset, err := common.NewAsset("XMR.XMR")
	c.Assert(err, IsNil)
	pool = NewPool()
	pool.Asset = xmrAsset
	pool.Status = PoolBootstrap
	pool.BalanceRune = sdk.NewUint(40 * common.One)
	pool.BalanceAsset = sdk.NewUint(0 * common.One)
	k.SetPool(ctx, pool)

	// should enable BTC
	enableNextPool(ctx, k)
	pool, err = k.GetPool(ctx, common.BTCAsset)
	c.Check(pool.Status, Equals, PoolEnabled)

	// should enable ETH
	enableNextPool(ctx, k)
	pool, err = k.GetPool(ctx, ethAsset)
	c.Check(pool.Status, Equals, PoolEnabled)

	// should NOT enable XMR, since it has no assets
	enableNextPool(ctx, k)
	pool, err = k.GetPool(ctx, xmrAsset)
	c.Check(pool.Status, Equals, PoolBootstrap)
}
