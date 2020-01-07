package thorchain

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// validateStakeMessage is to do some validation, and make sure it is legit
func validateStakeMessage(ctx sdk.Context, keeper Keeper, asset common.Asset, requestTxHash common.TxID, runeAddr, assetAddr common.Address) error {
	if asset.IsEmpty() {
		return errors.New("asset is empty")
	}
	if requestTxHash.IsEmpty() {
		return errors.New("request tx hash is empty")
	}
	if asset.Chain.IsBNB() {
		if runeAddr.IsEmpty() {
			return errors.New("rune address is empty")
		}
	} else {
		if assetAddr.IsEmpty() {
			return errors.New("asset address is empty")
		}
	}
	if !keeper.PoolExist(ctx, asset) {
		return fmt.Errorf("%s doesn't exist", asset)
	}
	return nil
}

func stake(ctx sdk.Context, keeper Keeper, asset common.Asset, stakeRuneAmount, stakeAssetAmount sdk.Uint, runeAddr, assetAddr common.Address, requestTxHash common.TxID) (sdk.Uint, sdk.Error) {
	ctx.Logger().Info(fmt.Sprintf("%s staking %s %s", asset, stakeRuneAmount, stakeAssetAmount))
	if err := validateStakeMessage(ctx, keeper, asset, requestTxHash, runeAddr, assetAddr); nil != err {
		ctx.Logger().Error("stake message fail validation", err)
		return sdk.ZeroUint(), sdk.NewError(DefaultCodespace, CodeStakeFailValidation, err.Error())
	}
	if stakeRuneAmount.IsZero() && stakeAssetAmount.IsZero() {
		return sdk.ZeroUint(), sdk.NewError(DefaultCodespace, CodeStakeFailValidation, "both rune and asset is zero")
	}
	if runeAddr.IsEmpty() {
		return sdk.ZeroUint(), sdk.NewError(DefaultCodespace, CodeStakeFailValidation, "Rune address cannot be empty")
	}

	pool, err := keeper.GetPool(ctx, asset)
	if err != nil {
		ctx.Logger().Error("fail to get pool", err)
		return sdk.ZeroUint(), sdk.ErrInternal(fmt.Sprintf("fail to get pool(%s)", asset))
	}

	// if THORNode have no balance, set the default pool status
	if pool.BalanceAsset.IsZero() && pool.BalanceRune.IsZero() {
		status := keeper.GetAdminConfigDefaultPoolStatus(ctx, nil)
		pool.Status = status
	}

	ps, err := keeper.GetPoolStaker(ctx, asset)
	if nil != err {
		ctx.Logger().Error("fail to get pool staker record", err)
		return sdk.ZeroUint(), sdk.NewError(DefaultCodespace, CodeFailGetPoolStaker, "fail to get pool staker record")
	}

	su := ps.GetStakerUnit(runeAddr)
	su.Height = ctx.BlockHeight()
	if su.RuneAddress.IsEmpty() {
		su.RuneAddress = runeAddr
	}
	if su.AssetAddress.IsEmpty() {
		su.AssetAddress = assetAddr
	} else {
		if !su.AssetAddress.Equals(assetAddr) {
			// mismatch of asset addresses from what is known to the address
			// given. Refund it.
			return sdk.ZeroUint(), sdk.NewError(DefaultCodespace, CodeStakeMismatchAssetAddr, "Mismatch of asset addresses")
		}
	}

	if !asset.Chain.IsBNB() {
		if stakeAssetAmount.IsZero() {
			su.PendingRune = su.PendingRune.Add(stakeRuneAmount)
			ps.UpsertStakerUnit(su)
			keeper.SetPoolStaker(ctx, ps)
			return sdk.ZeroUint(), nil
		}
		stakeRuneAmount = su.PendingRune.Add(stakeRuneAmount)
		su.PendingRune = sdk.ZeroUint()
	}

	fAssetAmt := stakeAssetAmount
	fRuneAmt := stakeRuneAmount

	ctx.Logger().Info(fmt.Sprintf("Pre-Pool: %sRUNE %sAsset", pool.BalanceRune, pool.BalanceAsset))
	ctx.Logger().Info(fmt.Sprintf("Staking: %sRUNE %sAsset", stakeRuneAmount, stakeAssetAmount))

	balanceRune := pool.BalanceRune
	balanceAsset := pool.BalanceAsset

	oldPoolUnits := pool.PoolUnits
	newPoolUnits, stakerUnits, err := calculatePoolUnits(oldPoolUnits, balanceRune, balanceAsset, fRuneAmt, fAssetAmt)
	if nil != err {
		ctx.Logger().Error("fail to calculate pool unit", err)
		return sdk.ZeroUint(), sdk.NewError(DefaultCodespace, CodeStakeInvalidPoolAsset, err.Error())
	}

	ctx.Logger().Info(fmt.Sprintf("current pool units : %s ,staker units : %s", newPoolUnits, stakerUnits))
	poolRune := balanceRune.Add(fRuneAmt)
	poolAsset := balanceAsset.Add(fAssetAmt)
	pool.PoolUnits = newPoolUnits
	pool.BalanceRune = poolRune
	pool.BalanceAsset = poolAsset
	ctx.Logger().Info(fmt.Sprintf("Post-Pool: %sRUNE %sAsset", pool.BalanceRune, pool.BalanceAsset))
	if err := keeper.SetPool(ctx, pool); err != nil {
		ctx.Logger().Error("fail to save pool", err)
		return sdk.ZeroUint(), sdk.ErrInternal("fail to save pool")
	}
	// maintain pool staker structure

	ps.TotalUnits = pool.PoolUnits
	fex := su.Units
	totalStakerUnits := fex.Add(stakerUnits)

	su.Units = totalStakerUnits
	ps.UpsertStakerUnit(su)
	keeper.SetPoolStaker(ctx, ps)
	// maintain stake pool structure
	sp, err := keeper.GetStakerPool(ctx, runeAddr)
	if nil != err {
		ctx.Logger().Error("fail to get staker pool object", err)
		return sdk.ZeroUint(), sdk.ErrInternal("fail to get staker pool object")
	}
	stakerPoolItem := sp.GetStakerPoolItem(asset)
	existUnit := stakerPoolItem.Units
	stakerPoolItem.Units = totalStakerUnits.Add(existUnit)
	stakerPoolItem.AddStakerTxDetail(requestTxHash, stakeRuneAmount, stakeAssetAmount)
	sp.UpsertStakerPoolItem(stakerPoolItem)
	keeper.SetStakerPool(ctx, sp)
	return stakerUnits, nil
}

