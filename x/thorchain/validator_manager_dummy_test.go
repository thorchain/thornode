package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// ValidatorDummyMgr is to manage a list of validators , and rotate them
type ValidatorDummyMgr struct {
	meta           *ValidatorMeta
	rotationPolicy ValidatorRotationPolicy
}

// NewValidatorDummyMgr create a new instance of ValidatorDummyMgr
func NewValidatorDummyMgr() *ValidatorDummyMgr {
	return &ValidatorDummyMgr{}
}

func (vm *ValidatorDummyMgr) Meta() *ValidatorMeta {
	return vm.meta
}

func (vm *ValidatorDummyMgr) RotationPolicy() ValidatorRotationPolicy {
	return vm.rotationPolicy
}

func (vm *ValidatorDummyMgr) BeginBlock(_ sdk.Context) {}
func (vm *ValidatorDummyMgr) EndBlock(_ sdk.Context, _ TxOutStore) []abci.ValidatorUpdate {
	return nil
}
func (vm *ValidatorDummyMgr) RequestYggReturn(_ sdk.Context, _ NodeAccount, _ PoolAddressManager, _ TxOutStore) error {
	return kaboom
}
