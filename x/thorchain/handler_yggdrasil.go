package thorchain

import (
	stdErrors "errors"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

type YggdrasilHandler struct {
	keeper       Keeper
	txOutStore   TxOutStore
	validatorMgr ValidatorManager
}

func NewYggdrasilHandler(keeper Keeper, txOutStore TxOutStore, validatorMgr ValidatorManager) YggdrasilHandler {
	return YggdrasilHandler{
		keeper:       keeper,
		txOutStore:   txOutStore,
		validatorMgr: validatorMgr,
	}
}

func (h YggdrasilHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgYggdrasil)
	if !ok {
		return errInvalidMessage.Result()
	}
	if err := h.validate(ctx, msg, version); err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	return h.handle(ctx, msg, version, constAccessor)
}

func (h YggdrasilHandler) validate(ctx sdk.Context, msg MsgYggdrasil, version semver.Version) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

func (h YggdrasilHandler) validateV1(ctx sdk.Context, msg MsgYggdrasil) error {
	if err := msg.ValidateBasic(); nil != err {
		ctx.Logger().Error(err.Error())
		return err
	}

	if !isSignedByActiveNodeAccounts(ctx, h.keeper, msg.GetSigners()) {
		ctx.Logger().Error(notAuthorized.Error())
		return notAuthorized
	}
	return nil
}

func (h YggdrasilHandler) handle(ctx sdk.Context, msg MsgYggdrasil, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	ctx.Logger().Info("receive MsgYggdrasil", "pubkey", msg.PubKey.String(), "add_funds", msg.AddFunds, "coins", msg.Coins)
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, msg, constAccessor)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return errBadVersion.Result()
	}
}

// Handle a message to set pooldata
func (h YggdrasilHandler) handleV1(ctx sdk.Context, msg MsgYggdrasil, constAccessor constants.ConstantValues) sdk.Result {
	ygg, err := h.keeper.GetVault(ctx, msg.PubKey)
	if nil != err && !stdErrors.Is(err, ErrVaultNotFound) {
		ctx.Logger().Error("fail to get yggdrasil", err)
		return sdk.ErrInternal(err.Error()).Result()
	}
	if len(ygg.Type) == 0 {
		ygg.Status = ActiveVault
		ygg.Type = YggdrasilVault
	}
	if !ygg.IsYggdrasil() {
		return sdk.ErrInternal("this is not a Yggdrasil vault").Result()
	}
	if msg.AddFunds {
		ygg.AddFunds(msg.Coins)
	} else {
		ygg.SubFunds(msg.Coins)
	}

	if err := h.keeper.SetVault(ctx, ygg); nil != err {
		ctx.Logger().Error("fail to save yggdrasil", err)
		return sdk.ErrInternal(err.Error()).Result()
	}

	if !msg.AddFunds {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent("yggdrasil_return",
				sdk.NewAttribute("pubkey", ygg.PubKey.String()),
				sdk.NewAttribute("coins", msg.Coins.String()),
				sdk.NewAttribute("tx", msg.RequestTxHash.String())))

		na, err := h.keeper.GetNodeAccountByPubKey(ctx, msg.PubKey)
		if err != nil {
			ctx.Logger().Error("unable to get node account", "error", err)
			return sdk.ErrInternal(err.Error()).Result()
		}
		if err := refundBond(ctx, msg.RequestTxHash, na, h.keeper, h.txOutStore); err != nil {
			ctx.Logger().Error("fail to refund bond", err)
			return sdk.ErrInternal(err.Error()).Result()
		}
	}

	total, err := h.keeper.TotalActiveNodeAccount(ctx)
	if nil != err {
		ctx.Logger().Error("can't get active nodes", err)
		return sdk.ErrInternal("can't get active nodes").Result()
	}

	minimumNodesForBFT := constAccessor.GetInt64Value(constants.MinimumNodesForBFT)
	// Ragnarok protocol get triggered, if all the Yggdrasil pool returned funds already, THORNode will continue Ragnarok
	// THORNode still have enough validators for BFT
	if int64(total) < minimumNodesForBFT {
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
