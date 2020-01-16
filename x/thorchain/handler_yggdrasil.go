package thorchain

import (
	stdErrors "errors"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

// YggdrasilHandler is to process yggdrasil messages
type YggdrasilHandler struct {
	keeper       Keeper
	txOutStore   VersionedTxOutStore
	validatorMgr VersionedValidatorManager
}

// NewYggdrasilHandler create a new Yggdrasil handler
func NewYggdrasilHandler(keeper Keeper, txOutStore VersionedTxOutStore, validatorMgr VersionedValidatorManager) YggdrasilHandler {
	return YggdrasilHandler{
		keeper:       keeper,
		txOutStore:   txOutStore,
		validatorMgr: validatorMgr,
	}
}

// Run execute the logic in Yggdrasil Handler
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
		ctx.Logger().Error(errInvalidVersion.Error())
		return errInvalidVersion
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
		return h.handleV1(ctx, msg, version)
	} else {
		ctx.Logger().Error(errInvalidVersion.Error())
		return errBadVersion.Result()
	}
}

func (h YggdrasilHandler) handleV1(ctx sdk.Context, msg MsgYggdrasil, version semver.Version) sdk.Result {
	ygg, err := h.keeper.GetVault(ctx, msg.PubKey)
	if nil != err && !stdErrors.Is(err, ErrVaultNotFound) {
		ctx.Logger().Error("fail to get yggdrasil", "error", err)
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
		ctx.EventManager().EmitEvent(
			sdk.NewEvent("yggdrasil_fund",
				sdk.NewAttribute("pubkey", ygg.PubKey.String()),
				sdk.NewAttribute("coins", msg.Coins.String()),
				sdk.NewAttribute("tx", msg.Tx.ID.String())))

		// Yggdrasil usually comes from Asgard , Asgard --> Yggdrasil
		// It will be an outbound tx from Asgard pool , and it will be an Inbound tx form Yggdrasil pool
		// incoming fund will be added to Vault as part of ObservedTxInHandler
		// Yggdrasil handler doesn't need to do anything
		return sdk.Result{
			Code:      sdk.CodeOK,
			Codespace: DefaultCodespace,
		}
	}

	// Yggdrasil return fund back to Asgard
	ctx.EventManager().EmitEvent(
		sdk.NewEvent("yggdrasil_return",
			sdk.NewAttribute("pubkey", ygg.PubKey.String()),
			sdk.NewAttribute("coins", msg.Coins.String()),
			sdk.NewAttribute("tx", msg.Tx.ID.String())))

	na, err := h.keeper.GetNodeAccountByPubKey(ctx, msg.PubKey)
	if err != nil {
		ctx.Logger().Error("unable to get node account", "error", err)
		return sdk.ErrInternal(err.Error()).Result()
	}
	if na.Status == NodeActive {
		// node still active , no refund bond
		return sdk.Result{
			Code:      sdk.CodeOK,
			Codespace: DefaultCodespace,
		}
	}
	txOutStore, err := h.txOutStore.GetTxOutStore(h.keeper, version)
	if nil != err {
		ctx.Logger().Error("fail to get txout store", "error", err)
		return errBadVersion.Result()
	}
	if err := refundBond(ctx, msg.Tx, na, h.keeper, txOutStore); err != nil {
		ctx.Logger().Error("fail to refund bond", "error", err)
		return sdk.ErrInternal(err.Error()).Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