// calculatePoolUnits calculate the pool units and staker units
// returns newPoolUnit,stakerUnit, error
func calculatePoolUnits(oldPoolUnits, poolRune, poolAsset, stakeRune, stakeAsset sdk.Uint) (sdk.Uint, sdk.Uint, error) {

	if stakeRune.Add(poolRune).IsZero() {
		return sdk.ZeroUint(), sdk.ZeroUint(), errors.New("total RUNE in the pool is zero")
	}
	if stakeAsset.Add(poolAsset).IsZero() {
		return sdk.ZeroUint(), sdk.ZeroUint(), errors.New("total asset in the pool is zero")
	}

	poolRuneAfter := poolRune.Add(stakeRune)
	poolAssetAfter := poolAsset.Add(stakeAsset)

	// ((R + A) * (r * A + R * a))/(4 * R * A)
	nominator1 := poolRuneAfter.Add(poolAssetAfter)
	nominator2 := stakeRune.Mul(poolAssetAfter).Add(poolRuneAfter.Mul(stakeAsset))
	denominator := sdk.NewUint(4).Mul(poolRuneAfter).Mul(poolAssetAfter)
	stakeUnits := nominator1.Mul(nominator2).Quo(denominator)
	newPoolUnit := oldPoolUnits.Add(stakeUnits)
	return newPoolUnit, stakeUnits, nil
}
