package swapservice

import (
	"fmt"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/common"
)

// validateStakeAmount
func validateStakeAmount(stakers PoolStaker, stakerUnits float64, stakeAmtInterval common.Amount) error {
	var minStakerAmt float64
	interval := stakeAmtInterval.Float64()
	stakerCount := float64(len(stakers.Stakers))
	if stakerCount <= interval {
		minStakerAmt = 0 // first 100 stakers there are no lower limits
	} else {
		totalUnits := stakers.TotalUnits.Float64()
		avgStake := totalUnits / stakerCount

		minStakerAmt = avgStake * ((stakerCount / interval) + 0.1) // Increases minStakeAmt by 10% every interval stakers
	}

	if stakerUnits < minStakerAmt {
		return fmt.Errorf("not enough to stake (%f/%f)", stakerUnits, minStakerAmt)
	}

	return nil
}

// validateStakeMessage is to do some validation , and make sure it is legit
func validateStakeMessage(ctx sdk.Context, keeper poolStorage, ticker common.Ticker, stakeRuneAmount, stakeTokenAmount common.Amount, requestTxHash common.TxID, publicAddress common.BnbAddress) error {
	if ticker.IsEmpty() {
		return errors.New("ticker is empty")
	}
	if stakeRuneAmount.IsEmpty() {
		return errors.New("stake rune amount is empty")
	}
	if stakeTokenAmount.IsEmpty() {
		return errors.New("stake token amount is empty")
	}
	if requestTxHash.IsEmpty() {
		return errors.New("request tx hash is empty")
	}
	if publicAddress.IsEmpty() {
		return errors.New("public address is empty")
	}
	if !keeper.PoolExist(ctx, ticker) {
		return errors.Errorf("%s doesn't exist", ticker)
	}
	return nil
}

func stake(ctx sdk.Context, keeper poolStorage, ticker common.Ticker, stakeRuneAmount, stakeTokenAmount common.Amount, publicAddress common.BnbAddress, requestTxHash common.TxID) (common.Amount, error) {
	ctx.Logger().Info(fmt.Sprintf("%s staking %s %s", ticker, stakeRuneAmount, stakeTokenAmount))
	if err := validateStakeMessage(ctx, keeper, ticker, stakeRuneAmount, stakeTokenAmount, requestTxHash, publicAddress); nil != err {
		return "0", errors.Wrap(err, "invalid request")
	}
	pool := keeper.GetPool(ctx, ticker)
	fTokenAmt := stakeTokenAmount.Float64()
	fRuneAmt := stakeRuneAmount.Float64()
	ctx.Logger().Info(fmt.Sprintf("Pre-Pool: %sRUNE %sToken", pool.BalanceRune, pool.BalanceToken))
	ctx.Logger().Info(fmt.Sprintf("Staking: %sRUNE %sToken", stakeRuneAmount, stakeTokenAmount))

	balanceRune := pool.BalanceRune.Float64()
	balanceToken := pool.BalanceToken.Float64()

	oldPoolUnits := pool.PoolUnits.Float64()
	newPoolUnits, stakerUnits, err := calculatePoolUnits(oldPoolUnits, balanceRune, balanceToken, fRuneAmt, fTokenAmt)
	if nil != err {
		return "0", errors.Wrapf(err, "fail to calculate pool units")
	}

	ctx.Logger().Info(fmt.Sprintf("current pool units : %f ,staker units : %f", newPoolUnits, stakerUnits))
	poolRune := balanceRune + fRuneAmt
	poolToken := balanceToken + fTokenAmt
	pool.PoolUnits = common.NewAmountFromFloat(newPoolUnits)
	pool.BalanceRune = common.NewAmountFromFloat(poolRune)
	pool.BalanceToken = common.NewAmountFromFloat(poolToken)
	ctx.Logger().Info(fmt.Sprintf("Post-Pool: %sRUNE %sToken", pool.BalanceRune, pool.BalanceToken))
	keeper.SetPool(ctx, pool)
	// maintain pool staker structure
	ps, err := keeper.GetPoolStaker(ctx, ticker)
	if nil != err {
		return "0", errors.Wrap(err, "fail to get pool staker..")
	}
	ps.TotalUnits = pool.PoolUnits
	su := ps.GetStakerUnit(publicAddress)
	fex := su.Units.Float64()
	totalStakerUnits := fex + stakerUnits

	stakeAmtInterval := keeper.GetAdminConfigStakerAmtInterval(ctx, common.NoBnbAddress)
	err = validateStakeAmount(ps, totalStakerUnits, stakeAmtInterval)
	if err != nil {
		return "0", errors.Wrapf(err, "invalid stake amount")
	}
	su.Units = common.NewAmountFromFloat(totalStakerUnits)
	ps.UpsertStakerUnit(su)
	keeper.SetPoolStaker(ctx, ticker, ps)
	// maintain stake pool structure
	sp, err := keeper.GetStakerPool(ctx, publicAddress)
	if nil != err {
		return "0", errors.Wrap(err, "fail to get stakepool object")
	}
	stakerPoolItem := sp.GetStakerPoolItem(ticker)
	existUnit := stakerPoolItem.Units.Float64()
	totalStakerUnits += existUnit
	stakerPoolItem.Units = common.NewAmountFromFloat(totalStakerUnits)
	stakerPoolItem.AddStakerTxDetail(requestTxHash, stakeRuneAmount, stakeTokenAmount)
	sp.UpsertStakerPoolItem(stakerPoolItem)
	keeper.SetStakerPool(ctx, publicAddress, sp)
	return common.NewAmountFromFloat(stakerUnits), nil
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
	//
	if (stakeRune + poolRune) == 0 {
		return 0, 0, errors.New("total RUNE in the pool is zero")
	}
	if (stakeToken + poolToken) == 0 {
		return 0, 0, errors.New("total token in the pool is zero")
	}
	stakerPercentage := ((stakeRune / (stakeRune + poolRune)) + (stakeToken / (stakeToken + poolToken))) / 2
	stakerUnit := (stakerPercentage*(stakeRune+poolRune) + stakerPercentage*(stakeToken+poolToken)) / 2
	if math.IsNaN(stakerUnit) {
		return 0, 0, errors.New("fail to calculate pool units")
	}
	newPoolUnit := oldPoolUnits + stakerUnit
	return newPoolUnit, stakerUnit, nil
}
