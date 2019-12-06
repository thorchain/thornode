package thorchain

import (
	stdErrors "errors"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/common"

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

func handleRagnarokProtocolStep2(ctx sdk.Context, keeper Keeper, txOut *TxOutStore, poolAddrMgr *PoolAddressManager, validatorManager *ValidatorManager) sdk.Result {
	// Ragnarok Protocol
	// If THORNode can no longer be BFT, do a graceful shutdown of the entire network.
	// 1) THORNode will request all yggdrasil pool to return fund , if THORNode don't have yggdrasil pool THORNode will go to step 3 directly
	// 2) upon receiving the yggdrasil fund,  THORNode will refund the validator's bond
	// 3) once all yggdrasil fund get returned, return all fund to stakes
	if !validatorManager.Meta.Ragnarok {
		// Ragnarok protocol didn't triggered , don't call this one
		return sdk.Result{
			Code:      sdk.CodeOK,
			Codespace: DefaultCodespace,
		}
	}
	// get the first observer
	nas, err := keeper.ListActiveNodeAccounts(ctx)
	if nil != err {
		ctx.Logger().Error("can't get active nodes", err)
		return sdk.ErrInternal("can't get active nodes").Result()
	}
	if len(nas) == 0 {
		return sdk.ErrInternal("can't find any active nodes").Result()
	}

	pools, err := keeper.GetPools(ctx)
	if err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}

	// go through all the pooles
	for _, pool := range pools {
		poolStaker, err := keeper.GetPoolStaker(ctx, pool.Asset)
		if nil != err {
			ctx.Logger().Error("fail to get pool staker", err)
			return sdk.ErrInternal(err.Error()).Result()
		}

		// everyone withdraw
		for _, item := range poolStaker.Stakers {
			unstakeMsg := NewMsgSetUnStake(
				common.GetRagnarokTx(pool.Asset.Chain),
				item.RuneAddress,
				sdk.NewUint(10000),
				pool.Asset,
				nas[0].NodeAddress,
			)

			result := handleMsgSetUnstake(ctx, keeper, txOut, poolAddrMgr, unstakeMsg)
			if !result.IsOK() {
				ctx.Logger().Error("fail to unstake", "staker", item.RuneAddress)
				return result
			}
		}
		pool.Status = PoolSuspended
		if err := keeper.SetPool(ctx, pool); err != nil {
			err = errors.Wrap(err, "fail to set pool")
			ctx.Logger().Error(err.Error())
			return sdk.ErrInternal(err.Error()).Result()
		}
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}

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
