package swapservice

import (
	"encoding/json"
	"fmt"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/common"
)

// validate if pools exist
func validatePools(ctx sdk.Context, keeper poolStorage, tickers ...common.Ticker) error {
	for _, ticker := range tickers {
		if !common.IsRune(ticker) {
			if !keeper.PoolExist(ctx, ticker) {
				return errors.New(fmt.Sprintf("%s doesn't exist", ticker))
			}
		}

	}
	return nil
}

// validateMessage is trying to validate the legitimacy of the incoming message and decide whether we can handle it
func validateMessage(source, target common.Ticker, amount common.Amount, requester, destination common.BnbAddress, requestTxHash common.TxID) error {
	if requestTxHash.IsEmpty() {
		return errors.New("request tx hash is empty")
	}
	if source.IsEmpty() {
		return errors.New("source is empty")
	}
	if target.IsEmpty() {
		return errors.New("target is empty")
	}
	if amount.IsEmpty() {
		return errors.New("amount is empty")
	}
	if requester.IsEmpty() {
		return errors.New("requester is empty")
	}
	if destination.IsEmpty() {
		return errors.New("destination is empty")
	}

	return nil
}

func swap(ctx sdk.Context, keeper poolStorage, txID common.TxID, source, target common.Ticker, amount common.Amount, requester, destination common.BnbAddress, requestTxHash common.TxID, tradeTarget, tradeSlipLimit, globalSlipLimit common.Amount) (common.Amount, error) {
	if err := validateMessage(source, target, amount, requester, destination, requestTxHash); nil != err {
		ctx.Logger().Error(err.Error())
		return "0", err
	}
	if err := validatePools(ctx, keeper, source, target); nil != err {
		ctx.Logger().Error(err.Error())
		return "0", err
	}

	isDoubleSwap := !common.IsRune(source) && !common.IsRune(target)

	if isDoubleSwap {
		runeAmount, err := swapOne(ctx, keeper, txID, source, common.RuneTicker, amount, requester, destination, tradeTarget, tradeSlipLimit, globalSlipLimit)
		if err != nil {
			return "0", errors.Wrapf(err, "fail to swap from %s to %s", source, common.RuneTicker)
		}
		tokenAmount, err := swapOne(ctx, keeper, txID, common.RuneTicker, target, runeAmount, requester, destination, tradeTarget, tradeSlipLimit, globalSlipLimit)
		return tokenAmount, err
	}
	tokenAmount, err := swapOne(ctx, keeper, txID, source, target, amount, requester, destination, tradeTarget, tradeSlipLimit, globalSlipLimit)
	return tokenAmount, err
}

