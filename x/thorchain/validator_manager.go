package thorchain

import (
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
	v1ValidatorMgr *validatorMgrV1
}

// NewVersionedValidatorMgr create a new versioned validator mgr , which require to pass in a version
func NewVersionedValidatorMgr(k Keeper, txOut TxOutStore, vaultMgr VaultManager) *VersionedValidatorMgr {
	return &VersionedValidatorMgr{
		v1ValidatorMgr: newValidatorMgrV1(k, txOut, vaultMgr),
	}
}

// BeginBlock start to process a new block
func (vm *VersionedValidatorMgr) BeginBlock(ctx sdk.Context, version semver.Version, constAccessor constants.ConstantValues) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return vm.v1ValidatorMgr.BeginBlock(ctx, constAccessor)
	}
	return errBadVersion
}
