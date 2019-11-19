package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
)

// TODO: move to constants.go
const (
	observingPenalty         int64 = 2 // add two slash point for each offense
	signingTransactionPeriod int64 = 100
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

func slashForNotSigning(ctx sdk.Context, keeper Keeper, txOutStore *TxOutStore) {
	incomplete, err := keeper.GetIncompleteEvents(ctx)
	if err != nil {
		ctx.Logger().Error("Unable to get list of active accounts", err)
		return
	}

	for _, evt := range incomplete {
		// NOTE: not checking the event type because all non-swap/unstake/etc
		// are completed immediately.
		if evt.Height+signingTransactionPeriod > ctx.BlockHeight() {
			txs, err := keeper.GetTxOut(ctx, uint64(evt.Height))
			if err != nil {
				ctx.Logger().Error("Unable to get tx out list", err)
				continue
			}

			for i, tx := range txs.TxArray {
				if tx.InHash.Equals(evt.InTx.ID) && tx.OutHash.IsEmpty() {
					// Slash our node account for not sending funds
					txs.TxArray[i].OutHash = common.BlankTxID
					na, err := keeper.GetNodeAccountByPubKey(ctx, tx.PoolAddress)
					if err != nil {
						ctx.Logger().Error("Unable to get node account", err)
						continue
					}
					na.SlashPoints += signingTransactionPeriod * 2
					keeper.SetNodeAccount(ctx, na)

					// Save the tx to as a new tx, select Asgard to send it this time.
					// Set the pool address to empty, it will overwrite it with the
					// current Asgard vault
					tx.PoolAddress = common.EmptyPubKey
					txOutStore.AddTxOutItem(ctx, keeper, tx, true, true)
				}
			}

			keeper.SetTxOut(ctx, txs)
		}
	}
}
