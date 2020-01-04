package thorchain

import (
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
	tradeTarget sdk.Uint,
	transactionFee sdk.Uint) (sdk.Uint, []EventSwap, error) {
	var swapEvents []EventSwap

	if err := validateMessage(tx, target, destination); nil != err {
		ctx.Logger().Error(err.Error())
		return sdk.ZeroUint(), nil, err
	}
	source := tx.Coins[0].Asset

	if err := validatePools(ctx, keeper, source, target); nil != err {
		ctx.Logger().Error(err.Error())
		return sdk.ZeroUint(), nil, err
	}
	var swapEvt EventSwap
	pools := make([]Pool, 0)
	isDoubleSwap := !source.IsRune() && !target.IsRune()
	if isDoubleSwap {
		var err error
		sourcePool, err := keeper.GetPool(ctx, source)
		if err != nil {
			return sdk.ZeroUint(), nil, err
		}
		tx.Coins[0].Amount, sourcePool, swapEvt, err = swapOne(ctx, keeper, tx, sourcePool, common.RuneAsset(), destination, tradeTarget, transactionFee)
		if err != nil {
			return sdk.ZeroUint(), nil, fmt.Errorf("fail to swap from %s to %s :%w", source, common.RuneAsset(), err)
		}
		pools = append(pools, sourcePool)
		tx.Coins[0].Asset = common.RuneAsset()
		swapEvents = append(swapEvents, swapEvt)
	}

	// Set asset to our non-rune asset asset
	asset := source
	if source.IsRune() {
		asset = target
	}
	pool, err := keeper.GetPool(ctx, asset)
	if err != nil {
		return sdk.ZeroUint(), nil, err
	}
	assetAmount, pool, swapEvt, err := swapOne(ctx, keeper, tx, pool, target, destination, tradeTarget, transactionFee)
	if err != nil {
		return sdk.ZeroUint(), nil, fmt.Errorf("fail to swap from %s to %s :%w", source, target, err)
	}
	pools = append(pools, pool)
	if !tradeTarget.IsZero() && assetAmount.LT(tradeTarget) {
		return sdk.ZeroUint(), nil, fmt.Errorf("emit asset %s less than price limit %s", assetAmount, tradeTarget)
	}
	if target.IsRune() {
		if assetAmount.LTE(transactionFee) {
			return sdk.ZeroUint(), nil, fmt.Errorf("output RUNE (%s) is not enough to pay transaction fee", assetAmount)
		}
	}
	// emit asset is zero
	if assetAmount.IsZero() {
		return sdk.ZeroUint(), nil, errors.New("emit asset is zero")
	}
	// Update pools
	for _, pool := range pools {
		if err := keeper.SetPool(ctx, pool); err != nil {
			return sdk.ZeroUint(), nil, err
		}
	}
	swapEvents = append(swapEvents, swapEvt)
	return assetAmount, swapEvents, nil
}

func swapOne(ctx sdk.Context,
	keeper Keeper, tx common.Tx, pool Pool,
	target common.Asset,
	destination common.Address,
	tradeTarget sdk.Uint,
	transactionFee sdk.Uint) (amt sdk.Uint, poolResult Pool, swapEvt EventSwap, err error) {

	source := tx.Coins[0].Asset
	amount := tx.Coins[0].Amount
	swapEvt = NewEventSwap(
		source,
		tradeTarget,
		sdk.ZeroUint(),
		sdk.ZeroUint(),
	)
	ctx.Logger().Info(fmt.Sprintf("%s Swapping %s(%s) -> %s to %s", tx.FromAddress, source, tx.Coins[0].Amount, target, destination))

	var X, x, Y, liquidityFee, emitAssets sdk.Uint
	var tradeSlip sdk.Uint

	// Set asset to our non-rune asset asset
	asset := source
	if source.IsRune() {
		asset = target
		if amount.LTE(transactionFee) {
			// stop swap , because the output will not enough to pay for transaction fee
			return sdk.ZeroUint(), Pool{}, swapEvt, fmt.Errorf("output RUNE (%s) is not enough to pay transaction fee", amount)
		}
	}

	// Check if pool exists
	if !keeper.PoolExist(ctx, asset) {
		ctx.Logger().Debug(fmt.Sprintf("pool %s doesn't exist", asset))
		return sdk.ZeroUint(), Pool{}, swapEvt, errors.New(fmt.Sprintf("pool %s doesn't exist", asset))
	}

	// Get our pool from the KVStore
	pool, err = keeper.GetPool(ctx, asset)
	if err != nil {
		return sdk.ZeroUint(), Pool{}, swapEvt, err
	}
	if pool.Status != PoolEnabled {
		return sdk.ZeroUint(), pool, swapEvt, errors.Errorf("pool %s is in %s status, can't swap", asset.String(), pool.Status)
	}

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
		return sdk.ZeroUint(), pool, swapEvt, errors.New("amount is invalid")
	}
	if X.IsZero() || Y.IsZero() {
		return sdk.ZeroUint(), pool, swapEvt, errors.New("invalid balance")
	}

	liquidityFee = calcLiquidityFee(X, x, Y)
	tradeSlip = calcTradeSlip(X, x)
	emitAssets = calcAssetEmission(X, x, Y)

	if source.IsRune() {
		liquidityFee = pool.AssetValueInRune(liquidityFee)
	}
	swapEvt.LiquidityFee = liquidityFee
	swapEvt.TradeSlip = tradeSlip
	err = keeper.AddToLiquidityFees(ctx, pool.Asset, liquidityFee)
	if err != nil {
		return sdk.ZeroUint(), pool, swapEvt, errors.Wrap(err, "failed to add liquidity")
	}

	// do THORNode have enough balance to swap?
	if emitAssets.GT(Y) {
		return sdk.ZeroUint(), pool, swapEvt, errors.New("asset :%s balance is 0, can't do swap")
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

	return emitAssets, pool, swapEvt, nil
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
	tradeSlipD := numD.Quo(denD) // Division with DECs

	tradeSlip := tradeSlipD.Mul(dec10k)                          // Adds 5 0's
	tradeSlipUint := sdk.NewUint(uint64(tradeSlip.RoundInt64())) // Casts back to Uint as Basis Points
	return tradeSlipUint
}
