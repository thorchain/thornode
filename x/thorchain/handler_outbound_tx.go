package thorchain

import (
	"fmt"

	"github.com/blang/semver"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
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
		return err.Result()
	}
	return h.handle(ctx, msg, version)
}

func (h OutboundTxHandler) validate(ctx sdk.Context, msg MsgOutboundTx, version semver.Version) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	}
	ctx.Logger().Error(errInvalidVersion.Error())
	return errBadVersion
}

func (h OutboundTxHandler) validateV1(ctx sdk.Context, msg MsgOutboundTx) sdk.Error {
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return err
	}

	if !isSignedByActiveObserver(ctx, h.keeper, msg.GetSigners()) {
		ctx.Logger().Error(notAuthorized.Error())
		return sdk.ErrUnauthorized("Not Authorized")
	}
	return nil
}

func (h OutboundTxHandler) handle(ctx sdk.Context, msg MsgOutboundTx, version semver.Version) sdk.Result {
	ctx.Logger().Info("receive MsgOutboundTx", "request outbound tx hash", msg.Tx.Tx.ID)
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, msg)
	}
	ctx.Logger().Error(errInvalidVersion.Error())
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

	// update txOut record with our TxID that sent funds out of the pool
	txOut, err := h.keeper.GetTxOut(ctx, voter.Height)
	if err != nil {
		ctx.Logger().Error("unable to get txOut record", "error", err)
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	// Save TxOut back with the TxID only when the TxOut on the block height is
	// not empty
	if !txOut.IsEmpty() {
		processedCoins := common.Coins{}
		for i, tx := range txOut.TxArray {

			c := msg.Tx.Tx.Coins.GetCoin(tx.Coin.Asset)
			if c.IsEmpty() {
				continue
			}
			// withdraw , refund etc, one inbound tx might result two outbound txes, THORNode have to correlate outbound tx back to the
			// inbound, and also txitem , thus THORNode could record both outbound tx hash correctly
			// given every tx item will only have one coin in it , THORNode could use that to identify which txit
			if tx.InHash.Equals(msg.InTxID) &&
				tx.OutHash.IsEmpty() &&
				msg.Tx.Tx.Coins.Contains(tx.Coin) {
				txOut.TxArray[i].OutHash = msg.Tx.Tx.ID
			}

			// fees might be taken from the txout , thus usually the amount send out from pool should a little bit less
			if c.Amount.GT(tx.Coin.Amount) {
				slashAmt := common.SafeSub(c.Amount, tx.Coin.Amount)
				// slash the difference from the node account's bond
				if err := h.slashNodeAccount(ctx, msg, tx.Coin.Asset, slashAmt); nil != err {
					ctx.Logger().Error("fail to slash account for sending extra fund", "error", err)
					return sdk.ErrInternal("fail to slash account").Result()
				}
			}
			processedCoins = append(processedCoins, c)
		}

		for _, c := range msg.Tx.Tx.Coins {
			if processedCoins.Contains(c) {
				continue
			}
			if err := h.slashNodeAccount(ctx, msg, c.Asset, c.Amount); nil != err {
				ctx.Logger().Error("fail to slash account for sending out extra fund", "error", err)
				return sdk.ErrInternal("fail to slash account").Result()
			}
		}
		if err := h.keeper.SetTxOut(ctx, txOut); nil != err {
			ctx.Logger().Error("fail to save tx out", "error", err)
			return sdk.ErrInternal("fail to save tx out").Result()
		}
	}
	h.keeper.SetLastSignedHeight(ctx, voter.Height)

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

// slashNodeAccount thorchain keep monitoring the outbound tx from asgard pool and yggdrasil pool, usually the txout is triggered by thorchain itself by
// adding an item into the txout array, refer to TxOutItem for the detail, the TxOutItem contains a specific coin and amount.
// if somehow thorchain discover signer send out fund more than the amount specified in TxOutItem, it will slash the node account who does that
// by taking 1.5 * extra fund from node account's bond and subsidise the pool that actually lost it.
func (h OutboundTxHandler) slashNodeAccount(ctx sdk.Context, msg MsgOutboundTx, asset common.Asset, slashAmount sdk.Uint) error {
	if slashAmount.IsZero() {
		return nil
	}
	thorAddr, err := msg.Tx.ObservedPubKey.GetThorAddress()
	if nil != err {
		return fmt.Errorf("fail to get thoraddress from pubkey(%s) %w", msg.Tx.ObservedPubKey, err)
	}
	nodeAccount, err := h.keeper.GetNodeAccount(ctx, thorAddr)
	if nil != err {
		return fmt.Errorf("fail to get node account with pubkey(%s), %w", msg.Tx.ObservedPubKey, err)
	}

	if asset.IsRune() {
		amountToReserve := slashAmount.QuoUint64(2)
		// if the diff asset is RUNE , just took 1.5 * diff from their bond
		slashAmount = slashAmount.MulUint64(3).QuoUint64(2)
		nodeAccount.Bond = common.SafeSub(nodeAccount.Bond, slashAmount)
		vaultData, err := h.keeper.GetVaultData(ctx)
		if nil != err {
			return fmt.Errorf("fail to get vault data: %w", err)
		}
		vaultData.TotalReserve = vaultData.TotalReserve.Add(amountToReserve)
		if err := h.keeper.SetVaultData(ctx, vaultData); nil != err {
			return fmt.Errorf("fail to save vault data: %w", err)
		}
		return h.keeper.SetNodeAccount(ctx, nodeAccount)
	}
	pool, err := h.keeper.GetPool(ctx, asset)
	if nil != err {
		return fmt.Errorf("fail to get %s pool : %w", asset, err)
	}
	// thorchain doesn't even have a pool for the asset, or the pool had been suspended, then who cares
	if pool.Empty() || pool.Status == PoolSuspended {
		return nil
	}
	runeValue := pool.AssetValueInRune(slashAmount).MulUint64(3).QuoUint64(2)
	pool.BalanceAsset = common.SafeSub(pool.BalanceAsset, slashAmount)
	pool.BalanceRune = pool.BalanceRune.Add(runeValue)
	nodeAccount.Bond = common.SafeSub(nodeAccount.Bond, runeValue)
	if err := h.keeper.SetPool(ctx, pool); nil != err {
		return fmt.Errorf("fail to save %s pool: %w", asset, err)
	}

	return h.keeper.SetNodeAccount(ctx, nodeAccount)
}
