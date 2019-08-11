package swapservice

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

const stakerLimit = 100.0 // TODO: make configurable

// validateStakeAmount
func validateStakeAmount(stakers PoolStaker, stakerUnits float64) error {
	var minStakerAmt float64
	stakerCount := float64(len(stakers.Stakers))
	if stakerCount <= stakerLimit {
		minStakerAmt = 0 // first 100 stakers there are no lower limits
	} else {
		totalUnits := stakers.TotalUnits.Float64()
		avgStake := totalUnits / stakerCount

		minStakerAmt = avgStake * ((stakerCount / stakerLimit) + 0.1) // Increases minStakeAmt by 10% every 100 stakers
	}

	if stakerUnits < minStakerAmt {
		return fmt.Errorf("Not enough to stake (%f/%f)", stakerUnits, minStakerAmt)
	}

	return nil
}

// validateStakeMessage is to do some validation , and make sure it is legit
func validateStakeMessage(ctx sdk.Context, keeper Keeper, ticker Ticker, stakeRuneAmount, stakeTokenAmount Amount, requestTxHash TxID, publicAddress BnbAddress) error {
	if ticker.Empty() {
		return errors.New("ticker is empty")
	}
	if stakeRuneAmount.Empty() {
		return errors.New("stake rune amount is empty")
	}
	if stakeTokenAmount.Empty() {
		return errors.New("stake token amount is empty")
	}
	if requestTxHash.Empty() {
		return errors.New("request tx hash is empty")
	}
	if publicAddress.Empty() {
		return errors.New("public address is empty")
	}
	if !keeper.PoolExist(ctx, ticker) {
		return errors.Errorf("%s doesn't exist", ticker)
	}
	return nil
}

func stake(ctx sdk.Context, keeper Keeper, ticker Ticker, stakeRuneAmount, stakeTokenAmount Amount, publicAddress BnbAddress, requestTxHash TxID) error {
	ctx.Logger().Info(fmt.Sprintf("%s staking %s %s", ticker, stakeRuneAmount, stakeTokenAmount))
	if err := validateStakeMessage(ctx, keeper, ticker, stakeRuneAmount, stakeTokenAmount, requestTxHash, publicAddress); nil != err {
		return errors.Wrap(err, "invalid request")
	}
	pool := keeper.GetPoolStruct(ctx, ticker)
	fTokenAmt := stakeTokenAmount.Float64()
	fRuneAmt := stakeRuneAmount.Float64()
	ctx.Logger().Info(fmt.Sprintf("Pre-Pool: %sRUNE %sToken", pool.BalanceRune, pool.BalanceToken))
	ctx.Logger().Info(fmt.Sprintf("Staking: %sRUNE %sToken", stakeRuneAmount, stakeTokenAmount))

	balanceRune := pool.BalanceRune.Float64()
	balanceToken := pool.BalanceToken.Float64()

	oldPoolUnits := pool.PoolUnits.Float64()
	newPoolUnits, stakerUnits, err := calculatePoolUnits(oldPoolUnits, balanceRune, balanceToken, fRuneAmt, fTokenAmt)
	if nil != err {
		return errors.Wrapf(err, "fail to calculate pool units")
	}
	ctx.Logger().Info(fmt.Sprintf("current pool units : %f ,staker units : %f", newPoolUnits, stakerUnits))
	poolRune := balanceRune + fRuneAmt
	poolToken := balanceToken + fTokenAmt
	pool.PoolUnits = NewAmountFromFloat(newPoolUnits)
	pool.BalanceRune = NewAmountFromFloat(poolRune)
	pool.BalanceToken = NewAmountFromFloat(poolToken)
	ctx.Logger().Info(fmt.Sprintf("Post-Pool: %sRUNE %sToken", pool.BalanceRune, pool.BalanceToken))
	keeper.SetPoolStruct(ctx, ticker, pool)
	// maintain pool staker structure
	ps, err := keeper.GetPoolStaker(ctx, ticker)
	if nil != err {
		return errors.Wrap(err, "fail to get pool staker..")
	}
	ps.TotalUnits = pool.PoolUnits
	su := ps.GetStakerUnit(publicAddress)
	fex := su.Units.Float64()
	stakerUnits += fex
	err = validateStakeAmount(ps, stakerUnits)
	if err != nil {
		return errors.Wrapf(err, "invalid stake amount")
	}
	su.Units = NewAmountFromFloat(stakerUnits)
	ps.UpsertStakerUnit(su)
	keeper.SetPoolStaker(ctx, ticker, ps)
	// maintain stake pool structure
	sp, err := keeper.GetStakerPool(ctx, publicAddress)
	if nil != err {
		return errors.Wrap(err, "fail to get stakepool object")
	}
	stakerPoolItem := sp.GetStakerPoolItem(ticker)
	existUnit := stakerPoolItem.Units.Float64()
	stakerUnits += existUnit
	stakerPoolItem.Units = NewAmountFromFloat(stakerUnits)
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
