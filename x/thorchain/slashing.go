package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// Slash node accounts that didn't observe a single inbound txn
func slashForObservingAddresses(ctx sdk.Context, consts constants.Constants, keeper Keeper) {
	accs, err := keeper.GetObservingAddresses(ctx)
	if err != nil {
		ctx.Logger().Error("fail to get observing addresses", err)
		return
	}

	if len(accs) == 0 {
		// nobody observed anything, THORNode must of had no input txs within this
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
			na.SlashPoints += consts.LackOfObservationPenalty
			if err := keeper.SetNodeAccount(ctx, na); nil != err {
				ctx.Logger().Error(fmt.Sprintf("fail to save node account(%s)", na), err)
			}
		}
	}

	// clear our list of observing addresses
	keeper.ClearObservingAddresses(ctx)

	return
}

func slashForNotSigning(ctx sdk.Context, consts constants.Constants, keeper Keeper, txOutStore TxOutStore) {
	incomplete, err := keeper.GetIncompleteEvents(ctx)
	if err != nil {
		ctx.Logger().Error("Unable to get list of active accounts", err)
		return
	}

	for _, evt := range incomplete {
		// NOTE: not checking the event type because all non-swap/unstake/etc
		// are completed immediately.
		if evt.Height+consts.SigningTransactionPeriod < ctx.BlockHeight() {
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
					na.SlashPoints += consts.SigningTransactionPeriod * 2
					if err := keeper.SetNodeAccount(ctx, na); nil != err {
						ctx.Logger().Error("fail to save node account")
					}

					// Save the tx to as a new tx, select Asgard to send it this time.
					// Set the pool address to empty, it will overwrite it with the
					// current Asgard vault
					tx.PoolAddress = common.EmptyPubKey
					// TODO: this creates a second tx out for this inTx, which
					// means the event will never be completed because only one
					// of the two out tx will occur.
					txOutStore.AddTxOutItem(ctx, keeper, tx, true)
				}
			}

			if err := keeper.SetTxOut(ctx, txs); nil != err {
				ctx.Logger().Error("fail to save tx out", err)
				return
			}
		}
	}
}
