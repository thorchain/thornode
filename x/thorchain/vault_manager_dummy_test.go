package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type VaultMgrDummy struct {
	VaultManager
}

func NewVaultMgrDummy() *VaultMgrDummy {
	return &VaultMgrDummy{}
}

func (vm *VaultMgrDummy) TriggerKeygen(_ sdk.Context, _ NodeAccounts) error { return kaboom }
