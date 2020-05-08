package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type SendHandler struct {
	keeper Keeper
}

func NewSendHandler(keeper Keeper) SendHandler {
	return SendHandler{
		keeper: keeper,
	}
}

func (h SendHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgSend)
	if !ok {
		return errInvalidMessage.Result()
	}
	if err := h.validate(ctx, msg, version); err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	return h.handle(ctx, msg, version, constAccessor)
}

func (h SendHandler) validate(ctx sdk.Context, msg MsgSend, version semver.Version) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	} else {
		ctx.Logger().Error(errInvalidVersion.Error())
		return errInvalidVersion
	}
}

func (h SendHandler) validateV1(ctx sdk.Context, msg MsgSend) error {
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error(err.Error())
		return err
	}

	return nil
}

func (h SendHandler) handle(ctx sdk.Context, msg MsgSend, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	ctx.Logger().Info("receive MsgSend", "from", msg.FromAddress, "to", msg.ToAddress, "coins", msg.Amount)
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, msg, version, constAccessor)
	} else {
		ctx.Logger().Error(errInvalidVersion.Error())
		return errBadVersion.Result()
	}
}

func (h SendHandler) handleV1(ctx sdk.Context, msg MsgSend, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	banker := h.keeper.CoinKeeper()
	supplier := h.keeper.Supply()
	// TODO: this shouldn't be tied to swaps, and should be cheaper. But
	// TransactionFee will be fine for now.
	transactionFee := constAccessor.GetInt64Value(constants.TransactionFee)

	gasFee, err := common.NewCoin(common.RuneNative, sdk.NewUint(uint64(transactionFee))).Native()
	if err != nil {
		ctx.Logger().Error("fail to get gas fee", "err", err)
		return sdk.ErrInternal("fail to get gas fee").Result()
	}

	totalCoins := sdk.NewCoins(gasFee).Add(msg.Amount)
	if !banker.HasCoins(ctx, msg.FromAddress, totalCoins) {
		ctx.Logger().Error("insufficient funds", "error", err)
		return sdk.ErrInsufficientCoins("insufficient funds").Result()
	}

	// send gas to reserve
	sdkErr := supplier.SendCoinsFromAccountToModule(ctx, msg.FromAddress, ReserveName, sdk.NewCoins(gasFee))
	if sdkErr != nil {
		ctx.Logger().Error("unable to send gas to reserve", "error", sdkErr)
		return sdkErr.Result()
	}

	sdkErr = banker.SendCoins(ctx, msg.FromAddress, msg.ToAddress, msg.Amount)
	if sdkErr != nil {
		return sdkErr.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return sdk.Result{
		Events:    ctx.EventManager().Events(),
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
