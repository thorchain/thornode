package swapservice

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

const floatPrecision = 8

// validateStakeMessage is to do some validation , and make sure it is legit
func validateStakeMessage(ctx sdk.Context, keeper Keeper, name, ticker, rune_amount, token_amount string) error {
	if isEmptyString(name) {
		return errors.New("name is empty")
	}
	if isEmptyString(ticker) {
		return errors.New("ticker is empty")
	}
	if isEmptyString(rune_amount) {
		return errors.New("rune_amount is empty")
	}
	if isEmptyString(token_amount) {
		return errors.New("token_amount is empty")
	}
	poolID := types.GetPoolNameFromTicker(ticker)
	if !keeper.PoolExist(ctx, poolID) {
		return errors.Errorf("%s doesn't exist", poolID)
	}
	return nil
}

func stake(ctx sdk.Context, keeper Keeper, name, ticker, stakeRuneAmount, stakeTokenAmount, publicAddress string) error {
	ctx.Logger().Info(fmt.Sprintf("%s staking %s %s %s", name, ticker, stakeRuneAmount, stakeTokenAmount))
	if err := validateStakeMessage(ctx, keeper, name, ticker, stakeRuneAmount, stakeTokenAmount); nil != err {
		ctx.Logger().Error("invalid request", err)
		return errors.Wrap(err, "invalid request")
	}
	ticker = strings.ToUpper(ticker)
	poolID := types.GetPoolNameFromTicker(ticker)
	pool := keeper.GetPoolStruct(ctx, poolID)
	fTokenAmt, err := strconv.ParseFloat(stakeTokenAmount, 64)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("%s is invalid token_amount", stakeTokenAmount), err)
		return errors.Wrapf(err, "%s is invalid token_amount", stakeTokenAmount)
	}
	fRuneAmt, err := strconv.ParseFloat(stakeRuneAmount, 64)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("%s is invalid rune_amount", stakeRuneAmount), err)
		return errors.Wrapf(err, "%s is invalid rune_amount", stakeRuneAmount)
	}
	ctx.Logger().Info(fmt.Sprintf("Pre-Pool: %sRUNE %sToken", pool.BalanceRune, pool.BalanceToken))
	ctx.Logger().Info(fmt.Sprintf("Staking: %sRUNE %sToken", stakeRuneAmount, stakeTokenAmount))

	balanceRune, err := strconv.ParseFloat(pool.BalanceRune, 64)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("%s is invalid pool rune balance", pool.BalanceRune))
		return errors.Wrapf(err, "%s is invalid pool rune balance", pool.BalanceRune)
	}

	balanceToken, err := strconv.ParseFloat(pool.BalanceToken, 64)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("%s is invalid pool token balance", pool.BalanceToken))
		return errors.Wrapf(err, "%s is invalid pool token balance", pool.BalanceRune)
	}
	oldPoolUnits, err := strconv.ParseFloat(pool.PoolUnits, 64)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("%s is invalid pool total units", pool.PoolUnits))
		return errors.Wrapf(err, "%s is invalid pool total units", pool.PoolUnits)
	}
	newPoolUnits, stakerUnits, err := calculatePoolUnits(oldPoolUnits, balanceRune, balanceToken, fRuneAmt, fTokenAmt)
	if nil != err {
		ctx.Logger().Error("fail to calculate poolUnits", err)
		return errors.Wrapf(err, "fail to calculate pool units")
	}
	ctx.Logger().Info(fmt.Sprintf("current pool units : %f ,staker units : %f", newPoolUnits, stakerUnits))
	poolRune := balanceRune + fRuneAmt
	poolToken := balanceToken + fTokenAmt
	pool.PoolUnits = strconv.FormatFloat(newPoolUnits, 'f', floatPrecision, 64)
	pool.BalanceRune = strconv.FormatFloat(poolRune, 'f', floatPrecision, 64)
	pool.BalanceToken = strconv.FormatFloat(poolToken, 'f', floatPrecision, 64)
	ctx.Logger().Info(fmt.Sprintf("Post-Pool: %sRUNE %sToken", pool.BalanceRune, pool.BalanceToken))
	keeper.SetPoolStruct(ctx, poolID, pool)
	ps, err := keeper.GetPoolStaker(ctx, poolID)
	if nil != err {
		ctx.Logger().Error("fail to get pool staker", "err:", err)
		return errors.Wrap(err, "fail to get pool staker..")
	}
	ps.TotalUnits = pool.PoolUnits
	su := ps.GetStakerUnit(publicAddress)

	fex, err := strconv.ParseFloat(su.Units, 64)
	if nil != err {
		ctx.Logger().Error("fail to parse staker's exist stake unit", "err", err)
	}
	stakerUnits += fex
	su.Units = strconv.FormatFloat(stakerUnits, 'f', floatPrecision, 64)
	ps.UpsertStakerUnit(su)
	keeper.SetPoolStaker(ctx, poolID, ps)
	sp, err := keeper.GetStakerPool(ctx, publicAddress)
	if nil != err {
		ctx.Logger().Error("fail to get stakerpool object", err)
		return errors.Wrap(err, "fail to get stakepool object")
	}
	stakerPoolItem := sp.GetStakerPoolItem(poolID)
	existUnit, err := strconv.ParseFloat(stakerPoolItem.Units, 64)
	if nil != err {
		ctx.Logger().Error("fail to parse exist unit", err)
		return errors.Wrap(err, "fail to parse exist unit")
	}
	existRune, err := strconv.ParseFloat(stakerPoolItem.RuneBalance, 64)
	if nil != err {
		ctx.Logger().Error("fail to parse exist RUNE", err)
		return errors.Wrap(err, "fail to parse exist RUNE")
	}
	existToken, err := strconv.ParseFloat(stakerPoolItem.TokenBalance, 64)
	if nil != err {
		ctx.Logger().Error("fail to parse exist token", err)
		return errors.Wrap(err, "fail to parse exist token")
	}
	stakerUnits += existUnit
	fRuneAmt += existRune
	fTokenAmt += existToken

	stakerPoolItem.Units = strconv.FormatFloat(stakerUnits, 'f', floatPrecision, 64)
	stakerPoolItem.RuneBalance = strconv.FormatFloat(fRuneAmt, 'f', floatPrecision, 64)
	stakerPoolItem.TokenBalance = strconv.FormatFloat(fTokenAmt, 'f', floatPrecision, 64)
	sp.UpsertStakerPoolItem(stakerPoolItem)
	keeper.SetStakerPool(ctx, publicAddress, sp)

	return nil
}

// calculatePoolUnits calculate the pool units and staker units
func calculatePoolUnits(oldPoolUnits, poolRune, poolToken, stakeRune, stakeToken float64) (float64, float64, error) {
	if poolRune < 0 {
		return 0, 0, errors.New("negative RUNE in the pool,likely it is corrupted")
	}
	if poolToken < 0 {
		return 0, 0, errors.New("negative token in the pool,likely it is corrupted")
	}
	if stakeRune < 0 {
		return 0, 0, errors.New("you can't stake negative rune")
	}
	if stakeToken < 0 {
		return 0, 0, errors.New("you can't stake negative token")
	}
	stakerPercentage := ((stakeRune / (stakeRune + poolRune)) + (stakeToken / (stakeToken + poolToken))) / 2
	stakerUnit := (stakerPercentage*(stakeRune+poolRune) + stakerPercentage*(stakeToken+poolToken)) / 2
	newPoolUnit := oldPoolUnits + stakerUnit
	return newPoolUnit, stakerUnit, nil
}
