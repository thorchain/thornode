package swapservice

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

const unstakeRecordPrefix = `unstakerecord-`

func validateUnstake(ctx sdk.Context, keeper Keeper, msg types.MsgSetUnStake) error {
	if isEmptyString(msg.Name) {
		return errors.New("empty name")
	}
	if isEmptyString(msg.PublicAddress) {
		return errors.New("empty public address")
	}
	if isEmptyString(msg.Percentage) {
		return errors.New("empty percentage")
	}
	if isEmptyString(msg.RequestTxHash) {
		return errors.New("request tx hash is empty")
	}
	if isEmptyString(msg.Ticker) {
		return errors.New("empty ticker")
	}
	fPercentage, err := strconv.ParseFloat(msg.Percentage, 64)
	if nil != err {
		return errors.Wrapf(err, "percentage %s is invalid ", msg.Percentage)
	}
	if fPercentage <= 0 || fPercentage > 100 {
		return errors.Errorf("percentage %s is invalid", msg.Percentage)
	}
	if !keeper.PoolExist(ctx, msg.Ticker) {
		// pool doesn't exist
		return errors.Errorf("pool-%s doesn't exist", msg.Ticker)
	}
	return nil
}

// unstake withdraw all the asset
func unstake(ctx sdk.Context, keeper Keeper, msg types.MsgSetUnStake) (string, string, error) {
	if err := validateUnstake(ctx, keeper, msg); nil != err {
		return "0", "0", err
	}
	fPercentage, err := strconv.ParseFloat(msg.Percentage, 64)
	if nil != err {
		return "0", "0", errors.Wrapf(err, " %s is invalid percentage", msg.Percentage)
	}
	// here fBalance should be valid , because we did the validation above
	pool := keeper.GetPoolStruct(ctx, msg.Ticker)
	poolStaker, err := keeper.GetPoolStaker(ctx, msg.Ticker)
	if nil != err {
		return "0", "0", errors.Wrap(err, "can't find pool staker")

	}
	stakerPool, err := keeper.GetStakerPool(ctx, msg.PublicAddress)
	if nil != err {
		return "0", "0", errors.Wrap(err, "can't find staker pool")

	}
	poolUnits, err := strconv.ParseFloat(pool.PoolUnits, 64)
	if nil != err {
		return "0", "0", errors.Wrapf(err, "poolUnits :%s is not valid", pool.PoolUnits)
	}
	poolRune, err := strconv.ParseFloat(pool.BalanceRune, 64)
	if nil != err {
		return "0", "0", errors.Wrapf(err, "pool RUNE balance (%s) is not valid", pool.BalanceRune)
	}
	poolToken, err := strconv.ParseFloat(pool.BalanceToken, 64)
	if nil != err {
		return "0", "0", errors.Wrapf(err, "pool token balance (%s) is invalid", pool.BalanceToken)
	}
	stakerUnit := poolStaker.GetStakerUnit(msg.PublicAddress)
	fStakerUnit, err := strconv.ParseFloat(stakerUnit.Units, 64)
	if nil != err {
		return "0", "0", errors.Wrapf(err, "staker unit (%s) is invalid", stakerUnit.Units)
	}
	ctx.Logger().Info("pool before unstake", "pool unit", poolUnits, "balance RUNE", poolRune, "balance token", poolToken)
	ctx.Logger().Info("staker before withdraw", "staker unit", fStakerUnit)
	withdrawRune, withDrawToken, unitAfter, err := calculateUnstake(poolUnits, poolRune, poolToken, fStakerUnit, fPercentage)
	if err != nil {
		return "0", "0", err
	}
	ctx.Logger().Info("client withdraw", "RUNE", withdrawRune, "token", withDrawToken, "units left", unitAfter)
	// update pool
	pool.PoolUnits = float64ToString(poolUnits - fStakerUnit + unitAfter)
	pool.BalanceRune = float64ToString(poolRune - withdrawRune)
	pool.BalanceToken = float64ToString(poolToken - withDrawToken)
	ctx.Logger().Info("pool after unstake", "pool unit", pool.PoolUnits, "balance RUNE", pool.BalanceRune, "balance token", pool.BalanceToken)
	// update pool staker
	poolStaker.TotalUnits = pool.PoolUnits
	if unitAfter == 0 {
		// just remove it
		poolStaker.RemoveStakerUnit(msg.PublicAddress)
	} else {
		stakerUnit.Units = float64ToString(unitAfter)
		poolStaker.UpsertStakerUnit(stakerUnit)
	}
	if unitAfter <= 0 {
		stakerPool.RemoveStakerPoolItem(msg.Ticker)
	} else {
		spi := stakerPool.GetStakerPoolItem(msg.Ticker)
		spi.Units = float64ToString(unitAfter)
		stakerPool.UpsertStakerPoolItem(spi)
	}
	// update staker pool
	keeper.SetPoolStruct(ctx, msg.Ticker, pool)
	keeper.SetPoolStaker(ctx, msg.Ticker, poolStaker)
	keeper.SetStakerPool(ctx, msg.PublicAddress, stakerPool)
	keeper.SetUnStakeRecord(ctx, types.UnstakeRecord{
		RequestTxHash: msg.RequestTxHash,
		Ticker:        msg.Ticker,
		PublicAddress: msg.PublicAddress,
		Percentage:    msg.Percentage,
	})
	return float64ToString(withdrawRune), float64ToString(withDrawToken), nil
}
func float64ToString(fvalue float64) string {
	return strconv.FormatFloat(fvalue, 'f', floatPrecision, 64)
}

func calculateUnstake(poolUnit, poolRune, poolToken, stakerUnit, percentage float64) (float64, float64, float64, error) {
	if poolUnit <= 0 {
		return 0, 0, 0, errors.New("poolUnits can't be zero or negative")
	}
	if poolRune <= 0 {
		return 0, 0, 0, errors.New("pool rune balance can't be zero or negative")
	}
	if poolToken <= 0 {
		return 0, 0, 0, errors.New("pool token balance can't be zero or negative")
	}
	if stakerUnit < 0 {
		return 0, 0, 0, errors.New("staker unit can't be negative")
	}
	if percentage < 0 || percentage > 100 {
		return 0, 0, 0, errors.Errorf("percentage %f is not valid", percentage)
	}
	stakerOwnership := stakerUnit / poolUnit
	withdrawRune := stakerOwnership * percentage / 100 * poolRune
	withdrawToken := stakerOwnership * percentage / 100 * poolToken
	unitAfter := stakerUnit * (100 - percentage) / 100
	return withdrawRune, withdrawToken, unitAfter, nil
}

// unStakeComplete  mark a swap to be in complete state
func unStakeComplete(ctx sdk.Context, keeper poolStorage, requestTxHash, completeTxHash string) error {
	if isEmptyString(requestTxHash) {
		return errors.New("request tx hash is empty")
	}
	if isEmptyString(completeTxHash) {
		return errors.New("complete tx hash is empty")
	}
	return keeper.UpdateUnStakeRecordCompleteTxHash(ctx, requestTxHash, completeTxHash)
}
