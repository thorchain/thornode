package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"gitlab.com/thorchain/thornode/constants"
)

const (
	genesisBlockHeight = 1
)

// VersionedValidatorManager is an interface define the contract of validator manager that has version support
type VersionedValidatorManager interface {
	BeginBlock(ctx sdk.Context, version semver.Version, constAccessor constants.ConstantValues) error
	EndBlock(ctx sdk.Context, version semver.Version, constAccessor constants.ConstantValues) []abci.ValidatorUpdate
	RequestYggReturn(ctx sdk.Context, version semver.Version, node NodeAccount) error
}

// VersionedValidatorMgr
type VersionedValidatorMgr struct {
	keeper                Keeper
	v1ValidatorMgr        *validatorMgrV1
	versionedTxOutStore   VersionedTxOutStore
	versionedVaultManager VersionedVaultManager
	versionedEventManager VersionedEventManager
}

// NewVersionedValidatorMgr create a new versioned validator mgr , which require to pass in a version
func NewVersionedValidatorMgr(k Keeper, versionedTxOutStore VersionedTxOutStore, versionedVaultManager VersionedVaultManager, versionedEventManager VersionedEventManager) *VersionedValidatorMgr {
	return &VersionedValidatorMgr{
		keeper:                k,
		versionedTxOutStore:   versionedTxOutStore,
		versionedVaultManager: versionedVaultManager,
		versionedEventManager: versionedEventManager,
	}
}

// BeginBlock start to process a new block
func (vm *VersionedValidatorMgr) BeginBlock(ctx sdk.Context, version semver.Version, constAccessor constants.ConstantValues) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		if vm.v1ValidatorMgr == nil {
			vm.v1ValidatorMgr = newValidatorMgrV1(vm.keeper, vm.versionedTxOutStore, vm.versionedVaultManager, vm.versionedEventManager)
		}
		return vm.v1ValidatorMgr.BeginBlock(ctx, constAccessor)
	}
	return errBadVersion
}

// EndBlock when a block need to commit
func (vm *VersionedValidatorMgr) EndBlock(ctx sdk.Context, version semver.Version, constAccessor constants.ConstantValues) []abci.ValidatorUpdate {
	if version.GTE(semver.MustParse("0.1.0")) {
		if vm.v1ValidatorMgr == nil {
			vm.v1ValidatorMgr = newValidatorMgrV1(vm.keeper, vm.versionedTxOutStore, vm.versionedVaultManager, vm.versionedEventManager)
		}
		return vm.v1ValidatorMgr.EndBlock(ctx, constAccessor)
	}
	ctx.Logger().Error(fmt.Sprintf("unsupported version (%s) in validator manager", version))
	return nil
}

// RequestYggReturn request yggdrasil pool to return fund
func (vm *VersionedValidatorMgr) RequestYggReturn(ctx sdk.Context, version semver.Version, node NodeAccount) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		if vm.v1ValidatorMgr == nil {
			vm.v1ValidatorMgr = newValidatorMgrV1(vm.keeper, vm.versionedTxOutStore, vm.versionedVaultManager, vm.versionedEventManager)
		}
		return vm.v1ValidatorMgr.RequestYggReturn(ctx, node)
	}
	return errBadVersion
}
