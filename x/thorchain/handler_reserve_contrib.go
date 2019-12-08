package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/constants"
)

type ReserveContributorHandler struct {
	keeper Keeper
}

func NewReserveContributorHandler(keeper Keeper) ReserveContributorHandler {
	return ReserveContributorHandler{
		keeper: keeper,
	}
}

func (h ReserveContributorHandler) Run(ctx sdk.Context, m sdk.Msg, consts constants.Constants, version semver.Version) sdk.Result {
	msg, ok := m.(MsgReserveContributor)
	if !ok {
		return errInvalidMessage.Result()
	}
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

func (h ReserveContributorHandler) Validate(ctx sdk.Context, msg MsgReserveContributor, version semver.Version) error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.ValidateV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

func (h ReserveContributorHandler) ValidateV1(ctx sdk.Context, msg MsgReserveContributor) error {
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

func (h ReserveContributorHandler) Handle(ctx sdk.Context, msg MsgReserveContributor, version semver.Version) error {
	ctx.Logger().Info("handleMsgReserveContributor request")
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.HandleV1(ctx, msg)
	} else {
		ctx.Logger().Error(badVersion.Error())
		return badVersion
	}
}

// Handle a message to set pooldata
func (h ReserveContributorHandler) HandleV1(ctx sdk.Context, msg MsgReserveContributor) error {
	reses, err := h.keeper.GetReservesContributors(ctx)
	if nil != err {
		ctx.Logger().Error("fail to get reserve contributors", err)
		return err
	}

	reses = reses.Add(msg.Contributor)
	if err := h.keeper.SetReserveContributors(ctx, reses); nil != err {
		ctx.Logger().Error("fail to save reserve contributors", err)
		return err
	}

	vault, err := h.keeper.GetVaultData(ctx)
	if nil != err {
		ctx.Logger().Error("fail to get vault data", err)
		return err
	}

	vault.TotalReserve = vault.TotalReserve.Add(msg.Contributor.Amount)
	if err := h.keeper.SetVaultData(ctx, vault); nil != err {
		ctx.Logger().Error("fail to save vault data", err)
		return err
	}

	return nil
}
