package swapservice

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

func validateUnstake(ctx sdk.Context, keeper poolStorage, msg MsgSetUnStake) error {
	if msg.PublicAddress.IsEmpty() {
		return errors.New("empty public address")
	}
	if msg.RequestTxHash.IsEmpty() {
		return errors.New("request tx hash is empty")
	}
	if msg.Ticker.IsEmpty() {
		return errors.New("empty ticker")
	}
	withdrawBasisPoints := msg.WithdrawBasisPoints
	if withdrawBasisPoints.GT(sdk.ZeroUint()) && withdrawBasisPoints.GT(sdk.NewUint(MaxWithdrawBasisPoints)) {
		return errors.Errorf("withdraw basis points %s is invalid", msg.WithdrawBasisPoints)
	}
	if !keeper.PoolExist(ctx, msg.Ticker) {
		// pool doesn't exist
		return errors.Errorf("pool-%s doesn't exist", msg.Ticker)
	}
	return nil
}

// unstake withdraw all the asset
func unstake(ctx sdk.Context, keeper poolStorage, msg MsgSetUnStake) (sdk.Uint, sdk.Uint, sdk.Uint, error) {
	if err := validateUnstake(ctx, keeper, msg); nil != err {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), err
	}

	// here fBalance should be valid , because we did the validation above
	pool := keeper.GetPool(ctx, msg.Ticker)
	poolStaker, err := keeper.GetPoolStaker(ctx, msg.Ticker)
	if nil != err {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.Wrap(err, "can't find pool staker")

	}
	stakerPool, err := keeper.GetStakerPool(ctx, msg.PublicAddress)
	if nil != err {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.Wrap(err, "can't find staker pool")
	}

	poolUnits := pool.PoolUnits
	poolRune := pool.BalanceRune
	poolToken := pool.BalanceToken
	stakerUnit := poolStaker.GetStakerUnit(msg.PublicAddress)
	fStakerUnit := stakerUnit.Units
	if !stakerUnit.Units.GT(sdk.ZeroUint()) {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.New("nothing to withdraw")
	}

	ctx.Logger().Info("pool before unstake", "pool unit", poolUnits, "balance RUNE", poolRune, "balance token", poolToken)
	ctx.Logger().Info("staker before withdraw", "staker unit", fStakerUnit)
	withdrawRune, withDrawToken, unitAfter, err := calculateUnstake(poolUnits, poolRune, poolToken, fStakerUnit, msg.WithdrawBasisPoints)
	if err != nil {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), err
	}
	ctx.Logger().Info("client withdraw", "RUNE", withdrawRune, "token", withDrawToken, "units left", unitAfter)
	// update pool
	pool.PoolUnits = poolUnits.Sub(fStakerUnit).Add(unitAfter)
	pool.BalanceRune = poolRune.Sub(withdrawRune)
	pool.BalanceToken = poolToken.Sub(withDrawToken)
	ctx.Logger().Info("pool after unstake", "pool unit", pool.PoolUnits, "balance RUNE", pool.BalanceRune, "balance token", pool.BalanceToken)
	// update pool staker
	poolStaker.TotalUnits = pool.PoolUnits
	if unitAfter.IsZero() {
		// just remove it
		poolStaker.RemoveStakerUnit(msg.PublicAddress)
	} else {
		stakerUnit.Units = unitAfter
		poolStaker.UpsertStakerUnit(stakerUnit)
	}
	if unitAfter.IsZero() {
		stakerPool.RemoveStakerPoolItem(msg.Ticker)
	} else {
		spi := stakerPool.GetStakerPoolItem(msg.Ticker)
		spi.Units = unitAfter
		stakerPool.UpsertStakerPoolItem(spi)
	}
	// update staker pool
	keeper.SetPool(ctx, pool)
	keeper.SetPoolStaker(ctx, msg.Ticker, poolStaker)
	keeper.SetStakerPool(ctx, msg.PublicAddress, stakerPool)
	return withdrawRune, withDrawToken, fStakerUnit.Sub(unitAfter), nil
}

func calculateUnstake(poolUnit, poolRune, poolToken, stakerUnit, withdrawBasisPoints sdk.Uint) (sdk.Uint, sdk.Uint, sdk.Uint, error) {
	if poolUnit.IsZero() {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.New("poolUnits can't be zero")
	}
	if poolRune.IsZero() {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.New("pool rune balance can't be zero")
	}
	if poolToken.IsZero() {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.New("pool token balance can't be zero")
	}
	if stakerUnit.IsZero() {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.New("staker unit can't be zero")
	}
	if withdrawBasisPoints.GT(sdk.NewUint(MaxWithdrawBasisPoints)) {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.Errorf("withdraw basis point %s is not valid", withdrawBasisPoints.String())
	}
	percentage := uintToFloat64(withdrawBasisPoints) / 100
	stakerOwnership := uintToFloat64(stakerUnit) / uintToFloat64(poolUnit)

	//withdrawRune := stakerOwnership.Mul(withdrawBasisPoints).Quo(sdk.NewUint(10000)).Mul(poolRune)
	//withdrawToken := stakerOwnership.Mul(withdrawBasisPoints).Quo(sdk.NewUint(10000)).Mul(poolToken)
	//unitAfter := stakerUnit.Mul(sdk.NewUint(MaxWithdrawBasisPoints).Sub(withdrawBasisPoints).Quo(sdk.NewUint(10000)))
	withdrawRune := stakerOwnership * percentage / 100 * uintToFloat64(poolRune)
	withdrawToken := stakerOwnership * percentage / 100 * uintToFloat64(poolToken)
	unitAfter := uintToFloat64(stakerUnit) * (100 - percentage) / 100
	return floatToUint(withdrawRune), floatToUint(withdrawToken), floatToUint(unitAfter), nil
}
