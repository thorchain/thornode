package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// VersionedSwapQueue
type VersionedSwapQueue interface {
	GetSwapQueue(ctx sdk.Context, keeper Keeper, version semver.Version) (SwapQueue, error)
}

// SwapQueue interface define the contract of Swap Queue
type SwapQueue interface {
}

// VersionedSwapQ is an implementation of versioned Vault Manager
type VersionedSwapQ struct {
	queue               SwapQueue
	versionedTxOutStore VersionedTxOutStore
}

// NewVersionedSwapQ create a new instance of VersionedSwapQ
func NewVersionedSwapQ(versionedTxOutStore VersionedTxOutStore) VersionedSwapQueue {
	return &VersionedSwapQ{
		versionedTxOutStore: versionedTxOutStore,
	}
}

// GetSwapQueue retrieve a SwapQueue that is compatible with the given version
func (v *VersionedSwapQ) GetSwapQueue(ctx sdk.Context, keeper Keeper, version semver.Version) (SwapQueue, error) {
	if version.GTE(semver.MustParse("0.1.0")) {
		if v.queue == nil {
			v.queue = NewSwapQv1(keeper, v.versionedTxOutStore)
		}
		return v.queue, nil
	}
	return nil, errInvalidVersion
}
