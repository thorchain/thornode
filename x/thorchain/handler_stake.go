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

func (h StakeHandler) validate(ctx sdk.Context, msg MsgSetStakeData, version semver.Version) sdk.Error {
	if version.GTE(semver.MustParse("0.1.0")) {
		return h.validateV1(ctx, msg)
	}
	return errBadVersion
}

func (h StakeHandler) validateV1(ctx sdk.Context, msg MsgSetStakeData) sdk.Error {
	if err := msg.ValidateBasic(); nil != err {
		return err
	}
	if !isSignedByActiveNodeAccounts(ctx, h.keeper, msg.GetSigners()) {
		return sdk.ErrUnauthorized("msg is not signed by an active node account")
	}
	totalRune, err := h.getTotalStakeRUNE(ctx)
	if nil != err {
		return sdk.ErrInternal(fmt.Errorf("fail to get total stake rune: %w", err).Error())
	}
	totalRune = totalRune.Add(msg.RuneAmount)
	if totalRune.GT(sdk.NewUint(constants.MaximumStakeRune * common.One)) {
		return errStakeRuneOverLimit
	}
	totalBond, err := h.getTotalBond(ctx)
	if nil != err {
		return sdk.ErrInternal(fmt.Errorf("fail to get total bond: %w", err).Error())
	}
	if totalBond.GT(sdk.ZeroUint()) && totalRune.GT(totalBond) {
		return sdk.NewError(DefaultCodespace, CodeStakeRUNEMoreThanBond, "total stake RUNE (%s) is more than bond (%s)", totalRune, totalBond)
	}

	return nil
}

// Run execute the handler
func (h StakeHandler) Run(ctx sdk.Context, m sdk.Msg, version semver.Version) sdk.Result {
	msg, ok := m.(MsgSetStakeData)
	if !ok {
		return errInvalidMessage.Result()
	}
	ctx.Logger().Info("received stake request",
		"asset", msg.Asset.String(),
		"tx", msg.Tx)
	if err := h.validate(ctx, msg, version); nil != err {
		ctx.Logger().Error("msg stake fail validation", err)
		return err.Result()
	}

	if err := h.handle(ctx, msg, version); nil != err {
		ctx.Logger().Error("fail to process msg stake", err)
		return err.Result()
	}

	return sdk.Result{
		Code:      sdk.CodeOK,
		Codespace: DefaultCodespace,
	}
}

func (h StakeHandler) handle(ctx sdk.Context, msg MsgSetStakeData, version semver.Version) (errResult sdk.Error) {
	stakeUnits := sdk.ZeroUint()
	defer func() {
		var status EventStatus
		if errResult == nil {
			status = EventSuccess
		} else {
			status = EventRefund
		}
		if err := processStakeEvent(ctx, h.keeper, msg, stakeUnits, status); nil != err {
			errResult = sdk.ErrInternal(fmt.Errorf("fail to save stake event: %w", err).Error())
		}
	}()

	pool, err := h.keeper.GetPool(ctx, msg.Asset)
	if err != nil {
		return sdk.ErrInternal(fmt.Errorf("fail to get pool: %w", err).Error())
	}

	if pool.Empty() {
		ctx.Logger().Info("pool doesn't exist yet, create a new one", "symbol", msg.Asset.String(), "creator", msg.RuneAddress)
		pool.Asset = msg.Asset
		if err := h.keeper.SetPool(ctx, pool); err != nil {
			return sdk.ErrInternal(fmt.Errorf("fail to save pool to key value store: %w", err).Error())
		}
	}
	if err := pool.EnsureValidPoolStatus(msg); nil != err {
		return sdk.ErrUnknownRequest(fmt.Errorf("fail to check pool status: %w", err).Error())
	}
	stakeUnits, err = stake(
		ctx,
		h.keeper,
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
	if eventStatus == EventRefund {
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

	if err := keeper.AddIncompleteEvents(ctx, evt); err != nil {
		return err
	}

	if eventStatus != EventRefund {
		// since there is no outbound tx for staking, we'll complete the event now
		tx := common.Tx{ID: common.BlankTxID}
		err := completeEvents(ctx, keeper, msg.Tx.ID, common.Txs{tx})
		if err != nil {
			return fmt.Errorf("unable to complete events: %w", err)
		}
	}
	return nil
}

// getTotalBond
func (h StakeHandler) getTotalBond(ctx sdk.Context) (sdk.Uint, error) {
	nodeAccounts, err := h.keeper.ListNodeAccounts(ctx)
	if nil != err {
		return sdk.ZeroUint(), nil
	}
	total := sdk.ZeroUint()
	for _, na := range nodeAccounts {
		if na.Status == NodeDisabled {
			continue
		}
		total = total.Add(na.Bond)
	}
	return total, nil
}

// getTotalStakeRUNE we have in all pools
func (h StakeHandler) getTotalStakeRUNE(ctx sdk.Context) (sdk.Uint, error) {
	pools, err := h.keeper.GetPools(ctx)
	if nil != err {
		return sdk.ZeroUint(), fmt.Errorf("fail to get pools from data store: %w", err)
	}
	total := sdk.ZeroUint()
	for _, p := range pools {
		total = total.Add(p.BalanceRune)
	}
	return total, nil
}
