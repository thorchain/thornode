package thorchain

import (
	"encoding/json"
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// ErrataTxHandler is to handle ErrataTx message
type ErrataTxHandler struct {
	keeper Keeper
}

// NewErrataTxHandler create new instance of ErrataTxHandler
func NewErrataTxHandler(keeper Keeper) ErrataTxHandler {
	return ErrataTxHandler{
		keeper: keeper,
	}
}

// Run it the main entry point to execute ErrataTx logic
func (h ErrataTxHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, _ constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgErrataTx)
	if !ok {
		return errInvalidMessage.Result()
	}
	if err := h.validate(ctx, msg, version); err != nil {
		ctx.Logger().Error("msg set version failed validation", "error", err)
		return err.Result()
	}
	return h.handle(ctx, msg, version)
}

func (h ErrataTxHandler) validate(ctx sdk.Context, msg MsgErrataTx, version semver.Version) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	} else {
		return errBadVersion
	}
}

func (h ErrataTxHandler) validateV1(ctx sdk.Context, msg MsgErrataTx) sdk.Error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	if !isSignedByActiveObserver(ctx, h.keeper, msg.GetSigners()) {
		return sdk.ErrUnauthorized(notAuthorized.Error())
	}

	return nil
}

func (h ErrataTxHandler) handle(ctx sdk.Context, msg MsgErrataTx, version semver.Version) sdk.Result {
	ctx.Logger().Info("handleMsgErrataTx request", "txid", msg.TxID.String())
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, msg)
	} else {
		ctx.Logger().Error(errInvalidVersion.Error())
		return errBadVersion.Result()
	}
}

func (h ErrataTxHandler) handleV1(ctx sdk.Context, msg MsgErrataTx) sdk.Result {

	eventIDs, err := h.keeper.GetEventsIDByTxHash(ctx, msg.TxID)
	if err != nil {
		errMsg := fmt.Sprintf("fail to get event ids by txhash(%s)", msg.TxID.String())
		ctx.Logger().Error(errMsg, "error", err)
		return sdk.ErrInternal(errMsg).Result()
	}

	// collect events with the given hash
	events := make(Events, 0)
	for _, id := range eventIDs {
		event, err := h.keeper.GetEvent(ctx, id)
		if err != nil {
			errMsg := fmt.Sprintf("fail to get event(%d)", id)
			ctx.Logger().Error(errMsg, "error", err)
			return sdk.ErrInternal(errMsg).Result()
		}

		if event.Empty() {
			continue
		}
		events = append(events, event)
	}

	mods := make(PoolMods, 0)

	// revert each in tx
	for _, event := range events {
		tx := event.InTx

		if !tx.Chain.Equals(msg.Chain) {
			// does not match chain
			continue
		}

		memo, _ := ParseMemo(tx.Memo)

		if !memo.IsType(txSwap) {
			// must be a swap transaction
			continue
		}

		// fetch pool from memo
		pool, err := h.keeper.GetPool(ctx, memo.GetAsset())
		if err != nil {
			ctx.Logger().Error("fail to get pool for errata tx", "error", err)
			continue
		}

		// subtract amounts from pool balances
		runeAmt := sdk.ZeroUint()
		assetAmt := sdk.ZeroUint()
		for _, coin := range tx.Coins {
			if coin.Asset.IsRune() {
				runeAmt = coin.Amount
			} else {
				assetAmt = coin.Amount
			}
		}

		pool.BalanceRune = common.SafeSub(pool.BalanceRune, runeAmt)
		pool.BalanceAsset = common.SafeSub(pool.BalanceAsset, assetAmt)
		mods = append(
			mods,
			NewPoolMod(pool.Asset, runeAmt, false, assetAmt, false),
		)

		if err := h.keeper.SetPool(ctx, pool); err != nil {
			ctx.Logger().Error("fail to save pool", "error", err)
		}
	}

	eventErrata := NewEventErrata(mods)
	errataBuf, err := json.Marshal(eventErrata)
	if err != nil {
		ctx.Logger().Error("fail to marshal errata event to buf", "error", err)
	}
	event := NewEvent(
		eventErrata.Type(),
		ctx.BlockHeight(),
		common.Tx{ID: msg.TxID},
		errataBuf,
		EventSuccess,
	)
	if err := h.keeper.UpsertEvent(ctx, event); err != nil {
		ctx.Logger().Error("fail to save errata event", "error", err)
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
