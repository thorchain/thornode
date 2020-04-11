package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// VersionedObserverManager
type VersionedObserverManager interface {
	GetObserverManager(ctx sdk.Context, version semver.Version) (ObserverManager, error)
}

// VersionedObserverMgr implements the VersionedObserverManager interface
// it provide methods to get a valid ObserverManager implementation by the given version
type VersionedObserverMgr struct {
	observerManagerV1 ObserverManager
}

// NewVersionedObserverMgr create a new instance of VersionedObserverMgr
func NewVersionedObserverMgr() *VersionedObserverMgr {
	return &VersionedObserverMgr{}
}

// GetObserverManager return an instance that implements ObserverManager interface
// when there is no version can match the given semver , it will return nil
func (m *VersionedObserverMgr) GetObserverManager(ctx sdk.Context, version semver.Version) (ObserverManager, error) {
	if version.GTE(semver.MustParse("0.1.0")) {
		if m.observerManagerV1 == nil {
			m.observerManagerV1 = NewObserverMgr()
		}
		return m.observerManagerV1, nil
	}
	return nil, errInvalidVersion
}
