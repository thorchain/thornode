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

func (vm *ValidatorDummyMgr) BeginBlock(ctx sdk.Context) {}
func (vm *ValidatorDummyMgr) EndBlock(ctx sdk.Context, store *TxOutStore) []abci.ValidatorUpdate {
	return nil
}
func (vm *ValidatorDummyMgr) processValidatorLeave(ctx sdk.Context, store *TxOutStore) (bool, error) {
	return false, kaboom
}
func (vm *ValidatorDummyMgr) rotateValidatorNodes(ctx sdk.Context, store *TxOutStore) (bool, error) {
	return false, kaboom
}
func (vm *ValidatorDummyMgr) requestYggReturn(ctx sdk.Context, node NodeAccount, poolAddrMgr PoolAddressManager, txOut *TxOutStore) error {
	return kaboom
}
func (vm *ValidatorDummyMgr) prepareToNodesToLeave(ctx sdk.Context, txOut *TxOutStore) error {
	return kaboom
}
func (vm *ValidatorDummyMgr) ragnarokProtocolStep1(ctx sdk.Context, activeNodes NodeAccounts, txOut *TxOutStore) error {
	return kaboom
}
func (vm *ValidatorDummyMgr) recallYggFunds(ctx sdk.Context, activeNodes NodeAccounts, txOut *TxOutStore) error {
	return kaboom
}
func (vm *ValidatorDummyMgr) prepareAddNode(ctx sdk.Context, height int64) error {
	return kaboom
}
func (vm *ValidatorDummyMgr) setupValidatorNodes(ctx sdk.Context, height int64) error {
	return kaboom
}
