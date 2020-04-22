package thorchain

import (
	"encoding/json"
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evi "github.com/cosmos/cosmos-sdk/x/evidence"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// Slasher implements SlashingModule interface provide the necessary functionality to slash node accounts
type Slasher struct {
	keeper  Keeper
	version semver.Version
}

// NewSlasher create a new instance of Slasher
func NewSlasher(keeper Keeper, version semver.Version) (*Slasher, error) {
	if version.GTE(semver.MustParse("0.1.0")) {
		return &Slasher{
			keeper:  keeper,
			version: version,
		}, nil
	}
	return nil, errBadVersion
}

func (s *Slasher) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	for _, tmEvidence := range req.ByzantineValidators {
		switch tmEvidence.Type {
		case tmtypes.ABCIEvidenceTypeDuplicateVote:
			evidence := evi.ConvertDuplicateVoteEvidence(tmEvidence)
			s.HandleDoubleSign(ctx, evidence.(Equivocation))
		default:
			ctx.Logger().Error(fmt.Sprintf("ignored unknown evidence type: %s", tmEvidence.Type))
		}
	}
}

func (s *Slasher) HandleDoubleSign(ctx sdk.Context, evidence types.Equivocation) error {
}

// LackObserving Slash node accounts that didn't observe a single inbound txn
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
			if err := s.keeper.SetNodeAccount(ctx, na); err != nil {
				ctx.Logger().Error(fmt.Sprintf("fail to save node account(%s)", na), "error", err)
				return err
			}
		}
	}

	return nil
}

// LackSigning slash account that fail to sign tx
func (s *Slasher) LackSigning(ctx sdk.Context, constAccessor constants.ConstantValues, txOutStore TxOutStore) error {
	pendingEvents, err := s.keeper.GetAllPendingEvents(ctx)
	if err != nil {
		ctx.Logger().Error("Unable to get all pending events", "error", err)
		return err
	}
	signingTransPeriod := constAccessor.GetInt64Value(constants.SigningTransactionPeriod)
	for _, evt := range pendingEvents {
		// NOTE: not checking the event type because all non-swap/unstake/etc
		// are completed immediately.
		if ctx.BlockHeight() == evt.Height+signingTransPeriod {
			txs, err := s.keeper.GetTxOut(ctx, evt.Height)
			if err != nil {
				ctx.Logger().Error("Unable to get tx out list", "error", err)
				continue
			}

			for _, tx := range txs.TxArray {
				if tx.InHash.Equals(evt.InTx.ID) && tx.OutHash.IsEmpty() {
					// Slash our node account for not sending funds
					vault, err := s.keeper.GetVault(ctx, tx.VaultPubKey)
					if err != nil {
						ctx.Logger().Error("Unable to get vault", "error", err)
						continue
					}
					// slash if its a yggdrasil vault
					if vault.IsYggdrasil() {
						na, err := s.keeper.GetNodeAccountByPubKey(ctx, tx.VaultPubKey)
						if err != nil {
							ctx.Logger().Error("Unable to get node account", "error", err)
							continue
						}
						na.SlashPoints += signingTransPeriod * 2
						if err := s.keeper.SetNodeAccount(ctx, na); err != nil {
							ctx.Logger().Error("fail to save node account", "error", err)
						}
					}

					active, err := s.keeper.GetAsgardVaultsByStatus(ctx, ActiveVault)
					if err != nil {
						ctx.Logger().Error("fail to get active vaults", "error", err)
						return err
					}

					vault = active.SelectByMinCoin(tx.Coin.Asset)
					if vault.IsEmpty() {
						return fmt.Errorf("unable to determine asgard vault to send funds")
					}

					// update original tx action in observed tx
					voter, err := s.keeper.GetObservedTxVoter(ctx, tx.InHash)
					if err != nil {
						return fmt.Errorf("fail to get observed tx voter: %w", err)
					}
					for i, action := range voter.Actions {
						if action.Equals(*tx) {
							voter.Actions[i].VaultPubKey = vault.PubKey
						}
					}
					s.keeper.SetObservedTxVoter(ctx, voter)

					// Save the tx to as a new tx, select Asgard to send it this time.
					tx.VaultPubKey = vault.PubKey
					err = txOutStore.UnSafeAddTxOutItem(ctx, tx)
					if err != nil {
						return fmt.Errorf("fail to add outbound tx: %w", err)
					}
				}
			}

			if err := s.keeper.SetTxOut(ctx, txs); err != nil {
				ctx.Logger().Error("fail to save tx out", "error", err)
				return err
			}
		}
	}
	return nil
}

