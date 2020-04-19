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

	if !isSignedByActiveNodeAccounts(ctx, h.keeper, msg.GetSigners()) {
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

func (h ErrataTxHandler) fetchEvents(ctx sdk.Context, msg MsgErrataTx) (Event, error) {
	eventIDs, err := h.keeper.GetEventsIDByTxHash(ctx, msg.TxID)
	if err != nil {
		errMsg := fmt.Sprintf("fail to get event ids by txhash(%s)", msg.TxID.String())
		ctx.Logger().Error(errMsg, "error", err)
	}

	if len(eventIDs) == 0 {
		return Event{}, fmt.Errorf("no event found for transaction id: %s", msg.TxID.String())
	}

	event, err := h.keeper.GetEvent(ctx, eventIDs[0])
	if err != nil {
		ctx.Logger().Error("fail to get event", "id", msg.TxID, "error", err)
	}

	return event, err
}

func (h ErrataTxHandler) handleV1(ctx sdk.Context, msg MsgErrataTx) sdk.Result {
	active, err := h.keeper.ListActiveNodeAccounts(ctx)
	if err != nil {
		err = wrapError(ctx, err, "fail to get list of active node accounts")
		return sdk.ErrInternal(err.Error()).Result()
	}

	voter, err := h.keeper.GetErrataTxVoter(ctx, msg.TxID, msg.Chain)
	if err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}

	voter.Sign(msg.Signer)
	h.keeper.SetErrataTxVoter(ctx, voter)
	// doesn't have consensus yet
	if !voter.HasConsensus(active) {
		ctx.Logger().Info("not having consensus yet, return")
		return sdk.Result{
			Code:      sdk.CodeOK,
			Codespace: DefaultCodespace,
		}
	}

	if voter.BlockHeight > 0 {
		// errata tx already processed
		return sdk.Result{
			Code:      sdk.CodeOK,
			Codespace: DefaultCodespace,
		}
	}

	voter.BlockHeight = ctx.BlockHeight()
	h.keeper.SetErrataTxVoter(ctx, voter)

	// fetch events
	event, err := h.fetchEvents(ctx, msg)
	if err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}

	tx := event.InTx

	if !tx.Chain.Equals(msg.Chain) {
		// does not match chain
		return sdk.Result{
			Code:      sdk.CodeOK,
			Codespace: DefaultCodespace,
		}
	}

	memo, _ := ParseMemo(tx.Memo)
	if !memo.IsType(TxSwap) && !memo.IsType(TxStake) {
		// must be a swap transaction
		return sdk.Result{
			Code:      sdk.CodeOK,
			Codespace: DefaultCodespace,
		}
	}

	// fetch pool from memo
	pool, err := h.keeper.GetPool(ctx, memo.GetAsset())
	if err != nil {
		ctx.Logger().Error("fail to get pool for errata tx", "error", err)
		return sdk.ErrInternal(err.Error()).Result()
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

	if memo.IsType(TxStake) {
		ps, err := h.keeper.GetPoolStaker(ctx, memo.GetAsset())
		if err != nil {
			ctx.Logger().Error("fail to get pool staker record", "error", err)
			return sdk.ErrInternal(err.Error()).Result()
		}

		// since this address is being malicious, zero their staking units
		su := ps.GetStakerUnit(tx.FromAddress)
		pool.PoolUnits = common.SafeSub(pool.PoolUnits, su.Units)
		ps.TotalUnits = pool.PoolUnits
		su.Units = sdk.ZeroUint()

		ps.UpsertStakerUnit(su)
		h.keeper.SetPoolStaker(ctx, ps)
	}

	if err := h.keeper.SetPool(ctx, pool); err != nil {
		ctx.Logger().Error("fail to save pool", "error", err)
	}

	// send errata event
	mods := PoolMods{
		NewPoolMod(pool.Asset, runeAmt, false, assetAmt, false),
	}

	eventErrata := NewEventErrata(mods)
	errataBuf, err := json.Marshal(eventErrata)
	if err != nil {
		ctx.Logger().Error("fail to marshal errata event to buf", "error", err)
	}
	evt := NewEvent(
		eventErrata.Type(),
		ctx.BlockHeight(),
		common.Tx{ID: msg.TxID},
		errataBuf,
		EventSuccess,
	)
	if err := h.keeper.UpsertEvent(ctx, evt); err != nil {
		ctx.Logger().Error("fail to save errata event", "error", err)
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
