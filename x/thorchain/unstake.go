package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/thornode/common"
)

func validateUnstake(ctx sdk.Context, keeper poolStorage, msg MsgSetUnStake) error {
	if msg.RuneAddress.IsEmpty() {
		return errors.New("empty rune address")
	}
	if msg.RequestTxHash.IsEmpty() {
		return errors.New("request tx hash is empty")
	}
	if msg.Asset.IsEmpty() {
		return errors.New("empty asset")
	}
	withdrawBasisPoints := msg.WithdrawBasisPoints
	if withdrawBasisPoints.GT(sdk.ZeroUint()) && withdrawBasisPoints.GT(sdk.NewUint(MaxWithdrawBasisPoints)) {
		return errors.Errorf("withdraw basis points %s is invalid", msg.WithdrawBasisPoints)
	}
	if !keeper.PoolExist(ctx, msg.Asset) {
		// pool doesn't exist
		return errors.Errorf("pool-%s doesn't exist", msg.Asset)
	}
	return nil
}

// unstake withdraw all the asset
func unstake(ctx sdk.Context, keeper poolStorage, msg MsgSetUnStake) (sdk.Uint, sdk.Uint, sdk.Uint, error) {
	if err := validateUnstake(ctx, keeper, msg); nil != err {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), err
	}

	// here fBalance should be valid , because we did the validation above
	pool := keeper.GetPool(ctx, msg.Asset)
	poolStaker, err := keeper.GetPoolStaker(ctx, msg.Asset)
	if nil != err {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.Wrap(err, "can't find pool staker")

	}
	stakerPool, err := keeper.GetStakerPool(ctx, msg.RuneAddress)
	if nil != err {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.Wrap(err, "can't find staker pool")
	}

	poolUnits := pool.PoolUnits
	poolRune := pool.BalanceRune
	poolAsset := pool.BalanceAsset
	stakerUnit := poolStaker.GetStakerUnit(msg.RuneAddress)
	fStakerUnit := stakerUnit.Units
	if !stakerUnit.Units.GT(sdk.ZeroUint()) {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.New("nothing to withdraw")
	}

	// check if we need to rate limit unstaking
	// https://gitlab.com/thorchain/bepswap/thornode/issues/166
	if !msg.Asset.Chain.Equals(common.BNBChain) {
		height := sdk.NewUint(uint64(ctx.BlockHeight()))
		if height.LT(stakerUnit.Height.Add(sdk.NewUint(17280))) {
			err := fmt.Errorf("You cannot unstake for 24 hours after staking for this blockchain")
			return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), err
		}
	}

	ctx.Logger().Info("pool before unstake", "pool unit", poolUnits, "balance RUNE", poolRune, "balance asset", poolAsset)
	ctx.Logger().Info("staker before withdraw", "staker unit", fStakerUnit)
	withdrawRune, withDrawAsset, unitAfter, err := calculateUnstake(poolUnits, poolRune, poolAsset, fStakerUnit, msg.WithdrawBasisPoints)
	if err != nil {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), err
	}

	withdrawRune = withdrawRune.Add(stakerUnit.PendingRune) // extract pending rune
	stakerUnit.PendingRune = sdk.ZeroUint()                 // reset pending to zero

	ctx.Logger().Info("client withdraw", "RUNE", withdrawRune, "asset", withDrawAsset, "units left", unitAfter)
	// update pool
	pool.PoolUnits = poolUnits.Sub(fStakerUnit).Add(unitAfter)
	pool.BalanceRune = poolRune.Sub(withdrawRune)
	pool.BalanceAsset = poolAsset.Sub(withDrawAsset)
	ctx.Logger().Info("pool after unstake", "pool unit", pool.PoolUnits, "balance RUNE", pool.BalanceRune, "balance asset", pool.BalanceAsset)
	// update pool staker
	poolStaker.TotalUnits = pool.PoolUnits
	if unitAfter.IsZero() {
		// just remove it
		poolStaker.RemoveStakerUnit(msg.RuneAddress)
	} else {
		stakerUnit.Units = unitAfter
		poolStaker.UpsertStakerUnit(stakerUnit)
	}
	if unitAfter.IsZero() {
		stakerPool.RemoveStakerPoolItem(msg.Asset)
	} else {
		spi := stakerPool.GetStakerPoolItem(msg.Asset)
		spi.Units = unitAfter
		stakerPool.UpsertStakerPoolItem(spi)
	}
	// update staker pool
	keeper.SetPool(ctx, pool)
	keeper.SetPoolStaker(ctx, msg.Asset, poolStaker)
	keeper.SetStakerPool(ctx, msg.RuneAddress, stakerPool)
	return withdrawRune, withDrawAsset, fStakerUnit.Sub(unitAfter), nil
}

func calculateUnstake(poolUnit, poolRune, poolAsset, stakerUnit, withdrawBasisPoints sdk.Uint) (sdk.Uint, sdk.Uint, sdk.Uint, error) {
	if poolUnit.IsZero() {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.New("poolUnits can't be zero")
	}
	if poolRune.IsZero() {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.New("pool rune balance can't be zero")
	}
	if poolAsset.IsZero() {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.New("pool asset balance can't be zero")
	}
	if stakerUnit.IsZero() {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.New("staker unit can't be zero")
	}
	if withdrawBasisPoints.GT(sdk.NewUint(MaxWithdrawBasisPoints)) {
		return sdk.ZeroUint(), sdk.ZeroUint(), sdk.ZeroUint(), errors.Errorf("withdraw basis point %s is not valid", withdrawBasisPoints.String())
	}
	percentage := common.UintToFloat64(withdrawBasisPoints) / 100
	stakerOwnership := common.UintToFloat64(stakerUnit) / common.UintToFloat64(poolUnit)

	//withdrawRune := stakerOwnership.Mul(withdrawBasisPoints).Quo(sdk.NewUint(10000)).Mul(poolRune)
	//withdrawAsset := stakerOwnership.Mul(withdrawBasisPoints).Quo(sdk.NewUint(10000)).Mul(poolAsset)
	//unitAfter := stakerUnit.Mul(sdk.NewUint(MaxWithdrawBasisPoints).Sub(withdrawBasisPoints).Quo(sdk.NewUint(10000)))
	withdrawRune := stakerOwnership * percentage / 100 * common.UintToFloat64(poolRune)
	withdrawAsset := stakerOwnership * percentage / 100 * common.UintToFloat64(poolAsset)
	unitAfter := common.UintToFloat64(stakerUnit) * (100 - percentage) / 100
	return common.FloatToUint(withdrawRune), common.FloatToUint(withdrawAsset), common.FloatToUint(unitAfter), nil
}