// slashNodeAccount thorchain keep monitoring the outbound tx from asgard pool
// and yggdrasil pool, usually the txout is triggered by thorchain itself by
// adding an item into the txout array, refer to TxOutItem for the detail, the
// TxOutItem contains a specific coin and amount.  if somehow thorchain
// discover signer send out fund more than the amount specified in TxOutItem,
// it will slash the node account who does that by taking 1.5 * extra fund from
// node account's bond and subsidise the pool that actually lost it.
func (s *Slasher) SlashNodeAccount(ctx sdk.Context, observedPubKey common.PubKey, asset common.Asset, slashAmount sdk.Uint) error {
	if slashAmount.IsZero() {
		return nil
	}
	nodeAccount, err := s.keeper.GetNodeAccountByPubKey(ctx, observedPubKey)
	if err != nil {
		return fmt.Errorf("fail to get node account with pubkey(%s), %w", observedPubKey, err)
	}

	if nodeAccount.Status == NodeUnknown {
		return nil
	}

	if asset.IsRune() {
		// If rune, we take 1.5x the amount, and take it from their bond. We
		// put 1/3rd of it into the reserve, and 2/3rds into the pools (but
		// keeping the rune pool balances unchanged)
		amountToReserve := slashAmount.QuoUint64(2)
		// if the diff asset is RUNE , just took 1.5 * diff from their bond
		slashAmount = slashAmount.MulUint64(3).QuoUint64(2)
		nodeAccount.Bond = common.SafeSub(nodeAccount.Bond, slashAmount)
		vaultData, err := s.keeper.GetVaultData(ctx)
		if err != nil {
			return fmt.Errorf("fail to get vault data: %w", err)
		}
		vaultData.TotalReserve = vaultData.TotalReserve.Add(amountToReserve)
		if err := s.keeper.SetVaultData(ctx, vaultData); err != nil {
			return fmt.Errorf("fail to save vault data: %w", err)
		}
		return s.keeper.SetNodeAccount(ctx, nodeAccount)
	}
	pool, err := s.keeper.GetPool(ctx, asset)
	if err != nil {
		return fmt.Errorf("fail to get %s pool : %w", asset, err)
	}
	// thorchain doesn't even have a pool for the asset, or the pool had been
	// suspended, then who cares
	if pool.Empty() || pool.Status == PoolSuspended {
		return nil
	}
	runeValue := pool.AssetValueInRune(slashAmount).MulUint64(3).QuoUint64(2)
	pool.BalanceAsset = common.SafeSub(pool.BalanceAsset, slashAmount)
	pool.BalanceRune = pool.BalanceRune.Add(runeValue)
	nodeAccount.Bond = common.SafeSub(nodeAccount.Bond, runeValue)
	if err := s.keeper.SetPool(ctx, pool); err != nil {
		return fmt.Errorf("fail to save %s pool: %w", asset, err)
	}

	poolSlashAmt := []PoolAmt{
		{
			Asset:  pool.Asset,
			Amount: 0 - int64(slashAmount.Uint64()),
		},
		{
			Asset:  common.RuneAsset(),
			Amount: int64(runeValue.Uint64()),
		},
	}
	eventSlash := NewEventSlash(pool.Asset, poolSlashAmt)
	slashBuf, err := json.Marshal(eventSlash)
	if err != nil {
		return fmt.Errorf("fail to marshal slash event to buf: %w", err)
	}
	event := NewEvent(
		eventSlash.Type(),
		ctx.BlockHeight(),
		common.Tx{ID: common.BlankTxID},
		slashBuf,
		EventSuccess,
	)
	if err := s.keeper.UpsertEvent(ctx, event); err != nil {
		return fmt.Errorf("fail to save event: %w", err)
	}

	return s.keeper.SetNodeAccount(ctx, nodeAccount)
}
