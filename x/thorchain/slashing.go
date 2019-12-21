package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type SlashingModule interface {
	LackObserving(_ sdk.Context) error
	LackSigning(_ sdk.Context) error
}

type Slasher struct {
	keeper      Keeper
	txOutStore  TxOutStore
	poolAddrMgr PoolAddressManager
}

func NewSlasher(keeper Keeper, txOutStore TxOutStore, poolAddrMgr PoolAddressManager) Slasher {
	return Slasher{
		keeper:      keeper,
		txOutStore:  txOutStore,
		poolAddrMgr: poolAddrMgr,
	}
}

// Slash node accounts that didn't observe a single inbound txn
func (s *Slasher) LackObserving(ctx sdk.Context, constAccessor constants.ConstantValues) error {
	accs, err := s.keeper.GetObservingAddresses(ctx)
	if err != nil {
		ctx.Logger().Error("fail to get observing addresses", err)
		return err
	}

	if len(accs) == 0 {
		// nobody observed anything, THORNode must of had no input txs within this
		// block
		return nil
	}

	nodes, err := s.keeper.ListActiveNodeAccounts(ctx)
	if err != nil {
		ctx.Logger().Error("Unable to get list of active accounts", err)
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
				ctx.Logger().Error(fmt.Sprintf("fail to save node account(%s)", na), err)
				return err
			}
		}
	}

	// clear our list of observing addresses
	s.keeper.ClearObservingAddresses(ctx)

	return nil
}

func (s *Slasher) LackSigning(ctx sdk.Context, constAccessor constants.ConstantValues) error {
	pendingEvnets, err := s.keeper.GetAllPendingEvents(ctx)
	if err != nil {
		ctx.Logger().Error("Unable to get list of active accounts", err)
		return err
	}
	signingTransPeriod := constAccessor.GetInt64Value(constants.SigningTransactionPeriod)
	for _, evt := range pendingEvnets {
		// NOTE: not checking the event type because all non-swap/unstake/etc
		// are completed immediately.
		if evt.Height+signingTransPeriod < ctx.BlockHeight() {
			txs, err := s.keeper.GetTxOut(ctx, uint64(evt.Height))
			if err != nil {
				ctx.Logger().Error("Unable to get tx out list", err)
				continue
			}

			for i, tx := range txs.TxArray {
				if tx.InHash.Equals(evt.InTx.ID) && tx.OutHash.IsEmpty() {
					// Slash our node account for not sending funds
					txs.TxArray[i].OutHash = common.BlankTxID
					na, err := s.keeper.GetNodeAccountByPubKey(ctx, tx.VaultPubKey)
					if err != nil {
						ctx.Logger().Error("Unable to get node account", err)
						continue
					}
					na.SlashPoints += signingTransPeriod * 2
					if err := s.keeper.SetNodeAccount(ctx, na); nil != err {
						ctx.Logger().Error("fail to save node account")
					}

					// Save the tx to as a new tx, select Asgard to send it this time.
					tx.VaultPubKey = s.txOutStore.GetAsgardPoolPubKey(tx.Chain).PubKey
					// TODO: this creates a second tx out for this inTx, which
					// means the event will never be completed because only one
					// of the two out tx will occur.
					s.txOutStore.AddTxOutItem(ctx, tx)
				}
			}

			if err := s.keeper.SetTxOut(ctx, txs); nil != err {
				ctx.Logger().Error("fail to save tx out", err)
				return err
			}
		}
	}
	return nil
}
