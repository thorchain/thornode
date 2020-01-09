package thorchain

import (
	"github.com/blang/semver"
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

type EndPoolHandler struct {
	keeper     Keeper
	txOutStore TxOutStore
}

func NewEndPoolHandler(keeper Keeper, txOutStore TxOutStore) EndPoolHandler {
	return EndPoolHandler{
		keeper:     keeper,
		txOutStore: txOutStore,
	}
}

func (h EndPoolHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgEndPool)
	if !ok {
		return errInvalidMessage.Result()
	}
	if err := h.validate(ctx, msg, version); err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	return h.handle(ctx, msg, version, constAccessor)
}

func (h EndPoolHandler) validate(ctx sdk.Context, msg MsgEndPool, version semver.Version) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	} else {
		logError(ctx, badVersion, "")
		return badVersion
	}
}

func (h EndPoolHandler) validateV1(ctx sdk.Context, msg MsgEndPool) error {
	if err := msg.ValidateBasic(); nil != err {
		logError(ctx, err, "")
		return err
	}

	if !isSignedByActiveNodeAccounts(ctx, h.keeper, msg.GetSigners()) {
		logError(ctx, notAuthorized, "")
		return notAuthorized
	}
	return nil
}

func (h EndPoolHandler) handle(ctx sdk.Context, msg MsgEndPool, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	ctx.Logger().Info("receive MsgEndPool", "asset", msg.Asset, "requester", msg.Tx.FromAddress, "signer", msg.Signer.String())
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, msg, version, constAccessor)
	} else {
		logError(ctx, badVersion, "")
		return errBadVersion.Result()
	}
}

func (h EndPoolHandler) handleV1(ctx sdk.Context, msg MsgEndPool, version semver.Version, constAccessor constants.ConstantValues) sdk.Result {
	poolStaker, err := h.keeper.GetPoolStaker(ctx, msg.Asset)
	if nil != err {
		logError(ctx, err, "fail to get pool staker")
		return sdk.ErrInternal(err.Error()).Result()
	}

	// everyone withdraw
	for _, item := range poolStaker.Stakers {
		unstakeMsg := NewMsgSetUnStake(
			msg.Tx,
			item.RuneAddress,
			sdk.NewUint(10000),
			msg.Asset,
			msg.Signer,
		)
		unstakeHandler := NewUnstakeHandler(h.keeper, h.txOutStore)
		result := unstakeHandler.Run(ctx, unstakeMsg, version, constAccessor)
		if !result.IsOK() {
			logError(ctx, err, "fail to unstake: %s", item.RuneAddress)
			return result
		}
	}
	pool, err := h.keeper.GetPool(ctx, msg.Asset)
	pool.Status = PoolBootstrap
	if err := h.keeper.SetPool(ctx, pool); err != nil {
		err = errors.Wrap(err, "fail to set pool")
		return sdk.ErrInternal(err.Error()).Result()
	}
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
