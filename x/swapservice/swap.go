package swapservice

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/jpthor/cosmos-swap/x/swapservice/types"
)

func isEmptyString(input string) bool {
	return strings.TrimSpace(input) == ""
}

// validateMessage is trying to validate the legitimacy of the incoming message and decide whether we can handle it
func validateMessage(ctx sdk.Context, keeper poolStorage, source, target, amount, requester, destination string) error {
	if isEmptyString(source) {
		return errors.New("source is empty")
	}
	if isEmptyString(target) {
		return errors.New("target is empty")
	}
	if isEmptyString(amount) {
		return errors.New("amount is empty")
	}
	if isEmptyString(requester) {
		return errors.New("requester is empty")
	}
	if isEmptyString(destination) {
		return errors.New("destination is empty")
	}
	if source != types.RuneTicker {
		poolID := types.GetPoolNameFromTicker(source)
		if !keeper.PoolExist(ctx, poolID) {
			return errors.New(fmt.Sprintf("%s doesn't exist", poolID))
		}
	}
	if !strings.EqualFold(target, types.RuneTicker) {
		poolID := types.GetPoolNameFromTicker(target)
		if !keeper.PoolExist(ctx, poolID) {
			return errors.New(fmt.Sprintf("%s doesn't exist", poolID))
		}
	}
	return nil
}

func swap(ctx sdk.Context, keeper poolStorage, source, target, amount, requester, destination string) (string, error) {
	if err := validateMessage(ctx, keeper, source, target, amount, requester, destination); nil != err {
		ctx.Logger().Error(err.Error())
		return "0", err
	}
	isDoubleSwap := !isRune(source) && !isRune(target)

	source = strings.ToUpper(source)
	target = strings.ToUpper(target)

	if isDoubleSwap {
		runeAmount, err := swapOne(ctx, keeper, source, types.RuneTicker, amount, requester, destination)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("fail to swap from %s to %s ", source, types.RuneTicker))
			return "0", errors.Wrapf(err, "fail to swap from %s to %s", source, types.RuneTicker)
		}
		return swapOne(ctx, keeper, types.RuneTicker, target, runeAmount, requester, destination)
	}
	return swapOne(ctx, keeper, source, target, amount, requester, destination)
}

func isRune(ticker string) bool {
	return strings.EqualFold(ticker, types.RuneTicker)
}

func swapOne(ctx sdk.Context, keeper poolStorage, source, target, amount, requester, destination string) (string, error) {
	ctx.Logger().Info(fmt.Sprintf("%s Swapping %s(%s) -> %s to %s", requester, source, amount, target, destination))
	poolID := types.GetPoolNameFromTicker(source)

	if isRune(source) {
		poolID = types.GetPoolNameFromTicker(target)
	}
	if !keeper.PoolExist(ctx, poolID) {
		ctx.Logger().Debug(fmt.Sprintf("pool %s doesn't exist", poolID))
		return "0", errors.New(fmt.Sprintf("pool %s doesn't exist", poolID))
	}

	amt, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("amount:%s is not valid", amount))
		return "0", err
	}
	pool := keeper.GetPoolStruct(ctx, poolID)

	balanceRune, err := strconv.ParseFloat(pool.BalanceRune, 64)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("pool rune balance  %s is invalid", pool.BalanceRune))
		return "0", errors.Wrap(err, "pool rune balance is invalid")
	}
	balanceToken, err := strconv.ParseFloat(pool.BalanceToken, 64)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("pool token balance %s is invalid", pool.BalanceToken))
		return "0", errors.Wrap(err, "pool token balance is invalid")
	}
	// do we have enough balance to swap?
	if isRune(source) {
		if balanceToken == 0 {
			return "0", errors.New("token :%s balance is 0, can't do swap")
		}
	} else {
		if balanceRune == 0 {
			return "0", errors.New(types.RuneTicker + " balance is 0, can't swap ")
		}
	}
	ctx.Logger().Info(fmt.Sprintf("Pre-Pool: %sRune %sToken", pool.BalanceRune, pool.BalanceToken))
	newBalanceRune, newBalanceToken, returnAmt, err := calculateSwap(source, balanceRune, balanceToken, amt)
	if nil != err {
		return "0", errors.Wrap(err, "fail to swap")
	}
	pool.BalanceRune = strconv.FormatFloat(newBalanceRune, 'f', 8, 64)
	pool.BalanceToken = strconv.FormatFloat(newBalanceToken, 'f', 8, 64)
	returnTokenAmount := strconv.FormatFloat(returnAmt, 'f', 8, 64)
	keeper.SetPoolStruct(ctx, poolID, pool)
	ctx.Logger().Info(fmt.Sprintf("Post-swap: %sRune %sToken , user get:%s ", pool.BalanceRune, pool.BalanceToken, returnTokenAmount))
	return returnTokenAmount, nil
}

// calculateSwap how much rune, token and amount to emit
// return (Rune,Token,Amount)
func calculateSwap(source string, balanceRune, balanceToken, amt float64) (float64, float64, float64, error) {
	if amt <= 0.0 {
		return balanceRune, balanceToken, 0.0, errors.New("amount is invalid")
	}
	if balanceRune <= 0 || balanceToken <= 0 {
		return balanceRune, balanceToken, amt, errors.New("invalid balance")
	}
	if isRune(source) {
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
