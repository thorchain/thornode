package thorchain

import (
	"encoding/json"
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// StakeHandler is to handle stake
type StakeHandler struct {
	keeper Keeper
}

// NewStakeHandler create a new instance of StakeHandler
func NewStakeHandler(keeper Keeper) StakeHandler {
	return StakeHandler{keeper: keeper}
}

func (sh StakeHandler) validate(ctx sdk.Context, msg MsgSetStakeData, version semver.Version) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return sh.validateV1(ctx, msg)
	}
	return errBadVersion
}

func (sh StakeHandler) validateV1(ctx sdk.Context, msg MsgSetStakeData) sdk.Error {
	if err := msg.ValidateBasic(); nil != err {
		return err
	}
	if !isSignedByActiveNodeAccounts(ctx, sh.keeper, msg.GetSigners()) {
		return sdk.ErrUnauthorized("msg is not signed by an active node account")
	}

	return nil
}

// Run execute the handler
func (sh StakeHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version, _ constants.ConstantValues) sdk.Result {
	msg, ok := m.(MsgSetStakeData)
	if !ok {
		return errInvalidMessage.Result()
	}
	ctx.Logger().Info("received stake request",
		"asset", msg.Asset.String(),
		"tx", msg.Tx)
	if err := sh.validate(ctx, msg, version); nil != err {
		ctx.Logger().Error("msg stake fail validation", err)
		return err.Result()
	}

	if err := sh.handle(ctx, msg, version); nil != err {
		ctx.Logger().Error("fail to process msg stake", err)
		return err.Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func (sh StakeHandler) handle(ctx sdk.Context, msg MsgSetStakeData, version semver.Version) (errResult sdk.Error) {
	stakeUnits := sdk.ZeroUint()
	defer func() {
		// if fail to stake ,then it just need to return, tx in observer will refund
		if errResult != nil && !errResult.Code().IsOK() {
			ctx.Logger().Error("No stake event due to ", errResult.Error())
			return
		}
		if err := processStakeEvent(ctx, sh.keeper, msg, stakeUnits, EventSuccess); nil != err {
			errResult = sdk.ErrInternal(fmt.Errorf("fail to save stake event: %w", err).Error())
		}
	}()

	pool, err := sh.keeper.GetPool(ctx, msg.Asset)
	if err != nil {
		return sdk.ErrInternal(fmt.Errorf("fail to get pool: %w", err).Error())
	}

	if pool.Empty() {
		ctx.Logger().Info("pool doesn't exist yet, create a new one", "symbol", msg.Asset.String(), "creator", msg.RuneAddress)
		pool.Asset = msg.Asset
		if err := sh.keeper.SetPool(ctx, pool); err != nil {
			return sdk.ErrInternal(fmt.Errorf("fail to save pool to key value store: %w", err).Error())
		}
	}
	if err := pool.EnsureValidPoolStatus(msg); nil != err {
		return sdk.ErrUnknownRequest(fmt.Errorf("fail to check pool status: %w", err).Error())
	}
	stakeUnits, err = stake(
		ctx,
		sh.keeper,
		msg.Asset,
		msg.RuneAmount,
		msg.AssetAmount,
		msg.RuneAddress,
		msg.AssetAddress,
		msg.Tx.ID,
	)
	if err != nil {
		return sdk.ErrUnknownRequest(fmt.Errorf("fail to process stake request: %w", err).Error())
	}
	return nil
}

func processStakeEvent(ctx sdk.Context, keeper Keeper, msg MsgSetStakeData, stakeUnits sdk.Uint, eventStatus EventStatus) error {
	var stakeEvt EventStake
	if eventStatus == EventFail {
		// do not log event if the stake failed
		return nil
	}

	stakeEvt = NewEventStake(
		msg.Asset,
		stakeUnits,
	)
	stakeBytes, err := json.Marshal(stakeEvt)
	if err != nil {
		return fmt.Errorf("fail to marshal stake event to json: %w", err)
	}
	evt := NewEvent(
		stakeEvt.Type(),
		ctx.BlockHeight(),
		msg.Tx,
		stakeBytes,
		eventStatus,
	)
	tx := common.Tx{ID: common.BlankTxID}
	evt.OutTxs = common.Txs{tx}
	return keeper.UpsertEvent(ctx, evt)
}
