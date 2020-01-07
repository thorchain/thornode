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
	txOutStore TxOutStore
}

// NewUnstakeHandler create a new instance of UnstakeHandler to process unstake request
func NewUnstakeHandler(keeper Keeper, txOutStore TxOutStore) UnstakeHandler {
	return UnstakeHandler{
		keeper:     keeper,
		txOutStore: txOutStore,
	}
}
func (uh UnstakeHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, _ constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgSetUnStake)
	if !ok {
		return errInvalidMessage.Result()
	}
	ctx.Logger().Info(fmt.Sprintf("receive MsgSetUnstake from : %s(%s) unstake (%s)", msg, msg.RuneAddress, msg.WithdrawBasisPoints))

	if err := uh.validate(ctx, msg, version); err != nil {
		ctx.Logger().Error("msg ack failed validation", err)
		return err.Result()
	}
	data, err := uh.handle(ctx, msg)
	if err != nil {
		ctx.Logger().Error("fail to process msg unstake", err)
		return err.Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Data:      data,
		Codespace: DefaultCodespace,
	}
}

func (uh UnstakeHandler) validate(ctx sdk.Context, msg MsgSetUnStake, version semver.Version) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return uh.validateV1(ctx, msg)
	} else {
		return errBadVersion
	}
}

func (uh UnstakeHandler) validateV1(ctx sdk.Context, msg MsgSetUnStake) sdk.Error {
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error("unstake msg fail validation", err.ABCILog())
		return sdk.NewError(DefaultCodespace, CodeUnstakeFailValidation, err.Error())
	}
	if !isSignedByActiveObserver(ctx, uh.keeper, msg.GetSigners()) {
		ctx.Logger().Error("message signed by unauthorized account",
			"request tx hash", msg.Tx.ID,
			"rune address", msg.RuneAddress,
			"asset", msg.Asset,
			"withdraw basis points", msg.WithdrawBasisPoints)
		return sdk.ErrUnauthorized("not authorized")
	}

	pool, err := uh.keeper.GetPool(ctx, msg.Asset)
	if err != nil {
		errMsg := fmt.Sprintf("fail to get pool(%s)", msg.Asset)
		ctx.Logger().Error(errMsg, err)
		return sdk.ErrInternal(errMsg)
	}

	if err := pool.EnsureValidPoolStatus(msg); nil != err {
		ctx.Logger().Error("fail to check pool status", err)
		return sdk.NewError(DefaultCodespace, CodeInvalidPoolStatus, err.Error())
	}

	return nil
}

func (uh UnstakeHandler) handle(ctx sdk.Context, msg MsgSetUnStake) ([]byte, sdk.Error) {
	poolStaker, err := uh.keeper.GetPoolStaker(ctx, msg.Asset)
	if nil != err {
		ctx.Logger().Error("fail to get pool staker: %w", err)
		return nil, sdk.NewError(DefaultCodespace, CodeFailGetPoolStaker, "fail to get pool staker")
	}
	stakerUnit := poolStaker.GetStakerUnit(msg.RuneAddress)

	runeAmt, assetAmount, units, err := unstake(ctx, uh.keeper, msg)
	if nil != err {
		return nil, sdk.ErrInternal(fmt.Errorf("fail to process UnStake request: %w", err).Error())
	}
	res, err := uh.keeper.Cdc().MarshalBinaryLengthPrefixed(struct {
		Rune  sdk.Uint `json:"rune"`
		Asset sdk.Uint `json:"asset"`
	}{
		Rune:  runeAmt,
		Asset: assetAmount,
	})
	if nil != err {
		return nil, sdk.ErrInternal(fmt.Errorf("fail to marshal result to json: %w", err).Error())
	}

	unstakeEvt := NewEventUnstake(
		msg.Asset,
		units,
		int64(msg.WithdrawBasisPoints.Uint64()),
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

	if err := uh.keeper.UpsertEvent(ctx, evt); nil != err {
		ctx.Logger().Error("fail to save event", err)
		return nil, sdk.NewError(DefaultCodespace, CodeFailSaveEvent, "fail to save event")
	}

	toi := &TxOutItem{
		Chain:     common.BNBChain,
		InHash:    msg.Tx.ID,
		ToAddress: stakerUnit.RuneAddress,
		Coin:      common.NewCoin(common.RuneAsset(), runeAmt),
	}
	_, err = uh.txOutStore.TryAddTxOutItem(ctx, toi)
	if err != nil {
		ctx.Logger().Error("fail to prepare outbound tx", err)
		return nil, sdk.NewError(DefaultCodespace, CodeFailAddOutboundTx, "fail to prepare outbound tx")

	}

	toi = &TxOutItem{
		Chain:     msg.Asset.Chain,
		InHash:    msg.Tx.ID,
		ToAddress: stakerUnit.AssetAddress,
		Coin:      common.NewCoin(msg.Asset, assetAmount),
	}
	_, err = uh.txOutStore.TryAddTxOutItem(ctx, toi)
	if err != nil {
		ctx.Logger().Error("fail to prepare outbound tx", err)
		return nil, sdk.NewError(DefaultCodespace, CodeFailAddOutboundTx, "fail to prepare outbound tx")
	}

	return res, nil
}
