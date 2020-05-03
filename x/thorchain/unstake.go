package thorchain

import (
	"errors"
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

func validateUnstake(ctx sdk.Context, keeper Keeper, msg MsgSetUnStake) error {
	if msg.RuneAddress.IsEmpty() {
		return errors.New("empty rune address")
	}
	if msg.Tx.ID.IsEmpty() {
		return errors.New("request tx hash is empty")
	}
	if msg.Asset.IsEmpty() {
		return errors.New("empty asset")
	}
	withdrawBasisPoints := msg.UnstakeBasisPoints
	if !withdrawBasisPoints.GTE(sdk.ZeroUint()) || withdrawBasisPoints.GT(sdk.NewUint(MaxUnstakeBasisPoints)) {
		return fmt.Errorf("withdraw basis points %s is invalid", msg.UnstakeBasisPoints)
	}
	if !keeper.PoolExist(ctx, msg.Asset) {
		// pool doesn't exist
		return fmt.Errorf("pool-%s doesn't exist", msg.Asset)
	}
	return nil
}

// unstake withdraw all the asset
// it returns runeAmt,assetAmount,units, lastUnstake,err
func unstake(ctx sdk.Context, version semver.Version, keeper Keeper, msg MsgSetUnStake, eventManager EventManager) (sdk.Uint, sdk.Uint, sdk.Uint, sdk.Uint, sdk.Error) {
	if err := validateUnstake(ctx, keeper, msg); err != nil {
		ctx.Logger().Error("msg unstake fail validation", "error", err)
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.NewError(DefaultCodespace, CodeUnstakeFailValidation, err.Error())
	}

	pool, err := keeper.GetPool(ctx, msg.Asset)
	if err != nil {
		ctx.Logger().Error("fail to get pool", "error", err)
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.ErrInternal("fail to get pool")
	}

	stakerUnit, err := keeper.GetStaker(ctx, msg.Asset, msg.RuneAddress)
	if err != nil {
		ctx.Logger().Error("can't find staker", "error", err)
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.NewError(DefaultCodespace, CodeStakerNotExist, "staker doesn't exist")

	}

	poolUnits := pool.PoolUnits
	poolRune := pool.BalanceRune
	poolAsset := pool.BalanceAsset
	fStakerUnit := stakerUnit.Units
	if stakerUnit.Units.IsZero() || msg.UnstakeBasisPoints.IsZero() {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.NewError(DefaultCodespace, CodeNoStakeUnitLeft, "nothing to withdraw")
	}

	cv := constants.GetConstantValues(version)
	// check if thorchain need to rate limit unstaking
	// https://gitlab.com/thorchain/thornode/issues/166
	if !msg.Asset.Chain.Equals(common.BNBChain) {
		height := ctx.BlockHeight()
		if height < (stakerUnit.LastStakeHeight + cv.GetInt64Value(constants.StakeLockUpBlocks)) {
			return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.NewError(DefaultCodespace, CodeUnstakeWithin24Hours, "you cannot unstake for 24 hours after staking for this blockchain")
		}
	}

	ctx.Logger().Info("pool before unstake", "pool unit", poolUnits, "balance RUNE", poolRune, "balance asset", poolAsset)
	ctx.Logger().Info("staker before withdraw", "staker unit", fStakerUnit)
	withdrawRune, withDrawAsset, unitAfter, err := calculateUnstake(poolUnits, poolRune, poolAsset, fStakerUnit, msg.UnstakeBasisPoints)
	if err != nil {
		ctx.Logger().Error("fail to unstake", "error", err)
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.NewError(DefaultCodespace, CodeUnstakeFail, err.Error())
	}
	gasAsset := sdk.ZeroUint()
	// If the pool is empty, and there is a gas asset, subtract required gas
	if common.SafeSub(poolUnits, fStakerUnit).Add(unitAfter).IsZero() {
		// minus gas costs for our transactions
		if pool.Asset.IsBNB() {
			gasInfo, err := keeper.GetGas(ctx, pool.Asset)
			if err != nil {
				ctx.Logger().Error("fail to get gas for asset", "asset", pool.Asset, "error", err)
				return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.NewError(DefaultCodespace, CodeUnstakeFail, err.Error())
			}
			originalAsset := withDrawAsset
			withDrawAsset = common.SafeSub(
				withDrawAsset,
				gasInfo[0].MulUint64(uint64(2)),
			)
			gasAsset = originalAsset.Sub(withDrawAsset)
		} else if pool.Asset.Chain.GetGasAsset().Equals(pool.Asset) {
			// leave half a RUNE as gas fee for BTC chain and ETH chain
			transactionFee := cv.GetInt64Value(constants.TransactionFee)
			gasAsset = pool.RuneValueInAsset(sdk.NewUint(uint64(transactionFee / 2)))
			withDrawAsset = common.SafeSub(withDrawAsset, gasAsset)
		}
	}

	withdrawRune = withdrawRune.Add(stakerUnit.PendingRune) // extract pending rune
	stakerUnit.PendingRune = sdk.ZeroUint()                 // reset pending to zero

	ctx.Logger().Info("client withdraw", "RUNE", withdrawRune, "asset", withDrawAsset, "units left", unitAfter)
	// update pool
	pool.PoolUnits = common.SafeSub(poolUnits, fStakerUnit).Add(unitAfter)
	pool.BalanceRune = common.SafeSub(poolRune, withdrawRune)
	pool.BalanceAsset = common.SafeSub(poolAsset, withDrawAsset)

	ctx.Logger().Info("pool after unstake", "pool unit", pool.PoolUnits, "balance RUNE", pool.BalanceRune, "balance asset", pool.BalanceAsset)
	// update staker
	stakerUnit.Units = unitAfter
	stakerUnit.LastUnStakeHeight = ctx.BlockHeight()

	// Create a pool event if THORNode have no rune or assets
	if pool.BalanceAsset.IsZero() || pool.BalanceRune.IsZero() {
		poolEvt := NewEventPool(pool.Asset, PoolBootstrap)
		if err := eventManager.EmitPoolEvent(ctx, keeper, common.BlankTxID, EventSuccess, poolEvt); nil != err {
			ctx.Logger().Error("fail to emit pool event", "error", err)
		}
		pool.Status = PoolBootstrap
	}

	if err := keeper.SetPool(ctx, pool); err != nil {
		ctx.Logger().Error("fail to save pool", "error", err)
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.ErrInternal("fail to save pool")
	}
	if !stakerUnit.Units.IsZero() {
		keeper.SetStaker(ctx, stakerUnit)
	} else {
		keeper.RemoveStaker(ctx, stakerUnit)
	}
	return withdrawRune, withDrawAsset, common.SafeSub(fStakerUnit, unitAfter), gasAsset, nil
}

func calculateUnstake(poolUnits, poolRune, poolAsset, stakerUnits, withdrawBasisPoints sdk.Uint) (sdk.Uint, sdk.Uint, sdk.Uint, error) {
	if poolUnits.IsZero() {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.New("poolUnits can't be zero")
	}
	if poolRune.IsZero() {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.New("pool rune balance can't be zero")
	}
	if poolAsset.IsZero() {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.New("pool asset balance can't be zero")
	}
	if stakerUnits.IsZero() {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.New("staker unit can't be zero")
	}
	if withdrawBasisPoints.GT(sdk.NewUint(MaxUnstakeBasisPoints)) {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), fmt.Errorf("withdraw basis point %s is not valid", withdrawBasisPoints.String())
	}

	unitsToClaim := common.GetShare(withdrawBasisPoints, sdk.NewUint(10000), stakerUnits)
	withdrawRune := common.GetShare(unitsToClaim, poolUnits, poolRune)
	withdrawAsset := common.GetShare(unitsToClaim, poolUnits, poolAsset)
	unitAfter := common.SafeSub(stakerUnits, unitsToClaim)
	return withdrawRune, withdrawAsset, unitAfter, nil
}
