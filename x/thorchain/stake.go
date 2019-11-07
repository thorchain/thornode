package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/thornode/common"
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
func validateStakeMessage(ctx sdk.Context, keeper poolStorage, asset common.Asset, requestTxHash common.TxID, runeAddr, assetAddr common.Address) error {
	if asset.IsEmpty() {
		return errors.New("asset is empty")
	}
	if requestTxHash.IsEmpty() {
		return errors.New("request tx hash is empty")
	}
	if common.IsBNBChain(asset.Chain) {
		if runeAddr.IsEmpty() {
			return errors.New("rune address is empty")
		}
	} else {
		if assetAddr.IsEmpty() {
			return errors.New("asset address is empty")
		}
	}
	if !keeper.PoolExist(ctx, asset) {
		return errors.Errorf("%s doesn't exist", asset)
	}
	return nil
}

func stake(ctx sdk.Context, keeper poolStorage, asset common.Asset, stakeRuneAmount, stakeAssetAmount sdk.Uint, runeAddr, assetAddr common.Address, requestTxHash common.TxID) (sdk.Uint, error) {
	ctx.Logger().Info(fmt.Sprintf("%s staking %s %s", asset, stakeRuneAmount, stakeAssetAmount))
	if err := validateStakeMessage(ctx, keeper, asset, requestTxHash, runeAddr, assetAddr); nil != err {
		return sdk.ZeroUint(), errors.Wrap(err, "invalid request")
	}
	if stakeRuneAmount.IsZero() && stakeAssetAmount.IsZero() {
		return sdk.ZeroUint(), errors.New("both rune and asset is zero")
	}
	if runeAddr.IsEmpty() {
		return sdk.ZeroUint(), errors.New("Rune address cannot be empty")
	}

	pool := keeper.GetPool(ctx, asset)

	ps, err := keeper.GetPoolStaker(ctx, asset)
	if nil != err {
		return sdk.ZeroUint(), errors.Wrap(err, "fail to get pool staker..")
	}

	su := ps.GetStakerUnit(runeAddr)
	su.Height = sdk.NewUint(uint64(ctx.BlockHeight()))
	if su.RuneAddress.IsEmpty() {
		su.RuneAddress = runeAddr
	}
	if su.AssetAddress.IsEmpty() {
		su.AssetAddress = assetAddr
	} else {
		if !su.AssetAddress.Equals(assetAddr) {
			// mismatch of asset addresses from what is known to the address
			// given. Refund it.
			return sdk.ZeroUint(), errors.Wrap(err, "Mismatch of asset addresses")
		}
	}

	if !common.IsBNBChain(asset.Chain) {
		if stakeAssetAmount.IsZero() {
			su.PendingRune = su.PendingRune.Add(stakeRuneAmount)
			ps.UpsertStakerUnit(su)
			keeper.SetPoolStaker(ctx, asset, ps)
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
		return sdk.ZeroUint(), errors.Wrapf(err, "fail to calculate pool units")
	}

	ctx.Logger().Info(fmt.Sprintf("current pool units : %s ,staker units : %s", newPoolUnits, stakerUnits))
	poolRune := balanceRune.Add(fRuneAmt)
	poolAsset := balanceAsset.Add(fAssetAmt)
	pool.PoolUnits = newPoolUnits
	pool.BalanceRune = poolRune
	pool.BalanceAsset = poolAsset
	ctx.Logger().Info(fmt.Sprintf("Post-Pool: %sRUNE %sAsset", pool.BalanceRune, pool.BalanceAsset))
	keeper.SetPool(ctx, pool)
	// maintain pool staker structure

	ps.TotalUnits = pool.PoolUnits
	fex := su.Units
	totalStakerUnits := fex.Add(stakerUnits)

	stakeAmtInterval := keeper.GetAdminConfigStakerAmtInterval(ctx, EmptyAccAddress)
	err = validateStakeAmount(ps, totalStakerUnits, stakeAmtInterval)
	if err != nil {
		return sdk.ZeroUint(), errors.Wrapf(err, "invalid stake amount")
	}
	su.Units = totalStakerUnits
	ps.UpsertStakerUnit(su)
	keeper.SetPoolStaker(ctx, asset, ps)
	// maintain stake pool structure
	sp, err := keeper.GetStakerPool(ctx, runeAddr)
	if nil != err {
		return sdk.ZeroUint(), errors.Wrap(err, "fail to get stakepool object")
	}
	stakerPoolItem := sp.GetStakerPoolItem(asset)
	existUnit := stakerPoolItem.Units
	stakerPoolItem.Units = totalStakerUnits.Add(existUnit)
	stakerPoolItem.AddStakerTxDetail(requestTxHash, stakeRuneAmount, stakeAssetAmount)
	sp.UpsertStakerPoolItem(stakerPoolItem)
	keeper.SetStakerPool(ctx, runeAddr, sp)
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
	fStakeRune := common.UintToFloat64(stakeRune)
	fStakeAsset := common.UintToFloat64(stakeAsset)

	fPoolRune := common.UintToFloat64(poolRune)
	fPoolAsset := common.UintToFloat64(poolAsset)
	stakerPercentage := ((fStakeRune / (fStakeRune + fPoolRune)) + (fStakeAsset / (fStakeAsset + fPoolAsset))) / 2

	stakerUnit := (stakerPercentage*(fStakeRune+fPoolRune) + stakerPercentage*(fStakeAsset+fPoolAsset)) / 2
	newPoolUnit := oldPoolUnits.Add(common.FloatToUint(stakerUnit))
	return newPoolUnit, common.FloatToUint(stakerUnit), nil
}
