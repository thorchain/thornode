package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// SwitchHandler is to handle Switch message
type SwitchHandler struct {
	keeper Keeper
}

// NewSwitchHandler create new instance of SwitchHandler
func NewSwitchHandler(keeper Keeper) SwitchHandler {
	return SwitchHandler{
		keeper: keeper,
	}
}

// Run it the main entry point to execute Switch logic
func (h SwitchHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgSwitch)
	if !ok {
		return errInvalidMessage.Result()
	}
	if err := h.validate(ctx, msg, version); err != nil {
		ctx.Logger().Error("msg switch failed validation", "error", err)
		return err.Result()
	}
	return h.handle(ctx, msg, version)
}

func (h SwitchHandler) validate(ctx sdk.Context, msg MsgSwitch, version semver.Version) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	} else {
		return errBadVersion
	}
}

func (h SwitchHandler) validateV1(ctx sdk.Context, msg MsgSwitch) sdk.Error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	if !isSignedByActiveNodeAccounts(ctx, h.keeper, msg.GetSigners()) {
		return sdk.ErrUnauthorized(notAuthorized.Error())
	}

	return nil
}

func (h SwitchHandler) handle(ctx sdk.Context, msg MsgSwitch, version semver.Version) sdk.Result {
	ctx.Logger().Info("handleMsgSwitch request", "destination address", msg.Destination.String())
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, msg)
	} else {
		ctx.Logger().Error(errInvalidVersion.Error())
		return errBadVersion.Result()
	}
}

func (h SwitchHandler) handleV1(ctx sdk.Context, msg MsgSwitch) sdk.Result {
	bank := h.keeper.CoinKeeper()

	coin, err := common.NewCoin(common.RuneNative, msg.Tx.Coins[0].Amount).Native()
	if err != nil {
		ctx.Logger().Error("fail to get native coin", "error", err)
		return sdk.ErrInternal("fail to get native coin").Result()
	}

	if _, err := bank.AddCoins(ctx, msg.Destination, sdk.NewCoins(coin)); err != nil {
		ctx.Logger().Error("fail to mint native rune coins", "error", err)
		return sdk.ErrInternal("fail to mint native rune coins").Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
