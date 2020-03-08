package thorchain

import (
	"encoding/json"
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// UnstakeHandler
type UnstakeHandler struct {
	keeper     Keeper
	txOutStore VersionedTxOutStore
}

// NewUnstakeHandler create a new instance of UnstakeHandler to process unstake request
func NewUnstakeHandler(keeper Keeper, txOutStore VersionedTxOutStore) UnstakeHandler {
	return UnstakeHandler{
		keeper:     keeper,
		txOutStore: txOutStore,
	}
}

func (h UnstakeHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, _ constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgSetUnStake)
	if !ok {
		return errInvalidMessage.Result()
	}
	ctx.Logger().Info(fmt.Sprintf("receive MsgSetUnstake from : %s(%s) unstake (%s)", msg, msg.RuneAddress, msg.UnstakeBasisPoints))

	if err := h.validate(ctx, msg, version); err != nil {
		ctx.Logger().Error("msg ack failed validation", "error", err)
		return err.Result()
	}
	data, err := h.handle(ctx, msg, version)
	if err != nil {
		ctx.Logger().Error("fail to process msg unstake", "error", err)
		return err.Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Data:      data,
		Codespace: DefaultCodespace,
	}
}

func (h UnstakeHandler) validate(ctx sdk.Context, msg MsgSetUnStake, version semver.Version) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	} else {
		return errBadVersion
	}
}

func (h UnstakeHandler) validateV1(ctx sdk.Context, msg MsgSetUnStake) sdk.Error {
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error("unstake msg fail validation", "error", err.ABCILog())
		return sdk.NewError(DefaultCodespace, CodeUnstakeFailValidation, err.Error())
	}
	if !isSignedByActiveObserver(ctx, h.keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account",
			"request tx hash", msg.Tx.ID,
			"rune address", msg.RuneAddress,
			"asset", msg.Asset,
			"withdraw basis points", msg.UnstakeBasisPoints)
		return sdk.ErrUnauthorized("not authorized")
	}

	pool, err := h.keeper.GetPool(ctx, msg.Asset)
	if err != nil {
		errMsg := fmt.Sprintf("fail to get pool(%s)", msg.Asset)
		ctx.Logger().Error(errMsg, "error", err)
		return sdk.ErrInternal(errMsg)
	}

	if err := pool.EnsureValidPoolStatus(msg); err != nil {
		ctx.Logger().Error("fail to check pool status", "error", err)
		return sdk.NewError(DefaultCodespace, CodeInvalidPoolStatus, err.Error())
	}

	return nil
}

func (h UnstakeHandler) handle(ctx sdk.Context, msg MsgSetUnStake, version semver.Version) ([]byte, sdk.Error) {
	// Get rune (if any) and donate it to the reserve
	coin := msg.Tx.Coins.GetCoin(common.RuneAsset())
	if !coin.IsEmpty() {
		if err := h.keeper.AddFeeToReserve(ctx, coin.Amount); err != nil {
			// Add to reserve
			ctx.Logger().Error("fail to add fee to reserve", "error", err)
		}
	}

	poolStaker, err := h.keeper.GetPoolStaker(ctx, msg.Asset)
	if err != nil {
		ctx.Logger().Error("fail to get pool staker", "error", err)
		return nil, sdk.NewError(DefaultCodespace, CodeFailGetPoolStaker, "fail to get pool staker")
	}
	stakerUnit := poolStaker.GetStakerUnit(msg.RuneAddress)

	runeAmt, assetAmount, units, err := unstake(ctx, h.keeper, msg)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Errorf("fail to process UnStake request: %w", err).Error())
	}
	res, err := h.keeper.Cdc().MarshalBinaryLengthPrefixed(struct {
		Rune  sdk.Uint `json:"rune"`
		Asset sdk.Uint `json:"asset"`
	}{
		Rune:  runeAmt,
		Asset: assetAmount,
	})
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Errorf("fail to marshal result to json: %w", err).Error())
	}

	unstakeEvt := NewEventUnstake(
		msg.Asset,
		units,
		int64(msg.UnstakeBasisPoints.Uint64()),
		sdk.ZeroDec(), // TODO: What is Asymmetry, how to calculate it?
	)
	unstakeBytes, err := json.Marshal(unstakeEvt)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Errorf("fail to marshal event: %w", err).Error())
	}

	// unstake event is pending , once signer send the fund to customer successfully, then this should be marked as success
	evt := NewEvent(
		unstakeEvt.Type(),
		ctx.BlockHeight(),
		msg.Tx,
		unstakeBytes,
		EventPending,
	)

	if err := h.keeper.UpsertEvent(ctx, evt); err != nil {
		ctx.Logger().Error("fail to save event", "error", err)
		return nil, sdk.NewError(DefaultCodespace, CodeFailSaveEvent, "fail to save event")
	}
	txOutStore, err := h.txOutStore.GetTxOutStore(h.keeper, version)
	if err != nil {
		ctx.Logger().Error("fail to get txout store", "error", err)
		return nil, errBadVersion
	}

	memo := ""
	if msg.Tx.ID.Equals(common.BlankTxID) {
		// tx id is blank, must be triggered by the ragnarok protocol
		memo = NewRagnarokMemo(ctx.BlockHeight()).String()
	}
	toi := &TxOutItem{
		Chain:     common.BNBChain,
		InHash:    msg.Tx.ID,
		ToAddress: stakerUnit.RuneAddress,
		Coin:      common.NewCoin(common.RuneAsset(), runeAmt),
		Memo:      memo,
	}
	_, err = txOutStore.TryAddTxOutItem(ctx, toi)
	if err != nil {
		ctx.Logger().Error("fail to prepare outbound tx", "error", err)
		return nil, sdk.NewError(DefaultCodespace, CodeFailAddOutboundTx, "fail to prepare outbound tx")

	}

	toi = &TxOutItem{
		Chain:     msg.Asset.Chain,
		InHash:    msg.Tx.ID,
		ToAddress: stakerUnit.AssetAddress,
		Coin:      common.NewCoin(msg.Asset, assetAmount),
	}
	_, err = txOutStore.TryAddTxOutItem(ctx, toi)
	if err != nil {
		ctx.Logger().Error("fail to prepare outbound tx", "error", err)
		return nil, sdk.NewError(DefaultCodespace, CodeFailAddOutboundTx, "fail to prepare outbound tx")
	}

	return res, nil
}
