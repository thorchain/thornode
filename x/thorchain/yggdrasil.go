package thorchain

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// Fund is a method to fund yggdrasil pool
func Fund(ctx sdk.Context, keeper Keeper, txOutStore TxOutStore, constAccessor constants.ConstantValues) error {
	// Check if we have triggered the ragnarok protocol
	ragnarokHeight, err := keeper.GetRagnarokBlockHeight(ctx)
	if err != nil {
		return fmt.Errorf("fail to get ragnarok height: %w", err)
	}
	if ragnarokHeight > 0 {
		return nil
	}

	// Check we're not migrating funds
	retiring, err := keeper.GetAsgardVaultsByStatus(ctx, RetiringVault)
	if err != nil {
		ctx.Logger().Error("fail to get retiring vaults", "error", err)
		return err
	}
	if len(retiring) > 0 {
		// skip yggdrasil funding while a migration is in progress
		return nil
	}

	// find total bonded
	totalBond := sdk.ZeroUint()
	nodeAccs, err := keeper.ListActiveNodeAccounts(ctx)
	if err != nil {
		return err
	}
	minimumNodesForYggdrasil := constAccessor.GetInt64Value(constants.MinimumNodesForYggdrasil)
	if int64(len(nodeAccs)) <= minimumNodesForYggdrasil {
		return nil
	}

	// Gather list of all pools
	pools, err := keeper.GetPools(ctx)
	if err != nil {
		return err
	}

	for _, na := range nodeAccs {
		totalBond = totalBond.Add(na.Bond)
	}

	// We don't want to check all Yggdrasil pools every time THORNode run this
	// function. So THORNode use modulus to determine which Ygg THORNode process. This
	// should behave as a "round robin" approach checking one Ygg per block.
	// With 100 Ygg pools, THORNode should check each pool every 8.33 minutes.
	na := nodeAccs[ctx.BlockHeight()%int64(len(nodeAccs))]

	// check that we have enough bond
	minBond, err := keeper.GetMimir(ctx, constants.MinimumBondInRune.String())
	if minBond < 0 || err != nil {
		minBond = constAccessor.GetInt64Value(constants.MinimumBondInRune)
	}
	if na.Bond.LT(sdk.NewUint(uint64(minBond))) {
		return nil
	}

	// figure out if THORNode need to send them assets.
	// get a list of coin/amounts this yggdrasil pool should have, ideally.
	// TODO: We are assuming here that the pub key is Secp256K1
	ygg, err := keeper.GetVault(ctx, na.PubKeySet.Secp256k1)
	if err != nil {
		if !errors.Is(err, ErrVaultNotFound) {
			return fmt.Errorf("fail to get yggdrasil: %w", err)
		}
		ygg = NewVault(ctx.BlockHeight(), ActiveVault, YggdrasilVault, na.PubKeySet.Secp256k1, nil)
		ygg.Membership = append(ygg.Membership, na.PubKeySet.Secp256k1)

		if err := keeper.SetVault(ctx, ygg); err != nil {
			return fmt.Errorf("fail to create yggdrasil pool: %w", err)
		}
	}
	if !ygg.IsYggdrasil() {
		return nil
	}
	pendingTxCount := ygg.LenPendingTxBlockHeights(ctx.BlockHeight(), constAccessor)
	if pendingTxCount > 0 {
		return fmt.Errorf("cannot send more yggdrasil funds while transactions are pending (%s: %d)", ygg.PubKey, pendingTxCount)
	}

	// calculate the total value of funds of this yggdrasil vault
	totalValue := sdk.ZeroUint()
	for _, coin := range ygg.Coins {
		if coin.Asset.IsRune() {
			totalValue = totalValue.Add(coin.Amount)
			continue
		}
		for _, pool := range pools {
			if pool.Asset.Equals(coin.Asset) {
				totalValue = totalValue.Add(pool.AssetValueInRune(coin.Amount))
			}
		}
	}

	// if the ygg total value is more than 25% bond, funds are low enough yet
	// to top up
	if totalValue.MulUint64(4).GTE(na.Bond) {
		return nil
	}

	targetCoins, err := calcTargetYggCoins(pools, ygg, na.Bond, totalBond)
	if err != nil {
		return err
	}

	var sendCoins common.Coins
	// iterate over each target coin amount and figure if THORNode need to reimburse
	// a Ygg pool of this particular asset.
	for _, targetCoin := range targetCoins {
		yggCoin := ygg.GetCoin(targetCoin.Asset)
		// check if the amount the ygg pool has is less that 50% of what
		// they are suppose to have, ideally. We refill them if they drop
		// below this line
		if yggCoin.Amount.LT(targetCoin.Amount.QuoUint64(2)) {
			sendCoins = append(
				sendCoins,
				common.NewCoin(
					targetCoin.Asset,
					common.SafeSub(targetCoin.Amount, yggCoin.Amount),
				),
			)
		}

	}

	if len(sendCoins) > 0 {
		count, err := sendCoinsToYggdrasil(ctx, keeper, sendCoins, ygg, txOutStore)
		if err != nil {
			return err
		}
		for i := 0; i < count; i++ {
			ygg.AppendPendingTxBlockHeights(ctx.BlockHeight(), constAccessor)
		}
		if err := keeper.SetVault(ctx, ygg); err != nil {
			return fmt.Errorf("fail to create yggdrasil pool: %w", err)
		}
	}

	return nil
}

