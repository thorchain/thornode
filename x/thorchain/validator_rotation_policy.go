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
func GetValidatorRotationPolicy(constAccessor constants.ConstantValues) ValidatorRotationPolicy {

	return ValidatorRotationPolicy{
		RotateInNumBeforeFull:      constAccessor.GetInt64Value(constants.ValidatorRotateInNumBeforeFull),
		RotateOutNumBeforeFull:     constAccessor.GetInt64Value(constants.ValidatorRotateOutNumBeforeFull),
		RotateNumAfterFull:         constAccessor.GetInt64Value(constants.ValidatorRotateNumAfterFull),
		RotatePerBlockHeight:       constAccessor.GetInt64Value(constants.RotatePerBlockHeight),
		ValidatorChangeWindow:      constAccessor.GetInt64Value(constants.ValidatorsChangeWindow),
		DesireValidatorSet:         constAccessor.GetInt64Value(constants.DesireValidatorSet),
		LeaveProcessPerBlockHeight: constAccessor.GetInt64Value(constants.LeaveProcessPerBlockHeight),
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
