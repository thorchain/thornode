package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/constants"
)

type VaultMgrDummy struct {
	nas   NodeAccounts
	vault Vault
}

func NewVaultMgrDummy() *VaultMgrDummy {
	return &VaultMgrDummy{}
}

func (vm *VaultMgrDummy) EndBlock(ctx sdk.Context, constAccessor constants.ConstantValues) error {
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
