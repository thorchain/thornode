package swapservice

import (
	"fmt"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

// validate if pools exist
func validatePools(ctx sdk.Context, keeper poolStorage, tickers ...Ticker) error {
	for _, ticker := range tickers {
		if !IsRune(ticker) {
			if !keeper.PoolExist(ctx, ticker) {
				return errors.New(fmt.Sprintf("%s doesn't exist", ticker))
			}
		}

	}
	return nil
}

// validateMessage is trying to validate the legitimacy of the incoming message and decide whether we can handle it
func validateMessage(source, target Ticker, amount Amount, requester, destination BnbAddress, requestTxHash TxID) error {
	if requestTxHash.Empty() {
		return errors.New("request tx hash is empty")
	}
	if source.Empty() {
		return errors.New("source is empty")
	}
	if target.Empty() {
		return errors.New("target is empty")
	}
	if amount.Empty() {
		return errors.New("amount is empty")
	}
	if requester.Empty() {
		return errors.New("requester is empty")
	}
	if destination.Empty() {
		return errors.New("destination is empty")
	}

	return nil
}

func swap(ctx sdk.Context, keeper poolStorage, source, target Ticker, amount Amount, requester, destination BnbAddress, requestTxHash TxID, tradeTarget, tradeSlipLimit, globalSlipLimit Amount) (Amount, error) {
	if err := validateMessage(source, target, amount, requester, destination, requestTxHash); nil != err {
		ctx.Logger().Error(err.Error())
		return "0", err
	}
	if err := validatePools(ctx, keeper, source, target); nil != err {
		ctx.Logger().Error(err.Error())
		return "0", err
	}

	isDoubleSwap := !IsRune(source) && !IsRune(target)

	swapRecord := NewSwapRecord(requestTxHash, source, target, requester, destination, amount, "", "")

	if isDoubleSwap {
		runeAmount, err := swapOne(ctx, keeper, source, RuneTicker, amount, requester, destination, tradeTarget, tradeSlipLimit, globalSlipLimit)
		if err != nil {
			return "0", errors.Wrapf(err, "fail to swap from %s to %s", source, RuneTicker)
		}
		tokenAmount, err := swapOne(ctx, keeper, RuneTicker, target, runeAmount, requester, destination, tradeTarget, tradeSlipLimit, globalSlipLimit)
		swapRecord.AmountPaidBack = tokenAmount
		if err := keeper.SetSwapRecord(ctx, swapRecord); nil != err {
			ctx.Logger().Error("fail to save swap record", "error", err)
		}
		return tokenAmount, err
	}
	tokenAmount, err := swapOne(ctx, keeper, source, target, amount, requester, destination, tradeTarget, tradeSlipLimit, globalSlipLimit)
	swapRecord.AmountPaidBack = tokenAmount
	if err := keeper.SetSwapRecord(ctx, swapRecord); nil != err {
		ctx.Logger().Error("fail to save swap record", "error", err)
	}
	return tokenAmount, err
}

func swapOne(ctx sdk.Context,
	keeper poolStorage,
	source, target Ticker,
	amount Amount, requester,
	destination BnbAddress,
	tradeTarget, tradeSlipLimit, globalSlipLimit Amount) (Amount, error) {

	ctx.Logger().Info(fmt.Sprintf("%s Swapping %s(%s) -> %s to %s", requester, source, amount, target, destination))

	ticker := source
	if IsRune(source) {
		ticker = target
	}
	if !keeper.PoolExist(ctx, ticker) {
		ctx.Logger().Debug(fmt.Sprintf("pool %s doesn't exist", ticker))
		return "0", errors.New(fmt.Sprintf("pool %s doesn't exist", ticker))
	}

	amt := amount.Float64()
	tsl := tradeSlipLimit.Float64()  // trade slip limit
	gsl := globalSlipLimit.Float64() // global slip limit

	pool := keeper.GetPool(ctx, ticker)
	if pool.Status != PoolEnabled {
		return "0", errors.Errorf("pool %s is in %s status, can't swap", ticker, pool.Status)
	}
	balanceRune := pool.BalanceRune.Float64()
	balanceToken := pool.BalanceToken.Float64()
	if !tradeTarget.Empty() && !tradeTarget.Zero() {
		tTarget := tradeTarget.Float64() // trade target
		userPrice := calculateUserPrice(source, balanceRune, balanceToken, amt)
		if math.Abs(userPrice-tTarget)/tTarget > tsl {
			return "0", errors.Errorf("user price %f is more than %.2f percent different than %f", userPrice, tsl*100, tTarget)
		}
	}
	// do we have enough balance to swap?
	if IsRune(source) {
		if balanceToken == 0 {
			return "0", errors.New("token :%s balance is 0, can't do swap")
		}
	} else {
		if balanceRune == 0 {
			return "0", errors.New(RuneTicker.String() + " balance is 0, can't swap ")
		}
	}
	poolSlip := calculatePoolSlip(source, balanceRune, balanceToken, amt)
	if poolSlip > gsl {
		return "0", errors.Errorf("pool slip:%f is over global pool slip limit :%f", poolSlip, gsl)
	}
	ctx.Logger().Info(fmt.Sprintf("Pre-Pool: %sRune %sToken", pool.BalanceRune, pool.BalanceToken))
	newBalanceRune, newBalanceToken, returnAmt, err := calculateSwap(source, balanceRune, balanceToken, amt)
	if nil != err {
		return "0", errors.Wrap(err, "fail to swap")
	}
	pool.BalanceRune = NewAmountFromFloat(newBalanceRune)
	pool.BalanceToken = NewAmountFromFloat(newBalanceToken)
	returnTokenAmount := NewAmountFromFloat(returnAmt)
	keeper.SetPool(ctx, ticker, pool)
	ctx.Logger().Info(fmt.Sprintf("Post-swap: %sRune %sToken , user get:%s ", pool.BalanceRune, pool.BalanceToken, returnTokenAmount))
	return returnTokenAmount, nil
}

// calculateUserPrice return trade slip
func calculateUserPrice(source Ticker, balanceRune, balanceToken, amt float64) float64 {
	if IsRune(source) {
		return math.Pow(balanceRune+amt, 2.0) / (balanceRune * balanceToken)
	}
	return math.Pow(balanceToken+amt, 2.0) / (balanceRune * balanceToken)
}

// calculatePoolSlip the slip of total pool
func calculatePoolSlip(source Ticker, balanceRune, balanceToken, amt float64) float64 {
	if IsRune(source) {
		return amt * (2*balanceRune + amt) / math.Pow(balanceRune, 2.0)
	}
	return amt * (2*balanceToken + amt) / math.Pow(balanceToken, 2.0)
}

// calculateSwap how much rune, token and amount to emit
// return (Rune,Token,Amount)
func calculateSwap(source Ticker, balanceRune, balanceToken, amt float64) (float64, float64, float64, error) {
	if amt <= 0.0 {
		return balanceRune, balanceToken, 0.0, errors.New("amount is invalid")
	}
	if balanceRune <= 0 || balanceToken <= 0 {
		return balanceRune, balanceToken, amt, errors.New("invalid balance")
	}
	if IsRune(source) {
		balanceRune += amt
		tokenAmount := (amt * balanceToken) / balanceRune
		liquidityFee := math.Pow(amt, 2.0) * balanceToken / math.Pow(balanceRune, 2.0)
		tokenAmount -= liquidityFee
		balanceToken = balanceToken - tokenAmount
		return balanceRune, balanceToken, tokenAmount, nil
	} else {
		balanceToken += amt
		runeAmt := (balanceRune * amt) / balanceToken
		liquidityFee := (math.Pow(amt, 2.0) * balanceRune) / math.Pow(balanceToken, 2.0)
		runeAmt -= liquidityFee
		balanceRune = balanceRune - runeAmt
		return balanceRune, balanceToken, runeAmt, nil
	}
}

// swapComplete  mark a swap to be in complete state
func swapComplete(ctx sdk.Context, keeper poolStorage, requestTxHash, payTxHash TxID) error {
	if requestTxHash.Empty() {
		return errors.New("request tx hash is empty")
	}
	if payTxHash.Empty() {
		return errors.New("pay tx hash is empty")
	}
	return keeper.UpdateSwapRecordPayTxHash(ctx, requestTxHash, payTxHash)
}
