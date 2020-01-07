package thorchain

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
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
	withdrawBasisPoints := msg.WithdrawBasisPoints
	if withdrawBasisPoints.GT(sdk.ZeroUint()) && withdrawBasisPoints.GT(sdk.NewUint(MaxWithdrawBasisPoints)) {
		return fmt.Errorf("withdraw basis points %s is invalid", msg.WithdrawBasisPoints)
	}
	if !keeper.PoolExist(ctx, msg.Asset) {
		// pool doesn't exist
		return fmt.Errorf("pool-%s doesn't exist", msg.Asset)
	}
	return nil
}

// unstake withdraw all the asset
func unstake(ctx sdk.Context, keeper Keeper, msg MsgSetUnStake) (sdk.Uint, sdk.Uint, sdk.Uint, sdk.Error) {
	if err := validateUnstake(ctx, keeper, msg); nil != err {
		ctx.Logger().Error("msg unstake fail validation", err)
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.NewError(DefaultCodespace, CodeUnstakeFailValidation, err.Error())
	}

	pool, err := keeper.GetPool(ctx, msg.Asset)
	if err != nil {
		ctx.Logger().Error("fail to get pool", err)
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.ErrInternal("fail to get pool")
	}

	poolStaker, err := keeper.GetPoolStaker(ctx, msg.Asset)
	if nil != err {
		ctx.Logger().Error("can't find pool staker", err)
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.NewError(DefaultCodespace, CodePoolStakerNotExist, "pool staker doesn't exist")

	}
	stakerPool, err := keeper.GetStakerPool(ctx, msg.RuneAddress)
	if nil != err {
		ctx.Logger().Error("can't find staker pool", err)
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.NewError(DefaultCodespace, CodeStakerPoolNotExist, "staker pool doesn't exist")
	}

	poolUnits := pool.PoolUnits
	poolRune := pool.BalanceRune
	poolAsset := pool.BalanceAsset
	stakerUnit := poolStaker.GetStakerUnit(msg.RuneAddress)
	fStakerUnit := stakerUnit.Units
	if !stakerUnit.Units.GT(sdk.ZeroUint()) {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.NewError(DefaultCodespace, CodeNoStakeUnitLeft, "nothing to withdraw")
	}

	// check if thorchain need to rate limit unstaking
	// https://gitlab.com/thorchain/thornode/issues/166
	if !msg.Asset.Chain.Equals(common.BNBChain) {
		height := ctx.BlockHeight()
		if height < (stakerUnit.Height + 17280) {

			return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.NewError(DefaultCodespace, CodeWithdrawWithin24Hours, "you cannot unstake for 24 hours after staking for this blockchain")
		}
	}

	ctx.Logger().Info("pool before unstake", "pool unit", poolUnits, "balance RUNE", poolRune, "balance asset", poolAsset)
	ctx.Logger().Info("staker before withdraw", "staker unit", fStakerUnit)
	withdrawRune, withDrawAsset, unitAfter, err := calculateUnstake(poolUnits, poolRune, poolAsset, fStakerUnit, msg.WithdrawBasisPoints)
	if err != nil {
		ctx.Logger().Error("fail to unstake", err)
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.NewError(DefaultCodespace, CodeUnstakeFail, err.Error())
	}

	withdrawRune = withdrawRune.Add(stakerUnit.PendingRune) // extract pending rune
	stakerUnit.PendingRune = sdk.ZeroUint()                 // reset pending to zero

	ctx.Logger().Info("client withdraw", "RUNE", withdrawRune, "asset", withDrawAsset, "units left", unitAfter)
	// update pool
	pool.PoolUnits = common.SafeSub(poolUnits, fStakerUnit).Add(unitAfter)
	pool.BalanceRune = common.SafeSub(poolRune, withdrawRune)
	pool.BalanceAsset = common.SafeSub(poolAsset, withDrawAsset)
	ctx.Logger().Info("pool after unstake", "pool unit", pool.PoolUnits, "balance RUNE", pool.BalanceRune, "balance asset", pool.BalanceAsset)
	// update pool staker
	poolStaker.TotalUnits = pool.PoolUnits
	if unitAfter.IsZero() {
		// just remove it
		poolStaker.RemoveStakerUnit(msg.RuneAddress)
	} else {
		stakerUnit.Units = unitAfter
		poolStaker.UpsertStakerUnit(stakerUnit)
	}
	if unitAfter.IsZero() {
		stakerPool.RemoveStakerPoolItem(msg.Asset)
	} else {
		spi := stakerPool.GetStakerPoolItem(msg.Asset)
		spi.Units = unitAfter
		stakerPool.UpsertStakerPoolItem(spi)
	}

	// Create a pool event if THORNode have no rune or assets
	if pool.BalanceAsset.IsZero() || pool.BalanceRune.IsZero() {
		pool.Status = PoolBootstrap
	}

	// update staker pool
	if err := keeper.SetPool(ctx, pool); err != nil {
		ctx.Logger().Error("fail to save pool", err)
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), sdk.ErrInternal("fail to save pool")
	}
	keeper.SetPoolStaker(ctx, poolStaker)
	keeper.SetStakerPool(ctx, stakerPool)
	return withdrawRune, withDrawAsset, common.SafeSub(fStakerUnit, unitAfter), nil
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
	if withdrawBasisPoints.GT(sdk.NewUint(MaxWithdrawBasisPoints)) {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), fmt.Errorf("withdraw basis point %s is not valid", withdrawBasisPoints.String())
	}

	unitsToClaim := common.GetShare(withdrawBasisPoints, sdk.NewUint(10000), stakerUnits)
	withdrawRune := common.GetShare(unitsToClaim, poolUnits, poolRune)
	withdrawAsset := common.GetShare(unitsToClaim, poolUnits, poolAsset)
	unitAfter := common.SafeSub(stakerUnits, unitsToClaim)
	return withdrawRune, withdrawAsset, unitAfter, nil
}
