package thorchain

import (
	stdErrors "errors"

	"github.com/blang/semver"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type YggdrasilHandler struct {
	keeper       Keeper
	txOutStore   *TxOutStore
	poolAddrMgr  *PoolAddressManager
	validatorMgr *ValidatorManager
}

func NewYggdrasilHandler(keeper Keeper, txOutStore *TxOutStore, poolAddrMgr *PoolAddressManager, validatorMgr *ValidatorManager) YggdrasilHandler {
	return YggdrasilHandler{
		keeper:       keeper,
		txOutStore:   txOutStore,
		poolAddrMgr:  poolAddrMgr,
		validatorMgr: validatorMgr,
	}
}

func (h YggdrasilHandler) Run(ctx sdk.Context, msg MsgYggdrasil, version semver.Version) sdk.Result {
	if err := h.Validate(ctx, msg, version); err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	if err := h.Handle(ctx, msg, version); err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func (h YggdrasilHandler) Validate(ctx sdk.Context, msg MsgYggdrasil, version semver.Version) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.ValidateV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

func (h YggdrasilHandler) ValidateV1(ctx sdk.Context, msg MsgYggdrasil) error {
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

func (h YggdrasilHandler) Handle(ctx sdk.Context, msg MsgYggdrasil, version semver.Version) error {
	ctx.Logger().Info("receive MsgYggdrasil", "pubkey", msg.PubKey.String(), "add_funds", msg.AddFunds, "coins", msg.Coins)
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.HandleV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

// Handle a message to set pooldata
func (h YggdrasilHandler) HandleV1(ctx sdk.Context, msg MsgYggdrasil) error {
	ygg, err := h.keeper.GetYggdrasil(ctx, msg.PubKey)
	if nil != err && !stdErrors.Is(err, ErrYggdrasilNotFound) {
		ctx.Logger().Error("fail to get yggdrasil", err)
		return err
	}
	if msg.AddFunds {
		ygg.AddFunds(msg.Coins)
	} else {
		ygg.SubFunds(msg.Coins)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent("yggdrasil_return",
				sdk.NewAttribute("pubkey", ygg.PubKey.String()),
				sdk.NewAttribute("coins", msg.Coins.String()),
				sdk.NewAttribute("tx", msg.RequestTxHash.String())))

		na, err := h.keeper.GetNodeAccountByPubKey(ctx, msg.PubKey)
		if err != nil {
			ctx.Logger().Error("unable to get node account", "error", err)
			return err
		}
		// TODO: slash their bond for any Yggdrasil funds that are unaccounted
		// for before sending their bond back. Keep in mind that THORNode won't get
		// back 100% of the funds (due to gas).
		RefundBond(ctx, msg.RequestTxHash, na, h.keeper, h.txOutStore)
	}
	if err := h.keeper.SetYggdrasil(ctx, ygg); nil != err {
		ctx.Logger().Error("fail to save yggdrasil", err)
		return err
	}

	// Ragnarok protocol get triggered, if all the Yggdrasil pool returned funds already, THORNode will continue Ragnarok
	if h.validatorMgr.Meta.Ragnarok {
		hasYggdrasilPool, err := h.keeper.HasValidYggdrasilPools(ctx)
		if nil != err {
			return err
		}
		if !hasYggdrasilPool {
			return handleRagnarokProtocolStep2(ctx, h.keeper, h.txOutStore, h.poolAddrMgr, h.validatorMgr)
		}
	}

	return nil
}
