package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

type SlashingModule interface {
	LackObserving(_ sdk.Context) error
	LackSigning(_ sdk.Context) error
}

type Slasher struct {
	keeper     Keeper
	txOutStore TxOutStore
}

func NewSlasher(keeper Keeper, txOutStore TxOutStore) Slasher {
	return Slasher{
		keeper:     keeper,
		txOutStore: txOutStore,
	}
}

// Slash node accounts that didn't observe a single inbound txn
func (s *Slasher) LackObserving(ctx sdk.Context, constAccessor constants.ConstantValues) error {
	accs, err := s.keeper.GetObservingAddresses(ctx)
	if err != nil {
		ctx.Logger().Error("fail to get observing addresses", "error", err)
		return err
	}

	if len(accs) == 0 {
		// nobody observed anything, THORNode must of had no input txs within this
		// block
		return nil
	}

	nodes, err := s.keeper.ListActiveNodeAccounts(ctx)
	if err != nil {
		ctx.Logger().Error("Unable to get list of active accounts", "error", err)
		return err
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
			lackOfObservationPenalty := constAccessor.GetInt64Value(constants.LackOfObservationPenalty)
			na.SlashPoints += lackOfObservationPenalty
			if err := s.keeper.SetNodeAccount(ctx, na); nil != err {
				ctx.Logger().Error(fmt.Sprintf("fail to save node account(%s)", na), "error", err)
				return err
			}
		}
	}

	// clear our list of observing addresses
	s.keeper.ClearObservingAddresses(ctx)

	return nil
}

func (s *Slasher) LackSigning(ctx sdk.Context, constAccessor constants.ConstantValues) error {
	pendingEvents, err := s.keeper.GetAllPendingEvents(ctx)
	if err != nil {
		ctx.Logger().Error("Unable to get all pending events", "error", err)
		return err
	}
	signingTransPeriod := constAccessor.GetInt64Value(constants.SigningTransactionPeriod)
	for _, evt := range pendingEvents {
		// NOTE: not checking the event type because all non-swap/unstake/etc
		// are completed immediately.
		if evt.Height+signingTransPeriod < ctx.BlockHeight() {
			txs, err := s.keeper.GetTxOut(ctx, uint64(evt.Height))
			if err != nil {
				ctx.Logger().Error("Unable to get tx out list", "error", err)
				continue
			}

			for i, tx := range txs.TxArray {
				if tx.InHash.Equals(evt.InTx.ID) && tx.OutHash.IsEmpty() {
					// Slash our node account for not sending funds
					na, err := s.keeper.GetNodeAccountByPubKey(ctx, tx.VaultPubKey)
					if err != nil {
						ctx.Logger().Error("Unable to get node account", "error", err)
						continue
					}
					na.SlashPoints += signingTransPeriod * 2
					if err := s.keeper.SetNodeAccount(ctx, na); nil != err {
						ctx.Logger().Error("fail to save node account", "error", err)
					}

					active, err := s.keeper.GetAsgardVaultsByStatus(ctx, ActiveVault)
					if err != nil {
						ctx.Logger().Error("fail to get active vaults", "error", err)
						return err
					}

					vault := active.SelectByMinCoin(tx.Coin.Asset)
					if vault.IsEmpty() {
						return fmt.Errorf("unable to determine asgard vault to send funds")
					}

					// Save the tx to as a new tx, select Asgard to send it this time.
					tx.VaultPubKey = vault.PubKey
					// TODO: this creates a second tx out for this inTx, which
					// means the event will never be completed because only one
					// of the two out tx will occur.
					_, err = s.txOutStore.TryAddTxOutItem(ctx, tx)
					if err != nil {
						return fmt.Errorf("fail to add outbound tx: %w", err)
					}
				}
			}

			if err := s.keeper.SetTxOut(ctx, txs); nil != err {
				ctx.Logger().Error("fail to save tx out", "error", err)
				return err
			}
		}
	}
	return nil
}
