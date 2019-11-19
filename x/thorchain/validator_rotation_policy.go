package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidatorRotationPolicy store all the relevant parameters we need to perform validator rotation
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
func GetValidatorRotationPolicy(ctx sdk.Context, k Keeper) ValidatorRotationPolicy {
	return ValidatorRotationPolicy{
		RotateInNumBeforeFull:      k.GetAdminConfigValidatorRotateInNumBeforeFull(ctx, nil),
		RotateOutNumBeforeFull:     k.GetAdminConfigValidatorRotateOutNumBeforeFull(ctx, nil),
		RotateNumAfterFull:         k.GetAdminConfigValidatorRotateNumAfterFull(ctx, nil),
		RotatePerBlockHeight:       k.GetAdminConfigRotatePerBlockHeight(ctx, nil),
		ValidatorChangeWindow:      k.GetAdminConfigValidatorsChangeWindow(ctx, nil),
		DesireValidatorSet:         k.GetAdminConfigDesireValidatorSet(ctx, nil),
		LeaveProcessPerBlockHeight: 4320,
	}
}
func (vp ValidatorRotationPolicy) IsValid() error {
	if vp.ValidatorChangeWindow > vp.RotatePerBlockHeight {
		return fmt.Errorf("validator change window :%d is larger than rotate per block height: %d", vp.ValidatorChangeWindow, vp.RotatePerBlockHeight)
	}
	if vp.RotateOutNumBeforeFull > vp.RotateInNumBeforeFull {
		return fmt.Errorf("rotate out %d is larger than rotate in %d we will never reach the desire validator set", vp.RotateOutNumBeforeFull, vp.RotateInNumBeforeFull)
	}
	return nil
}