// sendCoinsToYggdrasil - adds outbound txs to send the given coins to a
// yggdrasil pool
func sendCoinsToYggdrasil(ctx sdk.Context, keeper Keeper, coins common.Coins, ygg Vault, txOutStore TxOutStore) (int, error) {
	var count int

	active, err := keeper.GetAsgardVaultsByStatus(ctx, ActiveVault)
	if err != nil {
		return count, err
	}

	for _, coin := range coins {

		// select active vault to send funds from
		vault := active.SelectByMaxCoin(coin.Asset)
		if vault.IsEmpty() {
			continue
		}
		if coin.Amount.GT(vault.GetCoin(coin.Asset).Amount) {
			// not enough funds
			continue
		}

		to, err := ygg.PubKey.GetAddress(coin.Asset.Chain)
		if err != nil {
			ctx.Logger().Error("fail to get address for pubkey", "pubkey", ygg.PubKey, "chain", coin.Asset.Chain, "error", err)
			continue
		}

		toi := &TxOutItem{
			Chain:       coin.Asset.Chain,
			ToAddress:   to,
			InHash:      common.BlankTxID,
			Memo:        NewYggdrasilFund(ctx.BlockHeight()).String(),
			Coin:        coin,
			VaultPubKey: vault.PubKey,
		}
		if err := txOutStore.UnSafeAddTxOutItem(ctx, toi); err != nil {
			return count, err
		}
		count += 1
	}

	return count, nil
}

// calcTargetYggCoins - calculate the amount of coins of each pool a yggdrasil
// pool should have, relative to how much they have bonded (which should be
// target == bond / 2).
func calcTargetYggCoins(pools []Pool, ygg Vault, yggBond, totalBond sdk.Uint) (common.Coins, error) {
	runeCoin := common.NewCoin(common.RuneAsset(), sdk.ZeroUint())
	var coins common.Coins

	// calculate total rune in our pools
	totalRune := sdk.ZeroUint()
	for _, pool := range pools {
		totalRune = totalRune.Add(pool.BalanceRune)
	}
	if totalRune.IsZero() {
		// if nothing is staked, no coins should be issued
		return nil, nil
	}

	// figure out what percentage of the bond this yggdrasil pool has. They
	// should get half of that value.
	targetRune := common.GetShare(yggBond, totalBond.Mul(sdk.NewUint(2)), totalRune)
	// check if more rune would be allocated to this pool than their bond allows
	if targetRune.GT(yggBond.QuoUint64(2)) {
		targetRune = yggBond.QuoUint64(2)
	}

	// track how much value (in rune) we've associated with this ygg pool. This
	// is here just to be absolutely sure THORNode never send too many assets to the
	// ygg by accident.
	counter := sdk.ZeroUint()
	for _, pool := range pools {
		runeAmt := common.GetShare(targetRune, totalRune, pool.BalanceRune)
		runeCoin.Amount = runeCoin.Amount.Add(runeAmt)
		assetAmt := common.GetShare(targetRune, totalRune, pool.BalanceAsset)
		// add rune amt (not asset since the two are considered to be equal)
		// in a single pool X, the value of 1% asset X in RUNE ,equals the 1% RUNE in the same pool
		yggCoin := ygg.GetCoin(pool.Asset)
		coin := common.NewCoin(pool.Asset, common.SafeSub(assetAmt, yggCoin.Amount))
		if !coin.IsEmpty() {
			counter = counter.Add(runeAmt)
			coins = append(coins, coin)
		}
	}

	yggRune := ygg.GetCoin(common.RuneAsset())
	runeCoin.Amount = common.SafeSub(runeCoin.Amount, yggRune.Amount)
	if !runeCoin.IsEmpty() {
		counter = counter.Add(runeCoin.Amount)
		coins = append(coins, runeCoin)
	}

	// ensure THORNode don't send too much value in coins to the ygg pool
	if counter.GT(yggBond.QuoUint64(2)) {
		return nil, fmt.Errorf("exceeded safe amounts of assets for given Yggdrasil pool (%d/%d)", counter.Uint64(), yggBond.QuoUint64(2).Uint64())
	}

	return coins, nil
}
