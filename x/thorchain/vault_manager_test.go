package thorchain

import (
	"errors"

	. "gopkg.in/check.v1"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type VaultManagerTestSuite struct{}

var _ = Suite(&VaultManagerTestSuite{})

func (s *VaultManagerTestSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

type TestRagnarokChainKeeper struct {
	KVStoreDummy
	activeVault Vault
	retireVault Vault
	yggVault    Vault
	pools       Pools
	ps          PoolStaker
	sp          StakerPool
	na          NodeAccount
	err         error
}

func (k *TestRagnarokChainKeeper) ListNodeAccounts(_ sdk.Context) (NodeAccounts, error) {
	return NodeAccounts{k.na}, k.err
}

func (k *TestRagnarokChainKeeper) ListActiveNodeAccounts(_ sdk.Context) (NodeAccounts, error) {
	return NodeAccounts{k.na}, k.err
}

func (k *TestRagnarokChainKeeper) GetAsgardVaultsByStatus(_ sdk.Context, vt VaultStatus) (Vaults, error) {
	if vt == ActiveVault {
		return Vaults{k.activeVault}, k.err
	}
	return Vaults{k.retireVault}, k.err
}

func (k *TestRagnarokChainKeeper) VaultExists(_ sdk.Context, _ common.PubKey) bool {
	return true
}

func (k *TestRagnarokChainKeeper) GetVault(_ sdk.Context, _ common.PubKey) (Vault, error) {
	return k.yggVault, k.err
}

func (k *TestRagnarokChainKeeper) GetPools(_ sdk.Context) (Pools, error) {
	return k.pools, k.err
}

func (k *TestRagnarokChainKeeper) GetPool(_ sdk.Context, asset common.Asset) (Pool, error) {
	for _, pool := range k.pools {
		if pool.Asset.Equals(asset) {
			return pool, nil
		}
	}
	return Pool{}, errors.New("pool not found")
}

func (k *TestRagnarokChainKeeper) SetPool(_ sdk.Context, pool Pool) error {
	for i, p := range k.pools {
		if p.Asset.Equals(pool.Asset) {
			k.pools[i] = pool
		}
	}
	return k.err
}

func (k *TestRagnarokChainKeeper) PoolExist(_ sdk.Context, _ common.Asset) bool {
	return true
}

func (k *TestRagnarokChainKeeper) GetPoolStaker(_ sdk.Context, asset common.Asset) (PoolStaker, error) {
	if asset.Equals(common.BTCAsset) {
		return k.ps, k.err
	}
	return PoolStaker{}, k.err
}

func (k *TestRagnarokChainKeeper) SetPoolStaker(_ sdk.Context, ps PoolStaker) {
	k.ps = ps
}

func (k *TestRagnarokChainKeeper) GetStakerPool(_ sdk.Context, addr common.Address) (StakerPool, error) {
	return k.sp, k.err
}

func (k *TestRagnarokChainKeeper) SetStakerPool(_ sdk.Context, sp StakerPool) {
	k.sp = sp
}

func (k *TestRagnarokChainKeeper) GetGas(_ sdk.Context, _ common.Asset) ([]sdk.Uint, error) {
	return []sdk.Uint{sdk.NewUint(10)}, k.err
}

func (k *TestRagnarokChainKeeper) GetLowestActiveVersion(_ sdk.Context) semver.Version {
	return constants.SWVersion
}

func (k *TestRagnarokChainKeeper) AddFeeToReserve(_ sdk.Context, _ sdk.Uint) error {
	return k.err
}

func (k *TestRagnarokChainKeeper) UpsertEvent(_ sdk.Context, _ Event) error {
	return k.err
}

func (k *TestRagnarokChainKeeper) IsActiveObserver(_ sdk.Context, _ sdk.AccAddress) bool {
	return true
}

func (s *ValidatorManagerTestSuite) TestRagnarokChain(c *C) {
	ctx, _ := setupKeeperForTest(c)
	ctx = ctx.WithBlockHeight(100000)
	ver := constants.SWVersion
	constAccessor := constants.GetConstantValues(ver)

	activeVault := GetRandomVault()
	retireVault := GetRandomVault()
	retireVault.Chains = common.Chains{common.BNBChain, common.BTCChain}
	yggVault := GetRandomVault()
	yggVault.Type = YggdrasilVault
	yggVault.Coins = common.Coins{
		common.NewCoin(common.BTCAsset, sdk.NewUint(3*common.One)),
		common.NewCoin(common.RuneAsset(), sdk.NewUint(300*common.One)),
	}

	btcPool := NewPool()
	btcPool.Asset = common.BTCAsset
	btcPool.BalanceRune = sdk.NewUint(1000 * common.One)
	btcPool.BalanceAsset = sdk.NewUint(10 * common.One)
	btcPool.PoolUnits = sdk.NewUint(1600)

	bnbPool := NewPool()
	bnbPool.Asset = common.BNBAsset
	bnbPool.BalanceRune = sdk.NewUint(1000 * common.One)
	bnbPool.BalanceAsset = sdk.NewUint(10 * common.One)
	bnbPool.PoolUnits = sdk.NewUint(1600)

	ps := NewPoolStaker(common.BTCAsset, sdk.NewUint(1600))
	addr := GetRandomBNBAddress()
	ps.Stakers = []StakerUnit{
		StakerUnit{
			RuneAddress: addr,
			Height:      5,
			Units:       ps.TotalUnits.QuoUint64(2),
			PendingRune: sdk.ZeroUint(),
		},
		StakerUnit{
			RuneAddress: GetRandomBNBAddress(),
			Height:      10,
			Units:       ps.TotalUnits.QuoUint64(2),
			PendingRune: sdk.ZeroUint(),
		},
	}
	sp := NewStakerPool(addr)

	keeper := &TestRagnarokChainKeeper{
		na:          GetRandomNodeAccount(NodeActive),
		activeVault: activeVault,
		retireVault: retireVault,
		yggVault:    yggVault,
		pools:       Pools{bnbPool, btcPool},
		ps:          ps,
		sp:          sp,
	}

	versionedTxOutStoreDummy := NewVersionedTxOutStoreDummy()
	vaultMgr := NewVaultMgr(keeper, versionedTxOutStoreDummy)

	err := vaultMgr.manageChains(ctx, constAccessor)
	c.Assert(err, IsNil)
	c.Check(keeper.pools[1].Asset.Equals(common.BTCAsset), Equals, true)
	c.Check(keeper.pools[1].PoolUnits.IsZero(), Equals, true, Commentf("%d\n", keeper.pools[1].PoolUnits.Uint64()))
	c.Check(keeper.pools[0].PoolUnits.Equal(sdk.NewUint(1600)), Equals, true)
	c.Check(keeper.ps.Stakers, HasLen, 0)

	// ensure we have requested for ygg funds to be returned
	txOutStore, err := versionedTxOutStoreDummy.GetTxOutStore(keeper, constants.SWVersion)
	c.Assert(err, IsNil)
	items, err := txOutStore.GetOutboundItems(ctx)
	c.Assert(err, IsNil)

	// 1 ygg return + 4 unstakes
	c.Check(items, HasLen, 5, Commentf("Len %d", items))
	c.Check(items[0].Memo, Equals, NewYggdrasilReturn(ctx.BlockHeight()).String())
	c.Check(items[0].Chain.Equals(common.BTCChain), Equals, true)
}
