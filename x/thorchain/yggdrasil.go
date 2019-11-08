package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
)

func Fund(ctx sdk.Context, keeper Keeper, txOutStore *TxOutStore, ygg Yggdrasil) error {
	nodeAcc, err := keeper.GetNodeAccountByPubKey(ctx, ygg.PubKey)
	if err != nil {
		return err
	}

	yggHoldings := getHoldingsValue(ctx, keeper, ygg)
	// calculate amount of assets (in rune) the ygg is entitled to (which is
	// half their bond)
	yggTarget := nodeAcc.Bond.QuoUint64(2).Sub(yggHoldings)

	// check if their target rune is greater than 1/4 of their bond, top up if
	// it is true
	if yggTarget.GT(nodeAcc.Bond.QuoUint64(4)) {
		coins, err := calculateTopUpYgg(ctx, keeper, yggTarget, ygg)
		if err != nil {
			return err
		}

		return sendCoinsToYggdrasil(ctx, keeper, coins, ygg, txOutStore)
	}

	return nil
}

// sendCoinsToYggdrasil - adds outbound txs to send the given coins to a
// yggdrasil pool
func sendCoinsToYggdrasil(ctx sdk.Context, keeper Keeper, coins common.Coins, ygg Yggdrasil, txOutStore *TxOutStore) error {
	for _, coin := range coins {
		to, err := ygg.PubKey.GetAddress(coin.Asset.Chain)
		if err != nil {
			return err
		}

		toi := &TxOutItem{
			Chain:     coin.Asset.Chain,
			ToAddress: to,
			Memo:      "yggdrasil+",
			Coins:     common.Coins{coin},
		}
		txOutStore.AddTxOutItem(ctx, keeper, toi, true)
	}

	return nil
}

// calculateTopUpYgg - with a given target (total value in rune), calculate
// yggdrasil pool assets from all pools, equally distributed relative to pool
// depth
func calculateTopUpYgg(ctx sdk.Context, keeper Keeper, target sdk.Uint, ygg Yggdrasil) (common.Coins, error) {
	assets, err := keeper.GetPoolIndex(ctx)
	if err != nil {
		return nil, err
	}

	runeCoin := common.NewCoin(common.RuneAsset(), sdk.ZeroUint())
	var coins common.Coins

	totalUnits := sdk.ZeroUint()
	var pools []Pool
	for _, asset := range assets {
		pool := keeper.GetPool(ctx, asset)
		totalUnits = totalUnits.Add(pool.PoolUnits)
		pools = append(pools, pool)
	}

	for _, pool := range pools {
		ratio := pool.PoolUnits.Quo(totalUnits)
		totalAmt := target.Mul(ratio)
		runeAmt := totalAmt.QuoUint64(2)
		runeCoin.Amount = runeCoin.Amount.Add(runeAmt)
		assetAmt := pool.RuneValueInAsset(runeAmt)
		coins = append(coins, common.NewCoin(pool.Asset, assetAmt))
	}

	coins = append(coins, runeCoin)
	return coins, nil
}

// getHoldingsValue - adds up all assets a yggdrasil pool has in rune
func getHoldingsValue(ctx sdk.Context, keeper Keeper, ygg Yggdrasil) sdk.Uint {
	yggHoldingsInRune := sdk.ZeroUint()
	for _, coin := range ygg.Coins {
		if coin.Asset.IsRune() {
			yggHoldingsInRune = yggHoldingsInRune.Add(coin.Amount)
		} else {
			pool := keeper.GetPool(ctx, coin.Asset)
			runeValue := pool.AssetValueInRune(coin.Amount)
			yggHoldingsInRune = yggHoldingsInRune.Add(runeValue)
		}
	}

	return yggHoldingsInRune
}
