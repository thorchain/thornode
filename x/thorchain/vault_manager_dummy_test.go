package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

// VersionedVaultMgrDummy used for test purpose
type VersionedVaultMgrDummy struct {
	versionedTxOutStore VersionedTxOutStore
	vaultMgrDummy       *VaultMgrDummy
}

func NewVersionedVaultMgrDummy(versionedTxOutStore VersionedTxOutStore) *VersionedVaultMgrDummy {
	return &VersionedVaultMgrDummy{
		versionedTxOutStore: versionedTxOutStore,
	}
}

func (v *VersionedVaultMgrDummy) GetVaultManager(ctx sdk.Context, keeper Keeper, version semver.Version) (VaultManager, error) {
	if v.vaultMgrDummy == nil {
		v.vaultMgrDummy = NewVaultMgrDummy()
	}
	return v.vaultMgrDummy, nil
}

type VaultMgrDummy struct {
	nas   NodeAccounts
	vault Vault
}

func NewVaultMgrDummy() *VaultMgrDummy {
	return &VaultMgrDummy{}
}

func (vm *VaultMgrDummy) EndBlock(ctx sdk.Context, version semver.Version, constAccessor constants.ConstantValues) error {
	return nil
}

func (vm *VaultMgrDummy) TriggerKeygen(_ sdk.Context, nas NodeAccounts) error {
	vm.nas = nas
	return nil
}

func (vm *VaultMgrDummy) RotateVault(ctx sdk.Context, vault Vault) error {
	vm.vault = vault
	return nil
}
