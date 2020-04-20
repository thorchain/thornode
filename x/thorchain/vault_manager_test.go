package thorchain

import (
	"errors"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
	. "gopkg.in/check.v1"
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
	pools       Pools
	ps          PoolStaker
	na          NodeAccount
	err         error
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

func (k *TestRagnarokChainKeeper) PoolExist(_ sdk.Context, _ common.Asset) bool {
	return true
}

func (k *TestRagnarokChainKeeper) GetPoolStaker(_ sdk.Context, asset common.Asset) (PoolStaker, error) {
	if asset.Equals(common.BTCAsset) {
		return k.ps, k.err
	}
	return PoolStaker{}, k.err
}

func (k *TestRagnarokChainKeeper) GetStakerPool(_ sdk.Context, addr common.Address) (StakerPool, error) {
	if asset.Equals(common.BTCAsset) {
		return k.ps, k.err
	}
	return PoolStaker{}, k.err
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

func (s *ValidatorManagerTestSuite) TestRagnarokChain(c *C) {
	ctx, _ := setupKeeperForTest(c)

	activeVault := GetRandomVault()
	retireVault := GetRandomVault()
	retireVault.Chains = common.Chains{common.BNBChain, common.BTCChain}

	btcPool := NewPool()
	btcPool.Asset = common.BTCAsset
	btcPool.BalanceRune = sdk.NewUint(1000 * common.One)
	btcPool.BalanceAsset = sdk.NewUint(10 * common.One)
	btcPool.PoolUnits = sdk.NewUint(1600)

	ps := NewPoolStaker(common.BTCAsset, sdk.NewUint(1600))
	addr := GetRandomBNBAddress()
	ps.Stakers = []StakerUnit{
		StakerUnit{
			RuneAddress: addr,
			Height:      5,
			Units:       ps.TotalUnits.QuoUint64(2),
		},
		StakerUnit{
			RuneAddress: GetRandomBNBAddress(),
			Height:      10,
			Units:       ps.TotalUnits.QuoUint64(2),
		},
	}

	keeper := &TestRagnarokChainKeeper{
		na:          GetRandomNodeAccount(NodeActive),
		activeVault: activeVault,
		retireVault: retireVault,
		pools:       Pools{btcPool},
		ps:          ps,
	}

	versionedTxOutStoreDummy := NewVersionedTxOutStoreDummy()
	vaultMgr := NewVaultMgr(keeper, versionedTxOutStoreDummy)
}
