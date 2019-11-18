package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: move to constants.go
const (
	observingPenalty int64 = 2 // add two slash point for each offense
)

// Slash node accounts that didn't observe a single inbound txn
func slashForObservingAddresses(ctx sdk.Context, keeper Keeper) {
	accs := keeper.GetObservingAddresses(ctx)

	if len(accs) == 0 {
		// nobody observed anything, we must of had no input txs within this
		// block
		return
	}

	nodes, err := keeper.ListActiveNodeAccounts(ctx)
	if err != nil {
		ctx.Logger().Error("Unable to get list of active accounts", err)
		return
	}

	for _, na := range nodes {
		found := false
		for _, addr := range accs {
			if na.NodeAddress.Equals(addr) {
				found = true
				break
			}
		}

		// this na is not found, therefore it should be slashed
		if !found {
			na.SlashPoints += observingPenalty
			keeper.SetNodeAccount(ctx, na)
		}
	}

	// clear our list of observing addresses
	keeper.ClearObservingAddresses(ctx)

	return
}
