package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// VersionedEventManager provide the ability to get an event manager based on version
type VersionedEventManager interface {
	GetEventManager(ctx sdk.Context, version semver.Version) (EventManager, error)
}

// VersionedEventMgr implement VersionedEventManager interface
type VersionedEventMgr struct {
	eventManagerV1 EventManager
}

// NewVersionedEventMgr create a new versioned event manager
func NewVersionedEventMgr() *VersionedEventMgr {
	return &VersionedEventMgr{}
}

func (m *VersionedEventMgr) GetEventManager(ctx sdk.Context, version semver.Version) (EventManager, error) {
	if version.GTE(semver.MustParse("0.1.0")) {
		if m.eventManagerV1 == nil {
			m.eventManagerV1 = NewEventMgr()
		}
		return m.eventManagerV1, nil
	}
	return nil, errInvalidVersion
}
