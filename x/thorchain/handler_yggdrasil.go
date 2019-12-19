package thorchain

import (
	stdErrors "errors"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type YggdrasilHandler struct {
	keeper       Keeper
	txOutStore   TxOutStore
	poolAddrMgr  PoolAddressManager
	validatorMgr ValidatorManager
}

func NewYggdrasilHandler(keeper Keeper, txOutStore TxOutStore, poolAddrMgr PoolAddressManager, validatorMgr ValidatorManager) YggdrasilHandler {
	return YggdrasilHandler{
		keeper:       keeper,
		txOutStore:   txOutStore,
		poolAddrMgr:  poolAddrMgr,
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

func handleRagnarokProtocolStep2(ctx sdk.Context, keeper Keeper, txOut TxOutStore, poolAddrMgr PoolAddressManager, constAccessor constants.ConstantValues) sdk.Result {
	// Ragnarok Protocol
	// If THORNode can no longer be BFT, do a graceful shutdown of the entire network.
	// 1) THORNode will request all yggdrasil pool to return fund , if THORNode don't have yggdrasil pool THORNode will go to step 3 directly
	// 2) upon receiving the yggdrasil fund,  THORNode will refund the validator's bond
	// 3) once all yggdrasil fund get returned, return all fund to stakes

	// get the first observer
	nas, err := keeper.ListActiveNodeAccounts(ctx)
	if nil != err {
		ctx.Logger().Error("can't get active nodes", err)
		return sdk.ErrInternal("can't get active nodes").Result()
	}
	if len(nas) == 0 {
		return sdk.ErrInternal("can't find any active nodes").Result()
	}
	minimumNodesForBFT := constAccessor.GetInt64Value(constants.MinimumNodesForBFT)
	if int64(len(nas)) > minimumNodesForBFT { // THORNode still have enough validators for BFT
		// Ragnarok protocol didn't triggered , don't call this one
		return sdk.Result{
			Code:      sdk.CodeOK,
			Codespace: DefaultCodespace,
		}
	}

	pools, err := keeper.GetPools(ctx)
	if err != nil {
		ctx.Logger().Error("can't get pools", err)
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

			version := keeper.GetLowestActiveVersion(ctx)
			unstakeHandler := NewUnstakeHandler(keeper, txOut, poolAddrMgr)
			result := unstakeHandler.Run(ctx, unstakeMsg, version, constAccessor)
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

// Handle a message to set pooldata
func (h YggdrasilHandler) handleV1(ctx sdk.Context, msg MsgYggdrasil, constAccessor constants.ConstantValues) sdk.Result {
	ygg, err := h.keeper.GetVault(ctx, msg.PubKey)
	if nil != err && !stdErrors.Is(err, ErrVaultNotFound) {
		ctx.Logger().Error("fail to get yggdrasil", err)
		return sdk.ErrInternal(err.Error()).Result()
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

		hasYggdrasilPool, err := h.keeper.HasValidVaultPools(ctx)
		if nil != err {
			ctx.Logger().Error("fail to find valid yggdrasil pools", err)
			return sdk.ErrInternal(err.Error()).Result()
		}
		if !hasYggdrasilPool {
			return handleRagnarokProtocolStep2(ctx, h.keeper, h.txOutStore, h.poolAddrMgr, constAccessor)
		}
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
