package swapservice

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/thornode/common"
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
func validateMessage(source, target common.Ticker, amount sdk.Uint, requester, destination common.Address, requestTxHash common.TxID) error {
	if requestTxHash.IsEmpty() {
		return errors.New("request tx hash is empty")
	}
	if source.IsEmpty() {
		return errors.New("source is empty")
	}
	if target.IsEmpty() {
		return errors.New("target is empty")
	}
	if amount.IsZero() {
		return errors.New("amount is zero")
	}
	if requester.IsEmpty() {
		return errors.New("requester is empty")
	}
	if destination.IsEmpty() {
		return errors.New("destination is empty")
	}

	return nil
}

func swap(ctx sdk.Context,
	keeper poolStorage, txID common.TxID,
	source, target common.Ticker,
	amount sdk.Uint,
	requester, destination common.Address,
	requestTxHash common.TxID,
	tradeTarget sdk.Uint,
	globalSlipLimit common.Amount) (sdk.Uint, error) {
	if err := validateMessage(source, target, amount, requester, destination, requestTxHash); nil != err {
		ctx.Logger().Error(err.Error())
		return sdk.ZeroUint(), err
	}
	if err := validatePools(ctx, keeper, source, target); nil != err {
		ctx.Logger().Error(err.Error())
		return sdk.ZeroUint(), err
	}

	isDoubleSwap := !common.IsRune(source) && !common.IsRune(target)

	if isDoubleSwap {
		runeAmount, err := swapOne(ctx, keeper, txID, source, common.RuneTicker, amount, requester, destination, tradeTarget, globalSlipLimit)
		if err != nil {
			return sdk.ZeroUint(), errors.Wrapf(err, "fail to swap from %s to %s", source, common.RuneTicker)
		}
		tokenAmount, err := swapOne(ctx, keeper, txID, common.RuneTicker, target, runeAmount, requester, destination, tradeTarget, globalSlipLimit)
		return tokenAmount, err
	}
	tokenAmount, err := swapOne(ctx, keeper, txID, source, target, amount, requester, destination, tradeTarget, globalSlipLimit)
	return tokenAmount, err
}

func swapOne(ctx sdk.Context,
	keeper poolStorage, txID common.TxID,
	source, target common.Ticker,
	amount sdk.Uint, requester,
	destination common.Address,
	tradeTarget sdk.Uint,
	globalSlipLimit common.Amount) (amt sdk.Uint, err error) {

	ctx.Logger().Info(fmt.Sprintf("%s Swapping %s(%s) -> %s to %s", requester, source, amount, target, destination))

	var X, x, Y, liquitityFee, emitTokens sdk.Uint
	var priceSlip, tradeSlip, poolSlip, outputSlip float64

	// Set ticker to our non-rune token ticker
	ticker := source
	if common.IsRune(source) {
		ticker = target
	}

	// emit swap event at the end of the swap
	defer func() {
		var swapEvt EventSwap
		var status EventStatus
		if err == nil {
			status = EventSuccess
			swapEvt = NewEventSwap(
				common.NewCoin(common.BNBChain, source, x),
				common.NewCoin(common.BNBChain, target, emitTokens),
				common.FloatToUint(priceSlip*common.One),
				common.FloatToUint(tradeSlip*common.One),
				common.FloatToUint(poolSlip*common.One),
				common.FloatToUint(outputSlip*common.One),
				liquitityFee,
			)

		} else {
			status = EventRefund
			swapEvt = NewEventSwap(
				common.NewCoin(common.BNBChain, source, x),
				common.NewCoin(common.BNBChain, target, sdk.ZeroUint()),
				sdk.ZeroUint(),
				sdk.ZeroUint(),
				sdk.ZeroUint(),
				sdk.ZeroUint(),
				sdk.ZeroUint(),
			)
		}

		swapBytes, errr := json.Marshal(swapEvt)
		if errr != nil {
			amt = sdk.ZeroUint()
			err = errr
		}
		evt := NewEvent(
			swapEvt.Type(),
			txID,
			ticker,
			swapBytes,
			status,
		)

		keeper.AddIncompleteEvents(ctx, evt)

	}()

	// Check if pool exists
	if !keeper.PoolExist(ctx, ticker) {
		ctx.Logger().Debug(fmt.Sprintf("pool %s doesn't exist", ticker))
		return sdk.ZeroUint(), errors.New(fmt.Sprintf("pool %s doesn't exist", ticker))
	}

	// Get our pool from the KVStore
	pool := keeper.GetPool(ctx, ticker)
	if pool.Status != PoolEnabled {
		return sdk.ZeroUint(), errors.Errorf("pool %s is in %s status, can't swap", ticker, pool.Status)
	}

	// Get our slip limits
	gsl := globalSlipLimit.Float64() // global slip limit

	// get our X, x, Y values
	if common.IsRune(source) {
		X = pool.BalanceRune
		Y = pool.BalanceToken
	} else {
		Y = pool.BalanceRune
		X = pool.BalanceToken
	}
	x = amount

	// check our X,x,Y values are valid
	if x.IsZero() {
		return sdk.ZeroUint(), errors.New("amount is invalid")
	}
	if X.IsZero() || Y.IsZero() {
		return sdk.ZeroUint(), errors.New("invalid balance")
	}

	outputSlip = calcOutputSlip(X, x)
	liquitityFee = calcLiquitityFee(X, x, Y)
	tradeSlip = calcTradeSlip(X, x)
	emitTokens = calcTokenEmission(X, x, Y)
	poolSlip = calcPoolSlip(X, x)
	priceSlip = calcPriceSlip(X, x, Y)

	// do we have enough balance to swap?

	if emitTokens.GT(Y) {
		return sdk.ZeroUint(), errors.New("token :%s balance is 0, can't do swap")
	}
	// Need to convert to float before the calculation , otherwise 0.1 becomes 0, which is bad
	if !tradeTarget.IsZero() && emitTokens.LT(tradeTarget) {
		return sdk.ZeroUint(), errors.Errorf("emit token %s less than price limit %s", emitTokens, tradeTarget)
	}

	if poolSlip > gsl {
		ctx.Logger().Info("poolslip over global pool slip limit", "poolslip", fmt.Sprintf("%.2f", poolSlip), "gsl", fmt.Sprintf("%.2f", gsl))
		return sdk.ZeroUint(), errors.Errorf("pool slip:%f is over global pool slip limit :%f", poolSlip, gsl)
	}
	ctx.Logger().Info(fmt.Sprintf("Pre-Pool: %sRune %sToken", pool.BalanceRune, pool.BalanceToken))

	if common.IsRune(source) {
		pool.BalanceRune = X.Add(x)
		pool.BalanceToken = Y.Sub(emitTokens)
	} else {
		pool.BalanceToken = X.Add(x)
		pool.BalanceRune = Y.Sub(emitTokens)
	}
	keeper.SetPool(ctx, pool)
	ctx.Logger().Info(fmt.Sprintf("Post-swap: %sRune %sToken , user get:%s ", pool.BalanceRune, pool.BalanceToken, emitTokens))

	return emitTokens, nil
}

