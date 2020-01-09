package thorchain

import (
	"encoding/json"
	"fmt"

	"github.com/blang/semver"
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/constants"
)

type SetAdminConfigHandler struct {
	keeper Keeper
}

func NewSetAdminConfigHandler(keeper Keeper) SetAdminConfigHandler {
	return SetAdminConfigHandler{
		keeper: keeper,
	}
}

func (h SetAdminConfigHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, _ constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgSetAdminConfig)
	if !ok {
		return errInvalidMessage.Result()
	}
	if err := h.validate(ctx, msg, version); err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	return h.handle(ctx, msg, version)
}

func (h SetAdminConfigHandler) validate(ctx sdk.Context, msg MsgSetAdminConfig, version semver.Version) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

func (h SetAdminConfigHandler) validateV1(ctx sdk.Context, msg MsgSetAdminConfig) error {
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

func (h SetAdminConfigHandler) handle(ctx sdk.Context, msg MsgSetAdminConfig, version semver.Version) sdk.Result {
	ctx.Logger().Info(fmt.Sprintf("receive MsgSetAdminConfig %s --> %s", msg.AdminConfig.Key, msg.AdminConfig.Value))
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.handleV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return errBadVersion.Result()
	}
}

func (h SetAdminConfigHandler) handleV1(ctx sdk.Context, msg MsgSetAdminConfig) sdk.Result {
	prevVal, err := h.keeper.GetAdminConfigValue(ctx, msg.AdminConfig.Key, nil)
	if err != nil {
		logError(ctx, err, "unable to get admin config")
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	h.keeper.SetAdminConfig(ctx, msg.AdminConfig)

	newVal, err := h.keeper.GetAdminConfigValue(ctx, msg.AdminConfig.Key, nil)
	if err != nil {
		logError(ctx, err, "unable to get admin config")
		return sdk.ErrUnknownRequest(err.Error()).Result()
	}

	if newVal != "" && prevVal != newVal {
		adminEvt := NewEventAdminConfig(
			msg.AdminConfig.Key.String(),
			msg.AdminConfig.Value,
		)
		stakeBytes, err := json.Marshal(adminEvt)
		if err != nil {
			logError(ctx, err, "fail to unmarshal admin config event")
			err = errors.Wrap(err, "fail to marshal admin config event to json")
			return sdk.ErrUnknownRequest(err.Error()).Result()
		}
		evt := NewEvent(
			adminEvt.Type(),
			ctx.BlockHeight(),
			msg.Tx,
			stakeBytes,
			EventSuccess,
		)
		if err := h.keeper.UpsertEvent(ctx, evt); nil != err {
			return sdk.ErrInternal(fmt.Errorf("fail to add event: %w", err).Error()).Result()
		}
	}
	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}
