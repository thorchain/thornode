package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

// VersionedVaultManager
type VersionedVaultManager interface {
	GetVaultManager(ctx sdk.Context, keeper Keeper, version semver.Version) (VaultManager, error)
}

// VaultManager interface define the contract of Vault Manager
type VaultManager interface {
	TriggerKeygen(ctx sdk.Context, nas NodeAccounts) error
	RotateVault(ctx sdk.Context, vault Vault) error
	EndBlock(ctx sdk.Context, version semver.Version, constAccessor constants.ConstantValues) error
}

// VersionedVaultMgr is an implementation of versioned Vault Manager
type VersionedVaultMgr struct {
	vaultMgrV1            *VaultMgr
	versionedTxOutStore   VersionedTxOutStore
	versionedEventManager VersionedEventManager
}

// NewVersionedVaultMgr create a new instance of VersionedVaultMgr
func NewVersionedVaultMgr(versionedTxOutStore VersionedTxOutStore, versionedEventManager VersionedEventManager) *VersionedVaultMgr {
	return &VersionedVaultMgr{
		versionedTxOutStore: versionedTxOutStore,
	}
}

// GetVaultManager retrieve a VaultManager that is compatible with the given version
func (v *VersionedVaultMgr) GetVaultManager(ctx sdk.Context, keeper Keeper, version semver.Version) (VaultManager, error) {
	if version.GTE(semver.MustParse("0.1.0")) {
		if v.vaultMgrV1 == nil {
			v.vaultMgrV1 = NewVaultMgr(keeper, v.versionedTxOutStore, v.versionedEventManager)
		}
		return v.vaultMgrV1, nil
	}
	return nil, errInvalidVersion
}
