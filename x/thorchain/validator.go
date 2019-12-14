package thorchain

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ValidatorManager2 interface {
}

type ValidatorMgr2 struct {
	keeper Keeper
}

func NewValidatorMgr2(keeper Keeper) *ValidatorMgr2 {
	return &ValidatorMgr2{
		keeper: keeper,
	}
}

// Iterate over active node accounts, finding the one with the most slash points
func (v *ValidatorMgr2) findBadActor(ctx sdk.Context) (NodeAccount, error) {
	na := NodeAccount{}
	nas, err := v.keeper.ListActiveNodeAccounts(ctx)
	if err != nil {
		return na, err
	}

	// TODO: return if we're at risk of loosing BTF

	// Find bad actor relative to slashpoints / age.
	// NOTE: avoiding the usage of float64, we use an alt method...
	na.SlashPoints = 1
	na.StatusSince = 9223372036854775807 // highest int64 value
	for _, n := range nas {
		if n.SlashPoints == 0 {
			continue
		}

		naVal := n.StatusSince / na.SlashPoints
		nVal := n.StatusSince / n.SlashPoints
		if nVal > (naVal) {
			na = n
		} else if nVal == naVal {
			if n.SlashPoints > na.SlashPoints {
				na = n
			}
		}
	}

	return na, nil
}

// Iterate over active node accounts, finding the one that has been active longest
func (v *ValidatorMgr2) findOldActor(ctx sdk.Context) (NodeAccount, error) {
	na := NodeAccount{}
	nas, err := v.keeper.ListActiveNodeAccounts(ctx)
	if err != nil {
		return na, err
	}

	// TODO: return if we're at risk of loosing BTF

	na.StatusSince = ctx.BlockHeight() // set the start status age to "now"
	for _, n := range nas {
		if n.StatusSince < na.StatusSince {
			na = n
		}
	}

	return na, nil
}

// Mark an old to be churned out
func (v *ValidatorMgr2) markActor(ctx sdk.Context, na NodeAccount) error {
	na.LeaveHeight = ctx.BlockHeight()
	return v.keeper.SetNodeAccount(ctx, na)
}

// Mark an old actor to be churned out
func (v *ValidatorMgr2) markOldActor(ctx sdk.Context, rate int64) error {
	if rate%ctx.BlockHeight() == 0 {
		na, err := v.findOldActor(ctx)
		if err != nil {
			return err
		}
		if err := v.markActor(ctx, na); err != nil {
			return err
		}
	}
	return nil
}

// Mark a bad actor to be churned out
func (v *ValidatorMgr2) markBadActor(ctx sdk.Context, rate int64) error {
	if rate%ctx.BlockHeight() == 0 {
		na, err := v.findBadActor(ctx)
		if err != nil {
			return err
		}
		if err := v.markActor(ctx, na); err != nil {
			return err
		}
	}
	return nil
}

// find any actor that are ready to become "ready" status
func (v *ValidatorMgr2) markReadyActors(ctx sdk.Context) error {
	standby, err := v.keeper.ListNodeAccountsByStatus(ctx, NodeStandby)
	if err != nil {
		return err
	}
	ready, err := v.keeper.ListNodeAccountsByStatus(ctx, NodeReady)
	if err != nil {
		return err
	}

	// find min version node has to be, to be "ready" status
	minVersion := v.keeper.GetMinJoinVersion(ctx)

	// check all ready and standby nodes are in "ready" state (upgrade/downgrade as needed)
	for _, na := range append(standby, ready...) {
		na.Status = NodeReady // everyone starts with the benefit of the doubt

		// TODO: check node is up to date on thorchain, binance, etc
		// must have made an observation that matched 2/3rds within the last 5 blocks

		// Check version number is still supported
		if na.Version.LT(minVersion) {
			na.Status = NodeStandby
		}

		if err := v.keeper.SetNodeAccount(ctx, na); err != nil {
			return err
		}
	}

	return nil
}

// Returns a list of nodes to include in the next pool
func (v *ValidatorMgr2) nextPoolNodeAccounts(ctx sdk.Context, targetCount int) (NodeAccounts, bool, error) {
	rotation := false // track if are making any changes to the current active node accounts

	// update list of ready actors
	if err := v.markReadyActors(ctx); err != nil {
		return nil, false, err
	}

	ready, err := v.keeper.ListNodeAccountsByStatus(ctx, NodeReady)
	if err != nil {
		return nil, false, err
	}
	// sort by bond size
	sort.Slice(ready, func(i, j int) bool {
		return ready[i].Bond.GT(ready[j].Bond)
	})

	active, err := v.keeper.ListActiveNodeAccounts(ctx)
	if err != nil {
		return nil, false, err
	}
	// sort by LeaveHeight
	sort.Slice(active, func(i, j int) bool {
		return active[i].LeaveHeight < active[j].LeaveHeight
	})

	// remove a node node account, if one is marked to leave
	if len(active) > 0 && active[0].LeaveHeight > 0 {
		rotation = true
		active = active[1:]
	}

	// add ready nodes to become active
	limit := 2 // Max limit of ready nodes to add. TODO: this should be a constant
	for i := 1; i <= targetCount-len(active); i++ {
		if len(ready) >= i {
			rotation = true
			active = append(active, ready[i-1])
		}
		if i == limit { // limit adding ready accounts
			break
		}
	}

	return active, rotation, nil
}

func (v *ValidatorMgr2) BeginBlock(ctx sdk.Context, desiredValCount int, oldRate, badRate, rotateRate int64) error {
	if err := v.markBadActor(ctx, badRate); err != nil {
		return err
	}

	if err := v.markOldActor(ctx, oldRate); err != nil {
		return err
	}

	if ctx.BlockHeight()%rotateRate == 0 {
		next, ok, err := v.nextPoolNodeAccounts(ctx, desiredValCount)
		if err != nil {
			return err
		}
		if ok {
			_ = next // TODO: trigger pool rotation...
		}
	}

	return nil
}
