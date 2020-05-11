package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// UnstakeHandler to process unstake requests
type UnstakeHandler struct {
	keeper                Keeper
	txOutStore            VersionedTxOutStore
	versionedEventManager VersionedEventManager
}

// NewUnstakeHandler create a new instance of UnstakeHandler to process unstake request
func NewUnstakeHandler(keeper Keeper, txOutStore VersionedTxOutStore, versionedEventManager VersionedEventManager) UnstakeHandler {
	return UnstakeHandler{
		keeper:                keeper,
		txOutStore:            txOutStore,
		versionedEventManager: versionedEventManager,
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
	if !isSignedByActiveNodeAccounts(ctx, h.keeper, msg.GetSigners()) {
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
	staker, err := h.keeper.GetStaker(ctx, msg.Asset, msg.RuneAddress)
	if err != nil {
		ctx.Logger().Error("fail to get staker", "error", err)
		return nil, sdk.NewError(DefaultCodespace, CodeFailGetStaker, "fail to get staker")
	}
	eventManager, err := h.versionedEventManager.GetEventManager(ctx, version)
	if err != nil {
		ctx.Logger().Error("fail to get event manager", "error", err)
		return nil, errFailGetEventManager
	}
	runeAmt, assetAmount, units, gasAsset, err := unstake(ctx, version, h.keeper, msg, eventManager)
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
		msg.Tx,
	)
	if err := eventManager.EmitUnstakeEvent(ctx, h.keeper, unstakeEvt); err != nil {
		ctx.Logger().Error("fail to emit unstake event", "error", err)
		return nil, sdk.NewError(DefaultCodespace, CodeFailSaveEvent, "fail to save unstake event")
	}
	txOutStore, err := h.txOutStore.GetTxOutStore(ctx, h.keeper, version)
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
		Chain:     common.RuneAsset().Chain,
		InHash:    msg.Tx.ID,
		ToAddress: staker.RuneAddress,
		Coin:      common.NewCoin(common.RuneAsset(), runeAmt),
		Memo:      memo,
	}
	if !gasAsset.IsZero() {
		if msg.Asset.IsBNB() {
			toi.MaxGas = common.Gas{
				common.NewCoin(common.RuneAsset().Chain.GetGasAsset(), gasAsset.QuoUint64(2)),
			}
		}
	}
	ok, err := txOutStore.TryAddTxOutItem(ctx, toi)
	if err != nil {
		ctx.Logger().Error("fail to prepare outbound tx", "error", err)
		return nil, sdk.NewError(DefaultCodespace, CodeFailAddOutboundTx, "fail to prepare outbound tx")
	}
	if !ok {
		return nil, sdk.NewError(DefaultCodespace, CodeFailAddOutboundTx, "prepare outbound tx not successful")
	}

	toi = &TxOutItem{
		Chain:     msg.Asset.Chain,
		InHash:    msg.Tx.ID,
		ToAddress: staker.AssetAddress,
		Coin:      common.NewCoin(msg.Asset, assetAmount),
		Memo:      memo,
	}
	if !gasAsset.IsZero() {
		if msg.Asset.IsBNB() {
			toi.MaxGas = common.Gas{
				common.NewCoin(common.RuneAsset().Chain.GetGasAsset(), gasAsset.QuoUint64(2)),
			}
		} else if msg.Asset.Chain.GetGasAsset().Equals(msg.Asset) {
			toi.MaxGas = common.Gas{
				common.NewCoin(msg.Asset.Chain.GetGasAsset(), gasAsset),
			}
		}
	}

	ok, err = txOutStore.TryAddTxOutItem(ctx, toi)
	if err != nil {
		ctx.Logger().Error("fail to prepare outbound tx", "error", err)
		return nil, sdk.NewError(DefaultCodespace, CodeFailAddOutboundTx, "fail to prepare outbound tx")
	}
	if !ok {
		return nil, sdk.NewError(DefaultCodespace, CodeFailAddOutboundTx, "prepare outbound tx not successful")
	}

	// Get rune (if any) and donate it to the reserve
	coin := msg.Tx.Coins.GetCoin(common.RuneAsset())
	if !coin.IsEmpty() {
		if err := h.keeper.AddFeeToReserve(ctx, coin.Amount); err != nil {
			// Add to reserve
			ctx.Logger().Error("fail to add fee to reserve", "error", err)
		}
	}

	return res, nil
}
