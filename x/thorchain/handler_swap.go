package thorchain

import (
	"encoding/json"
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type SwapHandler struct {
	keeper              Keeper
	versionedTxOutStore VersionedTxOutStore
}

func NewSwapHandler(keeper Keeper, versionedTxOutStore VersionedTxOutStore) SwapHandler {
	return SwapHandler{
		keeper:              keeper,
		versionedTxOutStore: versionedTxOutStore,
	}
}

func (h SwapHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgSwap)
	if !ok {
		return errInvalidMessage.Result()
	}
	if err := h.validate(ctx, msg, version); err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	return h.handle(ctx, msg, version, constAccessor)
}

func (h SwapHandler) validate(ctx sdk.Context, msg MsgSwap, version semver.Version) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	} else {
		ctx.Logger().Error(errInvalidVersion.Error())
		return errInvalidVersion
	}
}

func (h SwapHandler) validateV1(ctx sdk.Context, msg MsgSwap) error {
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

func (h SwapHandler) handle(ctx sdk.Context, msg MsgSwap, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	ctx.Logger().Info("receive MsgSwap", "request tx hash", msg.Tx.ID, "source asset", msg.Tx.Coins[0].Asset, "target asset", msg.TargetAsset, "signer", msg.Signer.String())
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, msg, version, constAccessor)
	} else {
		ctx.Logger().Error(errInvalidVersion.Error())
		return errBadVersion.Result()
	}
}

func (h SwapHandler) addSwapEvent(ctx sdk.Context, swapEvt EventSwap, tx common.Tx, status EventStatus) error {
	swapBytes, err := json.Marshal(swapEvt)
	if err != nil {
		return err
	}
	evt := NewEvent(swapEvt.Type(), ctx.BlockHeight(), tx, swapBytes, status)
	if err := h.keeper.UpsertEvent(ctx, evt); err != nil {
		return err
	}
	return nil
}

func (h SwapHandler) handleV1(ctx sdk.Context, msg MsgSwap, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	transactionFee := constAccessor.GetInt64Value(constants.TransactionFee)
	amount, swapEvents, swapErr := swap(
		ctx,
		h.keeper,
		msg.Tx,
		msg.TargetAsset,
		msg.Destination,
		msg.TradeTarget,
		sdk.NewUint(uint64(transactionFee)))
	if swapErr != nil {
		ctx.Logger().Error("fail to process swap message", "error", swapErr)
		return swapErr.Result()
	}
	for _, item := range swapEvents {
		if eventErr := h.addSwapEvent(ctx, item, msg.Tx, EventPending); eventErr != nil {
			return sdk.ErrInternal(eventErr.Error()).Result()
		}
	}

	res, err := h.keeper.Cdc().MarshalBinaryLengthPrefixed(
		struct {
			Asset sdk.Uint `json:"asset"`
		}{
			Asset: amount,
		})

	if err != nil {
		ctx.Logger().Error("fail to encode result to json", "error", err)
		return sdk.ErrInternal("fail to encode result to json").Result()
	}
	txOutStore, err := h.versionedTxOutStore.GetTxOutStore(h.keeper, version)
	if err != nil {
		ctx.Logger().Error("fail to get txout store", "error", err)
		return errBadVersion.Result()
	}
	toi := &TxOutItem{
		Chain:     msg.TargetAsset.Chain,
		InHash:    msg.Tx.ID,
		ToAddress: msg.Destination,
		Coin:      common.NewCoin(msg.TargetAsset, amount),
	}
	_, err = txOutStore.TryAddTxOutItem(ctx, toi)
	if err != nil {
		ctx.Logger().Error("fail to add outbound tx", "error", err)
		return sdk.ErrInternal(fmt.Errorf("fail to add outbound tx: %w", err).Error()).Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Data:      res,
		Codespace: DefaultCodespace,
	}
}