func swapOne(ctx sdk.Context,
	keeper poolStorage, txID common.TxID,
	source, target common.Ticker,
	amount common.Amount, requester,
	destination common.BnbAddress,
	tradeTarget, tradeSlipLimit, globalSlipLimit common.Amount) (common.Amount, error) {

	ctx.Logger().Info(fmt.Sprintf("%s Swapping %s(%s) -> %s to %s", requester, source, amount, target, destination))

	// Set ticker to our non-rune token ticker
	ticker := source
	if common.IsRune(source) {
		ticker = target
	}

	// Check if pool exists
	if !keeper.PoolExist(ctx, ticker) {
		ctx.Logger().Debug(fmt.Sprintf("pool %s doesn't exist", ticker))
		return "0", errors.New(fmt.Sprintf("pool %s doesn't exist", ticker))
	}

	// Get our pool from the KVStore
	pool := keeper.GetPool(ctx, ticker)
	if pool.Status != PoolEnabled {
		return "0", errors.Errorf("pool %s is in %s status, can't swap", ticker, pool.Status)
	}

	// Get our slip limits
	tsl := tradeSlipLimit.Float64()  // trade slip limit
	gsl := globalSlipLimit.Float64() // global slip limit

	// get our X, x, Y values
	var X, x, Y float64
	if common.IsRune(source) {
		X = pool.BalanceRune.Float64()
		Y = pool.BalanceToken.Float64()
	} else {
		Y = pool.BalanceRune.Float64()
		X = pool.BalanceToken.Float64()
	}
	x = amount.Float64()

	// check our X,x,Y values are valid
	if x <= 0.0 {
		return "0", errors.New("amount is invalid")
	}
	if X <= 0 || Y <= 0 {
		return "0", errors.New("invalid balance")
	}

	outputSlip := calcOutputSlip(X, x)
	liquitityFee := calcLiquitityFee(X, x, Y)
	tradeSlip := calcTradeSlip(X, x)
	emitTokens := calcTokenEmission(X, x, Y)
	poolSlip := calcPoolSlip(X, x)
	priceSlip := calcPriceSlip(X, x, Y)

	// do we have enough balance to swap?
	if emitTokens > Y {
		return "0", errors.New("token :%s balance is 0, can't do swap")
	}

	if tradeTarget.GreaterThen(0) {
		tTarget := tradeTarget.Float64() // trade target
		if math.Abs((priceSlip)-tTarget)/tTarget > tsl {
			return "0", errors.Errorf("trade slip %f is more than %.2f percent different than %f", priceSlip, tsl*100, tTarget)
		}
	}
	if poolSlip > gsl {
		return "0", errors.Errorf("pool slip:%f is over global pool slip limit :%f", poolSlip, gsl)
	}
	ctx.Logger().Info(fmt.Sprintf("Pre-Pool: %sRune %sToken", pool.BalanceRune, pool.BalanceToken))

	if common.IsRune(source) {
		pool.BalanceRune = common.NewAmountFromFloat(X + x)
		pool.BalanceToken = common.NewAmountFromFloat(Y - emitTokens)
	} else {
		pool.BalanceToken = common.NewAmountFromFloat(X + x)
		pool.BalanceRune = common.NewAmountFromFloat(Y - emitTokens)
	}
	keeper.SetPool(ctx, pool)
	ctx.Logger().Info(fmt.Sprintf("Post-swap: %sRune %sToken , user get:%g ", pool.BalanceRune, pool.BalanceToken, emitTokens))

	swapEvt := NewEventSwap(
		common.NewCoin(source, common.NewAmountFromFloat(x)),
		common.NewCoin(target, common.NewAmountFromFloat(emitTokens)),
		common.NewAmountFromFloat(priceSlip),
		common.NewAmountFromFloat(tradeSlip),
		common.NewAmountFromFloat(poolSlip),
		common.NewAmountFromFloat(outputSlip),
		common.NewAmountFromFloat(liquitityFee),
	)
	swapBytes, err := json.Marshal(swapEvt)
	if err != nil {
		return "0", errors.Wrap(err, "fail to marshal swap event")
	}
	evt := NewEvent(
		swapEvt.Type(),
		txID,
		ticker,
		swapBytes,
		EventSuccess,
	)
	keeper.AddIncompleteEvents(ctx, evt)

	return common.NewAmountFromFloat(emitTokens), nil
}

// calcPriceSlip - calculate the price slip
// This calculates the price slip by dividing the number of coins added, by the number of emitted tokens
func calcPriceSlip(X, x, Y float64) float64 {
	return x / calcTokenEmission(X, x, Y)
}

// calcTradeSlip - calculate the trade slip
func calcTradeSlip(X, x float64) float64 {
	// x * ( 2X + x) / ( X * X )
	return x * (2*X + x) / (X * X)
}

// calcOutputSlip - calculates the output slip
func calcOutputSlip(X, x float64) float64 {
	// ( x ) / ( x + X )
	return x / (x + X)
}

// Calculates the pool slip
func calcPoolSlip(X, x float64) float64 {
	// (x*(x^2 + 2*x*X + 2 X^2)) / (X*(x^2 + x*X + X^2))
	x2 := x * x
	X2 := X * X
	return (x * (x2 + 2*x*X + 2*X2)) / (X * (x2 + x*X + X2))
}

// calculateFee the fee of the swap
func calcLiquitityFee(X, x, Y float64) float64 {
	// ( x^2 *  Y ) / ( x + X )^2
	return ((x * x) * Y) / ((x + X) * (x + X))
}

// calculate the number of tokens sent to the address (includes liquidity fee)
func calcTokenEmission(X, x, Y float64) float64 {
	// ( x * X * Y ) / ( x + X )^2
	return (x * X * Y) / ((x + X) * (x + X))
}
