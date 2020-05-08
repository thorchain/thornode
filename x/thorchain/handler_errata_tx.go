package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// ErrataTxHandler is to handle ErrataTx message
type ErrataTxHandler struct {
	keeper                Keeper
	versionedEventManager VersionedEventManager
}

// NewErrataTxHandler create new instance of ErrataTxHandler
func NewErrataTxHandler(keeper Keeper, versionedEventManager VersionedEventManager) ErrataTxHandler {
	return ErrataTxHandler{
		keeper:                keeper,
		versionedEventManager: versionedEventManager,
	}
}

// Run it the main entry point to execute ErrataTx logic
func (h ErrataTxHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, _ constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgErrataTx)
	if !ok {
		return errInvalidMessage.Result()
	}
	if err := h.validate(ctx, msg, version); err != nil {
		ctx.Logger().Error("msg errata tx failed validation", "error", err)
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
		return h.handleV1(ctx, msg, version)
	} else {
		ctx.Logger().Error(errInvalidVersion.Error())
		return errBadVersion.Result()
	}
}

func (h ErrataTxHandler) handleV1(ctx sdk.Context, msg MsgErrataTx, version semver.Version) sdk.Result {
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

	observedVoter, err := h.keeper.GetObservedTxVoter(ctx, msg.TxID)
	if err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	if observedVoter.Tx.IsEmpty() {
		return sdk.ErrInternal(fmt.Sprintf("cannot find tx: %s", msg.TxID)).Result()
	}

	tx := observedVoter.Tx.Tx

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
		staker, err := h.keeper.GetStaker(ctx, memo.GetAsset(), tx.FromAddress)
		if err != nil {
			ctx.Logger().Error("fail to get staker", "error", err)
			return sdk.ErrInternal(err.Error()).Result()
		}

		// since this address is being malicious, zero their staking units
		pool.PoolUnits = common.SafeSub(pool.PoolUnits, staker.Units)
		staker.Units = sdk.ZeroUint()
		staker.LastStakeHeight = ctx.BlockHeight()

		h.keeper.SetStaker(ctx, staker)
	}

	if err := h.keeper.SetPool(ctx, pool); err != nil {
		ctx.Logger().Error("fail to save pool", "error", err)
	}

	// send errata event
	mods := PoolMods{
		NewPoolMod(pool.Asset, runeAmt, false, assetAmt, false),
	}

	eventErrata := NewEventErrata(msg.TxID, mods)
	eventMgr, err := h.versionedEventManager.GetEventManager(ctx, version)
	if err != nil {
		return errFailGetEventManager.Result()
	}
	if err := eventMgr.EmitErrataEvent(ctx, h.keeper, msg.TxID, eventErrata); err != nil {
		ctx.Logger().Error("fail to emit errata event", "error", err)
		return sdk.ErrInternal("fail to emit errata event").Result()
	}
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
