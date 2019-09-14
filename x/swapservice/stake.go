package swapservice

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/common"
)

// validateStakeAmount
func validateStakeAmount(stakers PoolStaker, stakerUnits sdk.Uint, stakeAmtInterval common.Amount) error {
	var minStakerAmt sdk.Uint
	stakerCount := float64(len(stakers.Stakers))
	if stakerCount <= stakeAmtInterval.Float64() {
		minStakerAmt = sdk.ZeroUint() // first 100 stakers there are no lower limits
	} else {
		totalUnits := stakers.TotalUnits
		avgStake := common.UintToFloat64(totalUnits) / stakerCount
		minStakerAmt = common.FloatToUint(avgStake * ((stakerCount / stakeAmtInterval.Float64()) + 0.1)) // Increases minStakeAmt by 10% every interval stakers
	}

	if stakerUnits.LT(minStakerAmt) {
		return fmt.Errorf("not enough to stake (%s/%s)", stakerUnits, minStakerAmt)
	}

	return nil
}

// validateStakeMessage is to do some validation , and make sure it is legit
func validateStakeMessage(ctx sdk.Context, keeper poolStorage, ticker common.Ticker, requestTxHash common.TxID, publicAddress common.BnbAddress) error {
	if ticker.IsEmpty() {
		return errors.New("ticker is empty")
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

func stake(ctx sdk.Context, keeper poolStorage, ticker common.Ticker, stakeRuneAmount, stakeTokenAmount sdk.Uint, publicAddress common.BnbAddress, requestTxHash common.TxID) (sdk.Uint, error) {
	ctx.Logger().Info(fmt.Sprintf("%s staking %s %s", ticker, stakeRuneAmount, stakeTokenAmount))
	if err := validateStakeMessage(ctx, keeper, ticker, requestTxHash, publicAddress); nil != err {
		return sdk.ZeroUint(), errors.Wrap(err, "invalid request")
	}
	if stakeTokenAmount.IsZero() && stakeTokenAmount.IsZero() {
		return sdk.ZeroUint(), errors.New("both rune and token is zero")
	}
	pool := keeper.GetPool(ctx, ticker)
	fTokenAmt := stakeTokenAmount
	fRuneAmt := stakeRuneAmount
	ctx.Logger().Info(fmt.Sprintf("Pre-Pool: %sRUNE %sToken", pool.BalanceRune, pool.BalanceToken))
	ctx.Logger().Info(fmt.Sprintf("Staking: %sRUNE %sToken", stakeRuneAmount, stakeTokenAmount))

	balanceRune := pool.BalanceRune
	balanceToken := pool.BalanceToken

	oldPoolUnits := pool.PoolUnits
	newPoolUnits, stakerUnits, err := calculatePoolUnits(oldPoolUnits, balanceRune, balanceToken, fRuneAmt, fTokenAmt)
	if nil != err {
		return sdk.ZeroUint(), errors.Wrapf(err, "fail to calculate pool units")
	}

	ctx.Logger().Info(fmt.Sprintf("current pool units : %f ,staker units : %f", newPoolUnits, stakerUnits))
	poolRune := balanceRune.Add(fRuneAmt)
	poolToken := balanceToken.Add(fTokenAmt)
	pool.PoolUnits = newPoolUnits
	pool.BalanceRune = poolRune
	pool.BalanceToken = poolToken
	ctx.Logger().Info(fmt.Sprintf("Post-Pool: %sRUNE %sToken", pool.BalanceRune, pool.BalanceToken))
	keeper.SetPool(ctx, pool)
	// maintain pool staker structure
	ps, err := keeper.GetPoolStaker(ctx, ticker)
	if nil != err {
		return sdk.ZeroUint(), errors.Wrap(err, "fail to get pool staker..")
	}
	ps.TotalUnits = pool.PoolUnits
	su := ps.GetStakerUnit(publicAddress)
	fex := su.Units
	totalStakerUnits := fex.Add(stakerUnits)

	stakeAmtInterval := keeper.GetAdminConfigStakerAmtInterval(ctx, common.NoBnbAddress)
	err = validateStakeAmount(ps, totalStakerUnits, stakeAmtInterval)
	if err != nil {
		return sdk.ZeroUint(), errors.Wrapf(err, "invalid stake amount")
	}
	su.Units = totalStakerUnits
	ps.UpsertStakerUnit(su)
	keeper.SetPoolStaker(ctx, ticker, ps)
	// maintain stake pool structure
	sp, err := keeper.GetStakerPool(ctx, publicAddress)
	if nil != err {
		return sdk.ZeroUint(), errors.Wrap(err, "fail to get stakepool object")
	}
	stakerPoolItem := sp.GetStakerPoolItem(ticker)
	existUnit := stakerPoolItem.Units
	stakerPoolItem.Units = totalStakerUnits.Add(existUnit)
	stakerPoolItem.AddStakerTxDetail(requestTxHash, stakeRuneAmount, stakeTokenAmount)
	sp.UpsertStakerPoolItem(stakerPoolItem)
	keeper.SetStakerPool(ctx, publicAddress, sp)
	return stakerUnits, nil
}

// calculatePoolUnits calculate the pool units and staker units
// returns newPoolUnit,stakerUnit, error
func calculatePoolUnits(oldPoolUnits, poolRune, poolToken, stakeRune, stakeToken sdk.Uint) (sdk.Uint, sdk.Uint, error) {

	if stakeRune.Add(poolRune).IsZero() {
		return sdk.ZeroUint(), sdk.ZeroUint(), errors.New("total RUNE in the pool is zero")
	}
	if stakeToken.Add(poolToken).IsZero() {
		return sdk.ZeroUint(), sdk.ZeroUint(), errors.New("total token in the pool is zero")
	}
	fStakeRune := common.UintToFloat64(stakeRune)
	fStakeToken := common.UintToFloat64(stakeToken)

	fPoolRune := common.UintToFloat64(poolRune)
	fPoolToken := common.UintToFloat64(poolToken)
	stakerPercentage := ((fStakeRune / (fStakeRune + fPoolRune)) + (fStakeToken / (fStakeToken + fPoolToken))) / 2

	stakerUnit := (stakerPercentage*(fStakeRune+fPoolRune) + stakerPercentage*(fStakeToken+fPoolToken)) / 2
	newPoolUnit := oldPoolUnits.Add(common.FloatToUint(stakerUnit))
	return newPoolUnit, common.FloatToUint(stakerUnit), nil
}
