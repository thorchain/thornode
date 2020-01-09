package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

// RefundHandler a handle to process tx that had refund memo
// usually this type or tx is because Thorchain fail to process the tx, which result in a refund, signer honour the tx and refund customer accordingly
type RefundHandler struct {
	keeper Keeper
}

// NewRefundHandler create a new refund handler
func NewRefundHandler(keeper Keeper) RefundHandler {
	return RefundHandler{keeper: keeper}
}
func (h RefundHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgRefundTx)
	if !ok {
		return errInvalidMessage.Result()
	}
	ctx.Logger().Info("receive MsgRefund",
		"tx ID", msg.InTxID.String())
	if err := h.validate(ctx, msg, version, constAccessor); nil != err {
		logError(ctx, err, "msg refund fail validation")
		return err.Result()
	}

	if err := h.handle(ctx, msg, version); nil != err {
		logError(ctx, err, "fail to process msg refund")
		return err.Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func (h RefundHandler) validate(ctx sdk.Context, msg MsgRefundTx, version semver.Version, constAccessor constants.ConstantValues) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, version, msg, constAccessor)
	}
	return errBadVersion
}

func (h RefundHandler) validateV1(ctx sdk.Context, version semver.Version, msg MsgRefundTx, constAccessor constants.ConstantValues) sdk.Error {
	if err := msg.ValidateBasic(); nil != err {
		return err
	}

	if !isSignedByActiveNodeAccounts(ctx, h.keeper, msg.GetSigners()) {
		return sdk.ErrUnauthorized("msg is not signed by an active node account")
	}
	return nil
}

func (h RefundHandler) handle(ctx sdk.Context, msg MsgRefundTx, version semver.Version) sdk.Error {
	voter, err := h.keeper.GetObservedTxVoter(ctx, msg.InTxID)
	if err != nil {
		ctx.Logger().Error(err.Error())
		return sdk.ErrInternal("fail to get observed tx voter")
	}
	voter.AddOutTx(msg.Tx.Tx)
	h.keeper.SetObservedTxVoter(ctx, voter)

	// complete events
	if voter.IsDone() {
		err := completeEvents(ctx, h.keeper, msg.InTxID, voter.OutTxs, EventRefund)
		if err != nil {
			return sdk.ErrInternal(fmt.Errorf("fail to set event to refund: %w", err).Error())
		}
	}

	// Apply Gas fees
	if err := AddGasFees(ctx, h.keeper, msg.Tx); nil != err {
		return sdk.ErrInternal(fmt.Errorf("fail to add gas fee: %w", err).Error())
	}

	// update txOut record with our TxID that sent funds out of the pool
	txOut, err := h.keeper.GetTxOut(ctx, uint64(voter.Height))
	if err != nil {
		return sdk.ErrUnknownRequest(fmt.Errorf("unable to get txout record: %w", err).Error())
	}

	// Save TxOut back with the TxID only when the TxOut on the block height is
	// not empty
	if !txOut.IsEmpty() {
		for i, tx := range txOut.TxArray {

			// withdraw , refund etc, one inbound tx might result two outbound txes, THORNode have to correlate outbound tx back to the
			// inbound, and also txitem , thus THORNode could record both outbound tx hash correctly
			// given every tx item will only have one coin in it , given that , THORNode could use that to identify which txit
			if tx.InHash.Equals(msg.InTxID) &&
				tx.OutHash.IsEmpty() &&
				msg.Tx.Tx.Coins.Contains(tx.Coin) {
				txOut.TxArray[i].OutHash = msg.Tx.Tx.ID
			}
		}
		if err := h.keeper.SetTxOut(ctx, txOut); nil != err {
			logError(ctx, err, "fail to save tx out")
			return sdk.ErrInternal(fmt.Errorf("fail to save tx out: %w", err).Error())
		}
	}
	h.keeper.SetLastSignedHeight(ctx, voter.Height)

	// If THORNode are sending from a yggdrasil pool, decrement coins on record
	if h.keeper.VaultExists(ctx, msg.Tx.ObservedPubKey) {
		ygg, err := h.keeper.GetVault(ctx, msg.Tx.ObservedPubKey)
		if nil != err {
			return sdk.ErrInternal(fmt.Errorf("fail to get yggdrasil: %w", err).Error())
		}
		ygg.SubFunds(msg.Tx.Tx.Coins)
		if err := h.keeper.SetVault(ctx, ygg); nil != err {
			return sdk.ErrInternal(fmt.Errorf("fail to save yggdrasil: %w", err).Error())
		}
	}
	return nil
}
