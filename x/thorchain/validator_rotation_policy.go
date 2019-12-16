package thorchain

import (
	"fmt"

	"gitlab.com/thorchain/thornode/constants"
)

// ValidatorRotationPolicy store all the relevant parameters THORNode need to perform validator rotation
type ValidatorRotationPolicy struct {
	RotateInNumBeforeFull      int64
	RotateOutNumBeforeFull     int64
	RotateNumAfterFull         int64
	RotatePerBlockHeight       int64
	ValidatorChangeWindow      int64
	DesireValidatorSet         int64
	LeaveProcessPerBlockHeight int64
}

// GetValidatorRotationPolicy from data store
func GetValidatorRotationPolicy() ValidatorRotationPolicy {
	return ValidatorRotationPolicy{
		RotateInNumBeforeFull:      constants.ValidatorRotateInNumBeforeFull,
		RotateOutNumBeforeFull:     constants.ValidatorRotateOutNumBeforeFull,
		RotateNumAfterFull:         constants.ValidatorRotateNumAfterFull,
		RotatePerBlockHeight:       constants.RotatePerBlockHeight,
		ValidatorChangeWindow:      constants.ValidatorsChangeWindow,
		DesireValidatorSet:         constants.DesireValidatorSet,
		LeaveProcessPerBlockHeight: constants.LeaveProcessPerBlockHeight,
	}
}

func (vp ValidatorRotationPolicy) IsValid() error {
	if vp.ValidatorChangeWindow > vp.RotatePerBlockHeight {
		return fmt.Errorf("validator change window :%d is larger than rotate per block height: %d", vp.ValidatorChangeWindow, vp.RotatePerBlockHeight)
	}
	if vp.RotateOutNumBeforeFull > vp.RotateInNumBeforeFull {
		return fmt.Errorf("rotate out %d is larger than rotate in %d THORNode will never reach the desire validator set", vp.RotateOutNumBeforeFull, vp.RotateInNumBeforeFull)
	}
	return nil
}
