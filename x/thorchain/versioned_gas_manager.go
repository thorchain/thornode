package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// VersionedGasManager
type VersionedGasManager interface {
	GetGasManager(ctx sdk.Context, version semver.Version) (GasManager, error)
}

// VersionedGasMgr implements the VersionedGasManager interface
// it provide methods to get a valid GasManager implementation by the given version
type VersionedGasMgr struct {
	gasManagerV1 GasManager
}

// NewVersionedGasMgr create a new instance of VersionedGasMgr
func NewVersionedGasMgr() *VersionedGasMgr {
	return &VersionedGasMgr{}
}

// GetGasManager return an instance that implements GasManager interface
// when there is no version can match the given semver , it will return nil
func (m *VersionedGasMgr) GetGasManager(ctx sdk.Context, version semver.Version) (GasManager, error) {
	if version.GTE(semver.MustParse("0.1.0")) {
		if m.gasManagerV1 == nil {
			m.gasManagerV1 = NewGasMgr()
		}
		return m.gasManagerV1, nil
	}
	return nil, errInvalidVersion
}
