package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

// VersionedSwapQueue
type VersionedSwapQueue interface {
	GetSwapQueue(ctx sdk.Context, keeper Keeper, version semver.Version) (SwapQueue, error)
}

// SwapQueue interface define the contract of Swap Queue
type SwapQueue interface {
	EndBlock(ctx sdk.Context, version semver.Version, constAccessor constants.ConstantValues) error
}

// VersionedSwapQ is an implementation of versioned Vault Manager
type VersionedSwapQ struct {
	queue                 SwapQueue
	versionedTxOutStore   VersionedTxOutStore
	versionedEventManager VersionedEventManager
}

// NewVersionedSwapQ create a new instance of VersionedSwapQ
func NewVersionedSwapQ(versionedTxOutStore VersionedTxOutStore, versionedEventManager VersionedEventManager) VersionedSwapQueue {
	return &VersionedSwapQ{
		versionedTxOutStore:   versionedTxOutStore,
		versionedEventManager: versionedEventManager,
	}
}

// GetSwapQueue retrieve a SwapQueue that is compatible with the given version
func (v *VersionedSwapQ) GetSwapQueue(ctx sdk.Context, keeper Keeper, version semver.Version) (SwapQueue, error) {
	if version.GTE(semver.MustParse("0.1.0")) {
		if v.queue == nil {
			v.queue = NewSwapQv1(keeper, v.versionedTxOutStore, v.versionedEventManager)
		}
		return v.queue, nil
	}
	return nil, errInvalidVersion
}
