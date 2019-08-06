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
const stakerLimit = 100.0 // TODO: make configurable

// validateStakeAmount
func validateStakeAmount(stakers PoolStaker, stakerUnits float64) error {
	var minStakerAmt float64
	stakerCount := float64(len(stakers.Stakers))
	if stakerCount <= stakerLimit {
		minStakerAmt = 0 // first 100 stakers there are no lower limits
	} else {
		totalUnits, err := strconv.ParseFloat(stakers.TotalUnits, 64)
		if err != nil {
			return errors.Wrapf(err, "%s is invalid units", stakers.TotalUnits)
		}
		avgStake := totalUnits / stakerCount

		minStakerAmt = avgStake * ((stakerCount / stakerLimit) + 0.1) // Increases minStakeAmt by 10% every 100 stakers
	}

	if stakerUnits < minStakerAmt {
		return fmt.Errorf("Not enough to stake (%f/%f)", stakerUnits, minStakerAmt)
	}

	return nil
}

// validateStakeMessage is to do some validation , and make sure it is legit
func validateStakeMessage(ctx sdk.Context, keeper Keeper, name, ticker, stakeRuneAmount, stakeTokenAmount, requestTxHash, publicAddress string) error {
	if isEmptyString(name) {
		return errors.New("name is empty")
	}
	if isEmptyString(ticker) {
		return errors.New("ticker is empty")
	}
	if isEmptyString(stakeRuneAmount) {
		return errors.New("stake rune amount is empty")
	}
	if isEmptyString(stakeTokenAmount) {
		return errors.New("stake token amount is empty")
	}
	if isEmptyString(requestTxHash) {
		return errors.New("request tx hash is empty")
	}
	if isEmptyString(publicAddress) {
		return errors.New("public address is empty")
	}
	poolID := types.GetPoolNameFromTicker(ticker)
	if !keeper.PoolExist(ctx, poolID) {
		return errors.Errorf("%s doesn't exist", poolID)
	}
	return nil
}

func stake(ctx sdk.Context, keeper Keeper, name, ticker, stakeRuneAmount, stakeTokenAmount, publicAddress, requestTxHash string) error {
	ctx.Logger().Info(fmt.Sprintf("%s staking %s %s %s", name, ticker, stakeRuneAmount, stakeTokenAmount))
	if err := validateStakeMessage(ctx, keeper, name, ticker, stakeRuneAmount, stakeTokenAmount, requestTxHash, publicAddress); nil != err {
		return errors.Wrap(err, "invalid request")
	}
	ticker = strings.ToUpper(ticker)
	poolID := types.GetPoolNameFromTicker(ticker)
	pool := keeper.GetPoolStruct(ctx, poolID)
	fTokenAmt, err := strconv.ParseFloat(stakeTokenAmount, 64)
	if err != nil {
		return errors.Wrapf(err, "%s is invalid token_amount", stakeTokenAmount)
	}
	fRuneAmt, err := strconv.ParseFloat(stakeRuneAmount, 64)
	if err != nil {
		return errors.Wrapf(err, "%s is invalid rune_amount", stakeRuneAmount)
	}
	ctx.Logger().Info(fmt.Sprintf("Pre-Pool: %sRUNE %sToken", pool.BalanceRune, pool.BalanceToken))
	ctx.Logger().Info(fmt.Sprintf("Staking: %sRUNE %sToken", stakeRuneAmount, stakeTokenAmount))

	balanceRune, err := strconv.ParseFloat(pool.BalanceRune, 64)
	if err != nil {
		return errors.Wrapf(err, "%s is invalid pool rune balance", pool.BalanceRune)
	}

	balanceToken, err := strconv.ParseFloat(pool.BalanceToken, 64)
	if err != nil {
		return errors.Wrapf(err, "%s is invalid pool token balance", pool.BalanceRune)
	}
	oldPoolUnits, err := strconv.ParseFloat(pool.PoolUnits, 64)
	if err != nil {
		return errors.Wrapf(err, "%s is invalid pool total units", pool.PoolUnits)
	}
	newPoolUnits, stakerUnits, err := calculatePoolUnits(oldPoolUnits, balanceRune, balanceToken, fRuneAmt, fTokenAmt)
	if nil != err {
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
	// maintain pool staker structure
	ps, err := keeper.GetPoolStaker(ctx, poolID)
	if nil != err {
		return errors.Wrap(err, "fail to get pool staker..")
	}
	ps.TotalUnits = pool.PoolUnits
	su := ps.GetStakerUnit(publicAddress)
	fex, err := strconv.ParseFloat(su.Units, 64)
	if nil != err {
		return errors.Wrap(err, "fail to parse staker's exist stake unit")

	}
	stakerUnits += fex
	err = validateStakeAmount(ps, stakerUnits)
	if err != nil {
		return errors.Wrapf(err, "invalid stake amount")
	}
	su.Units = strconv.FormatFloat(stakerUnits, 'f', floatPrecision, 64)
	ps.UpsertStakerUnit(su)
	keeper.SetPoolStaker(ctx, poolID, ps)
	// maintain stake pool structure
	sp, err := keeper.GetStakerPool(ctx, publicAddress)
	if nil != err {
		return errors.Wrap(err, "fail to get stakepool object")
	}
	stakerPoolItem := sp.GetStakerPoolItem(poolID)
	existUnit, err := strconv.ParseFloat(stakerPoolItem.Units, 64)
	if nil != err {
		return errors.Wrap(err, "fail to parse exist unit")
	}
	stakerUnits += existUnit
	stakerPoolItem.Units = strconv.FormatFloat(stakerUnits, 'f', floatPrecision, 64)
	stakerPoolItem.AddStakerTxDetail(requestTxHash, stakeRuneAmount, stakeTokenAmount)
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
