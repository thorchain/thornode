package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// ObservedTxInHandler to handle MsgObservedTxIn
type ObservedTxInHandler struct {
	keeper                   Keeper
	versionedTxOutStore      VersionedTxOutStore
	validatorMgr             VersionedValidatorManager
	versionedVaultManager    VersionedVaultManager
	versionedGasMgr          VersionedGasManager
	versionedObserverManager VersionedObserverManager
	versionedEventManager    VersionedEventManager
}

// NewObservedTxInHandler create a new instance of ObservedTxInHandler
func NewObservedTxInHandler(keeper Keeper,
	versionedObserverManager VersionedObserverManager,
	versionedTxOutStore VersionedTxOutStore,
	validatorMgr VersionedValidatorManager,
	versionedVaultManager VersionedVaultManager,
	versionedGasMgr VersionedGasManager,
	versionedEventManager VersionedEventManager) ObservedTxInHandler {
	return ObservedTxInHandler{
		keeper:                   keeper,
		versionedTxOutStore:      versionedTxOutStore,
		validatorMgr:             validatorMgr,
		versionedVaultManager:    versionedVaultManager,
		versionedGasMgr:          versionedGasMgr,
		versionedObserverManager: versionedObserverManager,
		versionedEventManager:    versionedEventManager,
	}
}

func (h ObservedTxInHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, _ constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgObservedTxIn)
	if !ok {
		return errInvalidMessage.Result()
	}
	isNewSigner, err := h.validate(ctx, msg, version)
	if err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	if isNewSigner {
		return sdk.Result{
			Code:      sdk.CodeOK,
			Codespace: DefaultCodespace,
		}
	}
	return h.handle(ctx, msg, version)
}

func (h ObservedTxInHandler) validate(ctx sdk.Context, msg MsgObservedTxIn, version semver.Version) (bool, error) {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	} else {
		ctx.Logger().Error(errInvalidVersion.Error())
		return false, errInvalidVersion
	}
}

func (h ObservedTxInHandler) validateV1(ctx sdk.Context, msg MsgObservedTxIn) (bool, error) {
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error(err.Error())
		return false, err
	}

	if !isSignedByActiveNodeAccounts(ctx, h.keeper, msg.GetSigners()) {
		ctx.Logger().Error(notAuthorized.Error())
		return false, notAuthorized
	}

	return false, nil
}

func (h ObservedTxInHandler) handle(ctx sdk.Context, msg MsgObservedTxIn, version semver.Version) sdk.Result {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, version, msg)
	} else {
		ctx.Logger().Error(errInvalidVersion.Error())
		return errBadVersion.Result()
	}
}

func (h ObservedTxInHandler) preflight(ctx sdk.Context, voter ObservedTxVoter, nas NodeAccounts, tx ObservedTx, signer sdk.AccAddress) (ObservedTxVoter, bool) {
	voter.Add(tx, signer)

	ok := false
	if voter.HasConsensus(nas) && !voter.ProcessedIn {
		ok = true
		voter.Height = ctx.BlockHeight()
		voter.ProcessedIn = true
		// this is the tx that has consensus
		voter.Tx = voter.GetTx(nas)
	}
	h.keeper.SetObservedTxVoter(ctx, voter)

	// Check to see if we have enough identical observations to process the transaction
	return voter, ok
}

