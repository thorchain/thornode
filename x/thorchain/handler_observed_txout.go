package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

type ObservedTxOutHandler struct {
	keeper                Keeper
	versionedTxOutStore   VersionedTxOutStore
	validatorMgr          VersionedValidatorManager
	versionedVaultManager VersionedVaultManager
}

func NewObservedTxOutHandler(keeper Keeper, txOutStore VersionedTxOutStore, validatorMgr VersionedValidatorManager, versionedVaultManager VersionedVaultManager) ObservedTxOutHandler {
	return ObservedTxOutHandler{
		keeper:                keeper,
		versionedTxOutStore:   txOutStore,
		validatorMgr:          validatorMgr,
		versionedVaultManager: versionedVaultManager,
	}
}

func (h ObservedTxOutHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, _ constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgObservedTxOut)
	if !ok {
		return errInvalidMessage.Result()
	}
	if err := h.validate(ctx, msg, version); err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	return h.handle(ctx, msg, version)
}

func (h ObservedTxOutHandler) validate(ctx sdk.Context, msg MsgObservedTxOut, version semver.Version) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	} else {
		ctx.Logger().Error(errInvalidVersion.Error())
		return errInvalidVersion
	}
}

func (h ObservedTxOutHandler) validateV1(ctx sdk.Context, msg MsgObservedTxOut) error {
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error(err.Error())
		return err
	}

	if !isSignedByActiveObserver(ctx, h.keeper, msg.GetSigners()) {
		ctx.Logger().Error(notAuthorized.Error())
		return notAuthorized
	}

	return nil
}

func (h ObservedTxOutHandler) handle(ctx sdk.Context, msg MsgObservedTxOut, version semver.Version) sdk.Result {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, version, msg)
	} else {
		ctx.Logger().Error(errInvalidVersion.Error())
		return errBadVersion.Result()
	}
}

func (h ObservedTxOutHandler) preflight(ctx sdk.Context, voter ObservedTxVoter, nas NodeAccounts, tx ObservedTx, signer sdk.AccAddress) (ObservedTxVoter, bool) {
	voter.Add(tx, signer)
	ok := false
	if voter.HasConsensus(nas) && !voter.ProcessedOut {
		ok = true
		voter.Height = ctx.BlockHeight()
		voter.ProcessedOut = true
	}
	h.keeper.SetObservedTxVoter(ctx, voter)

	// Check to see if we have enough identical observations to process the transaction
	return voter, ok
}

// Handle a message to observe outbound tx
func (h ObservedTxOutHandler) handleV1(ctx sdk.Context, version semver.Version, msg MsgObservedTxOut) sdk.Result {
	activeNodeAccounts, err := h.keeper.ListActiveNodeAccounts(ctx)
	if err != nil {
		err = wrapError(ctx, err, "fail to get list of active node accounts")
		return sdk.ErrInternal(err.Error()).Result()
	}

	handler := NewHandler(h.keeper, h.versionedTxOutStore, h.validatorMgr, h.versionedVaultManager)

	for _, tx := range msg.Txs {

		// check we are sending from a valid vault
		if !h.keeper.VaultExists(ctx, tx.ObservedPubKey) {
			ctx.Logger().Info("Not valid Observed Pubkey", tx.ObservedPubKey)
			continue
		}

		voter, err := h.keeper.GetObservedTxVoter(ctx, tx.Tx.ID)
		if err != nil {
			return sdk.ErrInternal(err.Error()).Result()
		}

		// check whether the tx has consensus
		voter, ok := h.preflight(ctx, voter, activeNodeAccounts, tx, msg.Signer)
		if !ok {
			if voter.Height == ctx.BlockHeight() {
				// we've already process the transaction, but we should still
				// update the observing addresses
				if err := h.keeper.AddObservingAddresses(ctx, msg.GetSigners()); err != nil {
					ctx.Logger().Error("fail to add observing address", "error", err)
				}
			}
			continue
		}
		ctx.Logger().Info("handleMsgObservedTxOut request", "Tx:", msg.Txs[0].String())

		// if memo isn't valid or its an inbound memo, and its funds moving
		// from a yggdrasil vault, slash the node
		memo, _ := ParseMemo(tx.Tx.Memo)
		if memo.IsEmpty() || memo.IsInbound() {
			vault, err := h.keeper.GetVault(ctx, tx.ObservedPubKey)
			if err != nil {
				ctx.Logger().Error("fail to get vault", "error", err)
				continue
			}
			if vault.IsYggdrasil() {
				slash, err := NewSlasher(h.keeper, version)
				if err != nil {
					ctx.Logger().Error("fail to create slasher:%w", err)
					continue
				}
				// a yggdrasil vault has apparently stolen funds, slash them
				for _, c := range append(tx.Tx.Coins, tx.Tx.Gas.ToCoins()...) {
					if err := slash.SlashNodeAccount(ctx, tx.ObservedPubKey, c.Asset, c.Amount); err != nil {
						ctx.Logger().Error("fail to slash account for sending extra fund", "error", err)
					}
				}
				vault.SubFunds(tx.Tx.Coins)
				vault.SubFunds(tx.Tx.Gas.ToCoins()) // we don't subsidize the gas when it's theft
				if err := h.keeper.SetVault(ctx, vault); err != nil {
					ctx.Logger().Error("fail to save vault", "error", err)
				}
				continue
			}
		}

		txOut := voter.GetTx(activeNodeAccounts) // get consensus tx, in case our for loop is incorrect
		m, err := processOneTxIn(ctx, h.keeper, txOut, msg.Signer)
		if err != nil || tx.Tx.Chain.IsEmpty() {
			ctx.Logger().Error("fail to process txOut",
				"error", err,
				"tx", tx.Tx.String())
			continue
		}

		// Apply Gas fees
		if err := AddGasFees(ctx, h.keeper, tx); err != nil {
			return sdk.ErrInternal(fmt.Errorf("fail to add gas fee: %w", err).Error()).Result()
		}

		// If sending from one of our vaults, decrement coins
		vault, err := h.keeper.GetVault(ctx, tx.ObservedPubKey)
		if err != nil {
			ctx.Logger().Error("fail to get vault", "error", err)
			continue
		}
		vault.SubFunds(tx.Tx.Coins)
		vault.OutboundTxCount += 1
		if vault.IsAsgard() && memo.IsType(txMigrate) {
			// only remove the block height that had been specified in the memo
			vault.RemovePendingTxBlockHeights(memo.GetBlockHeight())
		}
		if err := h.keeper.SetVault(ctx, vault); err != nil {
			ctx.Logger().Error("fail to save vault", "error", err)
			return sdk.ErrInternal("fail to save vault").Result()
		}

		// add addresses to observing addresses. This is used to detect
		// active/inactive observing node accounts
		if err := h.keeper.AddObservingAddresses(ctx, txOut.Signers); err != nil {
			return sdk.ErrInternal(err.Error()).Result()
		}

		result := handler(ctx, m)
		if !result.IsOK() {
			ctx.Logger().Error("Handler failed:", "error", result.Log)
		}
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
