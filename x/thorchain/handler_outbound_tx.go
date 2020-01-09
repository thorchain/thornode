package thorchain

import (
	"github.com/blang/semver"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

type OutboundTxHandler struct {
	keeper Keeper
}

func NewOutboundTxHandler(keeper Keeper) OutboundTxHandler {
	return OutboundTxHandler{
		keeper: keeper,
	}
}

func (h OutboundTxHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, _ constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgOutboundTx)
	if !ok {
		return errInvalidMessage.Result()
	}
	if err := h.validate(ctx, msg, version); err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	return h.handle(ctx, msg, version)
}

func (h OutboundTxHandler) validate(ctx sdk.Context, msg MsgOutboundTx, version semver.Version) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	}
	ctx.Logger().Error(badVersion.Error())
	return badVersion
}

func (h OutboundTxHandler) validateV1(ctx sdk.Context, msg MsgOutboundTx) error {
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return err
	}

	if !isSignedByActiveObserver(ctx, h.keeper, msg.GetSigners()) {
		ctx.Logger().Error(notAuthorized.Error())
		return notAuthorized
	}
	return nil
}

func (h OutboundTxHandler) handle(ctx sdk.Context, msg MsgOutboundTx, version semver.Version) sdk.Result {
	ctx.Logger().Info("receive MsgOutboundTx", "request outbound tx hash", msg.Tx.Tx.ID)
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, msg)
	}
	ctx.Logger().Error(badVersion.Error())
	return errBadVersion.Result()
}

func (h OutboundTxHandler) handleV1(ctx sdk.Context, msg MsgOutboundTx) sdk.Result {
	voter, err := h.keeper.GetObservedTxVoter(ctx, msg.InTxID)
	if err != nil {
		ctx.Logger().Error(err.Error())
		return sdk.ErrInternal("fail to get observed tx voter").Result()
	}
	voter.AddOutTx(msg.Tx.Tx)
	h.keeper.SetObservedTxVoter(ctx, voter)

	// complete events
	if voter.IsDone() {
		err := completeEvents(ctx, h.keeper, msg.InTxID, voter.OutTxs, EventSuccess)
		if err != nil {
			ctx.Logger().Error("unable to complete events", "error", err)
			return sdk.ErrInternal(err.Error()).Result()
		}
	}

	// Apply Gas fees
	if err := AddGasFees(ctx, h.keeper, msg.Tx); nil != err {
		ctx.Logger().Error("fail to add gas fee", "error", err)
		return sdk.ErrInternal("fail to add gas fee").Result()
	}

	// update txOut record with our TxID that sent funds out of the pool
	txOut, err := h.keeper.GetTxOut(ctx, uint64(voter.Height))
	if err != nil {
		ctx.Logger().Error("unable to get txOut record", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	// Save TxOut back with the TxID only when the TxOut on the block height is
	// not empty
	if !txOut.IsEmpty() {
		for i, tx := range txOut.TxArray {

			// withdraw , refund etc, one inbound tx might result two outbound txes, THORNode have to correlate outbound tx back to the
			// inbound, and also txitem , thus THORNode could record both outbound tx hash correctly
			// given every tx item will only have one coin in it , THORNode could use that to identify which txit
			if tx.InHash.Equals(msg.InTxID) &&
				tx.OutHash.IsEmpty() &&
				msg.Tx.Tx.Coins.Contains(tx.Coin) {
				txOut.TxArray[i].OutHash = msg.Tx.Tx.ID
			}
		}
		if err := h.keeper.SetTxOut(ctx, txOut); nil != err {
			ctx.Logger().Error("fail to save tx out", "error", err)
			return sdk.ErrInternal("fail to save tx out").Result()
		}
	}
	h.keeper.SetLastSignedHeight(ctx, voter.Height)

	// If sending from one of our vaults, decrement coins
	if h.keeper.VaultExists(ctx, msg.Tx.ObservedPubKey) {
		vault, err := h.keeper.GetVault(ctx, msg.Tx.ObservedPubKey)
		if nil != err {
			ctx.Logger().Error("fail to get vault", "error", err)
			return sdk.ErrInternal("fail to get vault").Result()
		}
		vault.SubFunds(msg.Tx.Tx.Coins)
		if err := h.keeper.SetVault(ctx, vault); nil != err {
			ctx.Logger().Error("fail to save vault", "error", err)
			return sdk.ErrInternal("fail to save vault").Result()
		}
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
