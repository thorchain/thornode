package thorchain

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"gitlab.com/thorchain/thornode/common"
)

// validate if pools exist
func validatePools(ctx sdk.Context, keeper Keeper, assets ...common.Asset) error {
	for _, asset := range assets {
		if !asset.IsRune() {
			if !keeper.PoolExist(ctx, asset) {
				return errors.New(fmt.Sprintf("%s doesn't exist", asset))
			}
			pool, err := keeper.GetPool(ctx, asset)
			if err != nil {
				return err
			}

			if pool.Status != PoolEnabled {
				return errors.Errorf("pool %s is in %s status, can't swap", asset, pool.Status)
			}
		}

	}
	return nil
}

// validateMessage is trying to validate the legitimacy of the incoming message and decide whether THORNode can handle it
func validateMessage(tx common.Tx, target common.Asset, destination common.Address) error {
	if err := tx.IsValid(); err != nil {
		return err
	}
	if target.IsEmpty() {
		return errors.New("target is empty")
	}
	if destination.IsEmpty() {
		return errors.New("destination is empty")
	}

	return nil
}

func swap(ctx sdk.Context,
	keeper Keeper, tx common.Tx,
	target common.Asset,
	destination common.Address,
	tradeTarget,
	globalSlipLimit sdk.Uint) (sdk.Uint, error) {
	if err := validateMessage(tx, target, destination); nil != err {
		ctx.Logger().Error(err.Error())
		return sdk.ZeroUint(), err
	}
	source := tx.Coins[0].Asset
	if err := validatePools(ctx, keeper, source, target); nil != err {
		ctx.Logger().Error(err.Error())
		return sdk.ZeroUint(), err
	}

	pools := make([]Pool, 0)

	isDoubleSwap := !source.IsRune() && !target.IsRune()

	if isDoubleSwap {
		var err error
		sourcePool, err := keeper.GetPool(ctx, source)
		if err != nil {
			return sdk.ZeroUint(), err
		}

		tx.Coins[0].Amount, sourcePool, err = swapOne(ctx, keeper, tx, sourcePool, common.RuneAsset(), destination, tradeTarget, globalSlipLimit)
		if err != nil {
			return sdk.ZeroUint(), errors.Wrapf(err, "fail to swap from %s to %s", source, common.RuneAsset())
		}
		pools = append(pools, sourcePool)
		tx.Coins[0].Asset = common.RuneAsset()
	}

	// Set asset to our non-rune asset asset
	asset := source
	if source.IsRune() {
		asset = target
	}
	pool, err := keeper.GetPool(ctx, asset)
	if err != nil {
		return sdk.ZeroUint(), err
	}
	assetAmount, pool, err := swapOne(ctx, keeper, tx, pool, target, destination, tradeTarget, globalSlipLimit)

	if err != nil {
		return sdk.ZeroUint(), errors.Wrapf(err, "fail to swap from %s to %s", source, target)
	}
	pools = append(pools, pool)
	if !tradeTarget.IsZero() && assetAmount.LT(tradeTarget) {
		return sdk.ZeroUint(), errors.Errorf("emit asset %s less than price limit %s", assetAmount, tradeTarget)
	}

	// Update pools
	for _, pool := range pools {
		if err := keeper.SetPool(ctx, pool); err != nil {
			return sdk.ZeroUint(), err
		}
	}
	return assetAmount, nil
}