// calcPriceSlip - calculate the price slip
// This calculates the price slip by dividing the number of coins added, by the number of emitted tokens
func calcPriceSlip(X, x, Y sdk.Uint) float64 {
	tokenEmission := calcTokenEmission(X, x, Y)
	return common.UintToFloat64(x) / common.UintToFloat64(tokenEmission)
}

// calcTradeSlip - calculate the trade slip
func calcTradeSlip(iX, ix sdk.Uint) float64 {
	// x * ( 2X + x) / ( X * X )
	// have to do this , otherwise numbers are too big
	// poolSlip is by nature a float
	x := common.UintToFloat64(ix) / common.One
	X := common.UintToFloat64(iX) / common.One
	return x * (2*X + x) / (X * X)
}

// calcOutputSlip - calculates the output slip
func calcOutputSlip(X, x sdk.Uint) float64 {
	// ( x ) / ( x + X )
	denominator := x.Add(X)
	return common.UintToFloat64(x) / common.UintToFloat64(denominator)
}

// Calculates the pool slip
func calcPoolSlip(X, x sdk.Uint) float64 {
	// (x*(x^2 + 2*x*X + 2 X^2)) / (X*(x^2 + x*X + X^2))
	// have to do this , otherwise numbers are too big
	// poolSlip is by nature a float
	cX := common.UintToFloat64(X) / common.One
	cx := common.UintToFloat64(x) / common.One
	x2 := cx * cx
	X2 := cX * cX

	return (cx * (x2 + 2*cx*cX + 2*X2)) / (cX * (x2 + cx*cX + X2))
}

// calculateFee the fee of the swap
func calcLiquitityFee(X, x, Y sdk.Uint) sdk.Uint {
	// ( x^2 *  Y ) / ( x + X )^2
	numerator := x.Mul(x).Mul(Y)
	denominator := x.Add(X).Mul(x.Add(X))
	return numerator.Quo(denominator)
}

// calculate the number of tokens sent to the address (includes liquidity fee)
func calcTokenEmission(X, x, Y sdk.Uint) sdk.Uint {
	// ( x * X * Y ) / ( x + X )^2
	numerator := x.Mul(X).Mul(Y)
	denominator := x.Add(X).Mul(x.Add(X))
	return numerator.Quo(denominator)
}