// Handle a message to observe inbound tx
func (h ObservedTxInHandler) handleV1(ctx sdk.Context, version semver.Version, msg MsgObservedTxIn) sdk.Result {
	constAccessor := constants.GetConstantValues(version)
	activeNodeAccounts, err := h.keeper.ListActiveNodeAccounts(ctx)
	if err != nil {
		err = wrapError(ctx, err, "fail to get list of active node accounts")
		return sdk.ErrInternal(err.Error()).Result()
	}
	txOutStore, err := h.versionedTxOutStore.GetTxOutStore(ctx, h.keeper, version)
	if err != nil {
		ctx.Logger().Error("fail to get txout store", "error", err)
		return errBadVersion.Result()
	}
	obMgr, err := h.versionedObserverManager.GetObserverManager(ctx, version)
	if err != nil {
		ctx.Logger().Error("fail to get observer manager", "error", err)
		return errBadVersion.Result()
	}
	eventMgr, err := h.versionedEventManager.GetEventManager(ctx, version)
	if err != nil {
		ctx.Logger().Error("fail to get event manager", "error", err)
		return errFailGetEventManager.Result()
	}
	handler := NewInternalHandler(h.keeper, h.versionedTxOutStore, h.validatorMgr, h.versionedVaultManager, h.versionedObserverManager, h.versionedGasMgr, h.versionedEventManager)

	for _, tx := range msg.Txs {

		// check we are sending to a valid vault
		if !h.keeper.VaultExists(ctx, tx.ObservedPubKey) {
			ctx.Logger().Info("Not valid Observed Pubkey", tx.ObservedPubKey)
			continue
		}

		voter, err := h.keeper.GetObservedTxVoter(ctx, tx.Tx.ID)
		if err != nil {
			return sdk.ErrInternal(err.Error()).Result()
		}

		voter, ok := h.preflight(ctx, voter, activeNodeAccounts, tx, msg.Signer)
		if !ok {
			if voter.Height == ctx.BlockHeight() {
				// we've already process the transaction, but we should still
				// update the observing addresses
				obMgr.AppendObserver(tx.Tx.Chain, msg.GetSigners())
			}
			continue
		}

		tx.Tx.Memo = fetchMemo(ctx, constAccessor, h.keeper, tx.Tx)
		if len(tx.Tx.Memo) == 0 {
			// we didn't find our memo, it might be yggdrasil return. These are
			// tx markers without coin amounts because we allow yggdrasil to
			// figure out the coin amounts
			txYgg := tx.Tx
			txYgg.Coins = common.Coins{
				common.NewCoin(common.RuneAsset(), sdk.ZeroUint()),
			}
			tx.Tx.Memo = fetchMemo(ctx, constAccessor, h.keeper, txYgg)
		}

		ctx.Logger().Info("handleMsgObservedTxIn request", "Tx:", tx.String())

		txIn := voter.GetTx(activeNodeAccounts)
		txIn.Tx.Memo = tx.Tx.Memo
		vault, err := h.keeper.GetVault(ctx, tx.ObservedPubKey)
		if err != nil {
			ctx.Logger().Error("fail to get vault", "error", err)
			return sdk.ErrInternal(err.Error()).Result()
		}

		vault.AddFunds(tx.Tx.Coins)
		vault.InboundTxCount += 1
		memo, _ := ParseMemo(tx.Tx.Memo) // ignore err
		if vault.IsYggdrasil() && memo.IsType(TxYggdrasilFund) {
			vault.RemovePendingTxBlockHeights(memo.GetBlockHeight())
		}
		if err := h.keeper.SetVault(ctx, vault); err != nil {
			ctx.Logger().Error("fail to save vault", "error", err)
			return sdk.ErrInternal(err.Error()).Result()
		}

		if !vault.IsAsgard() {
			ctx.Logger().Error("Vault is not an Asgard vault, transaction ignored.")
			continue
		}
		if vault.Status == InactiveVault {
			ctx.Logger().Error("Vault is inactive, transaction ignored.")
			continue
		}

		// tx is not observed at current vault - refund
		// yggdrasil pool is ok
		if ok := isCurrentVaultPubKey(ctx, h.keeper, tx); !ok {
			reason := fmt.Sprintf("vault %s is not current vault", tx.ObservedPubKey)
			ctx.Logger().Info("refund reason", reason)
			if err := refundTx(ctx, tx, txOutStore, h.keeper, constAccessor, CodeInvalidVault, reason, eventMgr); err != nil {
				return sdk.ErrInternal(err.Error()).Result()
			}
			continue
		}
		// chain is empty
		if tx.Tx.Chain.IsEmpty() {
			if err := refundTx(ctx, tx, txOutStore, h.keeper, constAccessor, CodeEmptyChain, "chain is empty", eventMgr); err != nil {
				return sdk.ErrInternal(err.Error()).Result()
			}
			continue
		}

		// construct msg from memo
		m, txErr := processOneTxIn(ctx, h.keeper, txIn, msg.Signer)
		if txErr != nil {
			ctx.Logger().Error("fail to process inbound tx", "error", txErr.Error(), "tx hash", tx.Tx.ID.String())
			if newErr := refundTx(ctx, tx, txOutStore, h.keeper, constAccessor, txErr.Code(), fmt.Sprint(txErr.Data()), eventMgr); nil != newErr {
				return sdk.ErrInternal(newErr.Error()).Result()
			}
			continue
		}

		if memo.IsOutbound() {
			// no one should send an outbound tx to vault
			continue
		}

		if err := h.keeper.SetLastChainHeight(ctx, tx.Tx.Chain, tx.BlockHeight); err != nil {
			return sdk.ErrInternal(err.Error()).Result()
		}

		// add addresses to observing addresses. This is used to detect
		// active/inactive observing node accounts
		obMgr.AppendObserver(tx.Tx.Chain, txIn.Signers)

		// check if we've halted trading
		_, isSwap := m.(MsgSwap)
		_, isStake := m.(MsgSetStakeData)
		haltTrading, err := h.keeper.GetMimir(ctx, "HaltTrading")
		if isSwap || isStake {
			if (haltTrading > 0 && haltTrading < ctx.BlockHeight() && err == nil) || h.keeper.RagnarokInProgress(ctx) {
				ctx.Logger().Info("trading is halted!!")
				if newErr := refundTx(ctx, tx, txOutStore, h.keeper, constAccessor, sdk.CodeUnauthorized, "trading halted", eventMgr); nil != newErr {
					return sdk.ErrInternal(newErr.Error()).Result()
				}
				continue
			}
		}

		// if its a swap, send it to our queue for processing later
		if isSwap {
			if err := h.keeper.SetSwapQueueItem(ctx, m.(MsgSwap)); err != nil {
				return sdk.ErrInternal(err.Error()).Result()
			}
			return sdk.Result{
				Code:      sdk.CodeOK,
				Codespace: DefaultCodespace,
			}
		}

		result := handler(ctx, m)
		if !result.IsOK() {
			refundMsg, err := getErrMessageFromABCILog(result.Log)
			if err != nil {
				ctx.Logger().Error(err.Error())
			}
			if err := refundTx(ctx, tx, txOutStore, h.keeper, constAccessor, result.Code, refundMsg, eventMgr); err != nil {
				return sdk.ErrInternal(err.Error()).Result()
			}
		}
	}
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