func swapOne(ctx sdk.Context,
	keeper Keeper, tx common.Tx, pool Pool,
	target common.Asset,
	destination common.Address,
	tradeTarget,
	globalSlipLimit sdk.Uint) (amt sdk.Uint, poolResult Pool, err error) {

	source := tx.Coins[0].Asset
	amount := tx.Coins[0].Amount
	ctx.Logger().Info(fmt.Sprintf("%s Swapping %s(%s) -> %s to %s", tx.FromAddress, source, tx.Coins[0].Amount, target, destination))

	var X, x, Y, liquidityFee, emitAssets sdk.Uint
	var tradeSlip sdk.Uint

	// Set asset to our non-rune asset asset
	asset := source
	if source.IsRune() {
		asset = target
	}

	// emit swap event at the end of the swap
	defer func() {
		var swapEvt EventSwap
		var status EventStatus
		if err == nil {
			status = EventPending
			swapEvt = NewEventSwap(
				source,
				tradeTarget,
				liquidityFee,
				tradeSlip,
			)

		} else {
			status = EventFail
			swapEvt = NewEventSwap(
				source,
				tradeTarget,
				sdk.ZeroUint(),
				sdk.ZeroUint(),
			)
		}

		swapBytes, errr := json.Marshal(swapEvt)
		if errr != nil {
			amt = sdk.ZeroUint()
			err = errr
			return
		}

		evt := NewEvent(

			swapEvt.Type(),
			ctx.BlockHeight(),
			tx,
			swapBytes,
			status,
		)
		// using errr instead of err because we don't want to return the error,
		// just log it because we are in a defer func
		if errr := keeper.UpsertEvent(ctx, evt); nil != errr {
			amt = sdk.ZeroUint()
			err = errr
		}

	}()

	// Check if pool exists
	if !keeper.PoolExist(ctx, asset) {
		ctx.Logger().Debug(fmt.Sprintf("pool %s doesn't exist", asset))
		return sdk.ZeroUint(), Pool{}, errors.New(fmt.Sprintf("pool %s doesn't exist", asset))
	}

	// Get our pool from the KVStore
	pool, err = keeper.GetPool(ctx, asset)
	if err != nil {
		return sdk.ZeroUint(), Pool{}, err
	}
	if pool.Status != PoolEnabled {
		return sdk.ZeroUint(), pool, errors.Errorf("pool %s is in %s status, can't swap", asset.String(), pool.Status)
	}

	// Get our slip limits
	gsl := globalSlipLimit

	// Get our X, x, Y values
	if source.IsRune() {
		X = pool.BalanceRune
		Y = pool.BalanceAsset
	} else {
		Y = pool.BalanceRune
		X = pool.BalanceAsset
	}
	x = amount

	// check our X,x,Y values are valid
	if x.IsZero() {
		return sdk.ZeroUint(), pool, errors.New("amount is invalid")
	}
	if X.IsZero() || Y.IsZero() {
		return sdk.ZeroUint(), pool, errors.New("invalid balance")
	}

	liquidityFee = calcLiquidityFee(X, x, Y)
	tradeSlip = calcTradeSlip(X, x)
	emitAssets = calcAssetEmission(X, x, Y)

	if source.IsRune() {
		liquidityFee = pool.AssetValueInRune(liquidityFee)
	}
	err = keeper.AddToLiquidityFees(ctx, pool.Asset, liquidityFee)
	if err != nil {
		return sdk.ZeroUint(), pool, errors.Wrap(err, "failed to add liquidity")
	}

	// do THORNode have enough balance to swap?
	if emitAssets.GT(Y) {
		return sdk.ZeroUint(), pool, errors.New("asset :%s balance is 0, can't do swap")
	}
	// Prevent exceeding the Global Slip Limit
	if tradeSlip.GT(gsl) {
		ctx.Logger().Info("poolslip over global slip limit", "tradeSlip", fmt.Sprintf("%s", tradeSlip), "gsl", fmt.Sprintf("%s", gsl))
		return sdk.ZeroUint(), pool, errors.Errorf("tradeSlip:%s is over global slip limit :%s", tradeSlip, gsl)
	}
	ctx.Logger().Info(fmt.Sprintf("Pre-Pool: %sRune %sAsset", pool.BalanceRune, pool.BalanceAsset))

	if source.IsRune() {
		pool.BalanceRune = X.Add(x)
		pool.BalanceAsset = common.SafeSub(Y, emitAssets)
	} else {
		pool.BalanceAsset = X.Add(x)
		pool.BalanceRune = common.SafeSub(Y, emitAssets)
	}
	ctx.Logger().Info(fmt.Sprintf("Post-swap: %sRune %sAsset , user get:%s ", pool.BalanceRune, pool.BalanceAsset, emitAssets))

	return emitAssets, pool, nil
}

// calculate the number of assets sent to the address (includes liquidity fee)
func calcAssetEmission(X, x, Y sdk.Uint) sdk.Uint {
	// ( x * X * Y ) / ( x + X )^2
	numerator := x.Mul(X).Mul(Y)
	denominator := x.Add(X).Mul(x.Add(X))
	return numerator.Quo(denominator)
}

// calculateFee the fee of the swap
func calcLiquidityFee(X, x, Y sdk.Uint) sdk.Uint {
	// ( x^2 *  Y ) / ( x + X )^2
	numerator := x.Mul(x).Mul(Y)
	denominator := x.Add(X).Mul(x.Add(X))
	return numerator.Quo(denominator)
}

// calcTradeSlip - calculate the trade slip, expressed in basis points (10000)
func calcTradeSlip(Xi, xi sdk.Uint) sdk.Uint {

	// Cast to DECs
	xD := sdk.NewDec(int64(xi.Uint64()))
	XD := sdk.NewDec(int64(Xi.Uint64()))
	dec2 := sdk.NewDec(int64(sdk.NewUint(2).Uint64()))
	dec10k := sdk.NewDec(int64(sdk.NewUint(10000).Uint64()))

	// x * (2*X + x) / (X * X)
	numD := xD.Mul((dec2.Mul(XD)).Add(xD))
	denD := XD.Mul(XD)
	tradeSlipD := (numD.Quo(denD)) // Division with DECs

	tradeSlip := tradeSlipD.Mul(dec10k)                            // Adds 5 0's
	tradeSlipUint := sdk.NewUint(uint64((tradeSlip.RoundInt64()))) // Casts back to Uint as Basis Points
	return tradeSlipUint
}

// numD := sdk.NewDecFromBigInt(num.BigInt())
// denD := sdk.NewDecFromBigInt(den.BigInt())
