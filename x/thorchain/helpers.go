package thorchain

import (
	"encoding/json"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

func refundTx(ctx sdk.Context, tx ObservedTx, store TxOutStore, keeper Keeper, constAccessor constants.ConstantValues, refundCode sdk.CodeType, refundReason string, eventMgr EventManager) error {
	// If THORNode recognize one of the coins, and therefore able to refund
	// withholding fees, refund all coins.
	var refundCoins common.Coins
	for _, coin := range tx.Tx.Coins {
		pool, err := keeper.GetPool(ctx, coin.Asset)
		if err != nil {
			return fmt.Errorf("fail to get pool: %w", err)
		}

		if coin.Asset.IsRune() || !pool.BalanceRune.IsZero() {
			toi := &TxOutItem{
				Chain:       tx.Tx.Chain,
				InHash:      tx.Tx.ID,
				ToAddress:   tx.Tx.FromAddress,
				VaultPubKey: tx.ObservedPubKey,
				Coin:        coin,
				Memo:        NewRefundMemo(tx.Tx.ID).String(),
			}

			success, err := store.TryAddTxOutItem(ctx, toi)
			if err != nil {
				return fmt.Errorf("fail to prepare outbund tx: %w", err)
			}
			if success {
				refundCoins = append(refundCoins, toi.Coin)
			}
		}
		// Zombie coins are just dropped.
	}
	eventRefund := NewEventRefund(refundCode, refundReason, tx.Tx, common.NewFee(common.Coins{}, sdk.ZeroUint()))
	status := EventSuccess
	if len(refundCoins) > 0 {
		// create a new TX based on the coins thorchain refund , some of the coins thorchain doesn't refund
		// coin thorchain doesn't have pool with , likely airdrop
		newTx := common.NewTx(tx.Tx.ID, tx.Tx.FromAddress, tx.Tx.ToAddress, tx.Tx.Coins, tx.Tx.Gas, tx.Tx.Memo)
		transactionFee := constAccessor.GetInt64Value(constants.TransactionFee)
		fee := getFee(tx.Tx.Coins, refundCoins, transactionFee)
		eventRefund = NewEventRefund(refundCode, refundReason, newTx, fee)
		status = EventPending

	}
	if err := eventMgr.EmitRefundEvent(ctx, keeper, eventRefund, status); err != nil {
		return fmt.Errorf("fail to emit refund event: %w", err)
	}
	return nil
}

func getFee(input, output common.Coins, transactionFee int64) common.Fee {
	var fee common.Fee
	assetTxCount := 0
	for _, out := range output {
		if !out.Asset.IsRune() {
			assetTxCount++
		}
	}
	for _, in := range input {
		outCoin := common.NoCoin
		for _, out := range output {
			if out.Asset.Equals(in.Asset) {
				outCoin = out
				break
			}
		}
		if outCoin.IsEmpty() {
			fee.Coins = append(fee.Coins, common.NewCoin(in.Asset, in.Amount))
		} else {
			fee.Coins = append(fee.Coins, common.NewCoin(in.Asset, in.Amount.Sub(outCoin.Amount)))
		}
	}
	fee.PoolDeduct = sdk.NewUint(uint64(transactionFee) * uint64(assetTxCount))
	return fee
}

func subsidizePoolWithSlashBond(ctx sdk.Context, keeper Keeper, ygg Vault, yggTotalStolen, slashRuneAmt sdk.Uint) error {
	// Thorchain did not slash the node account
	if slashRuneAmt.IsZero() {
		return nil
	}
	stolenRUNE := ygg.GetCoin(common.RuneAsset()).Amount
	slashRuneAmt = common.SafeSub(slashRuneAmt, stolenRUNE)
	yggTotalStolen = common.SafeSub(yggTotalStolen, stolenRUNE)
	type fund struct {
		stolenAsset   sdk.Uint
		subsidiseRune sdk.Uint
	}
	// here need to use a map to hold on to the amount of RUNE need to be subsidized to each pool
	// reason being , if ygg pool has both RUNE and BNB coin left, these two coin share the same pool
	// which is BNB pool , if add the RUNE directly back to pool , it will affect BNB price , which will affect the result
	subsidizeAmounts := make(map[common.Asset]fund)
	for _, coin := range ygg.Coins {
		asset := coin.Asset
		if coin.Asset.IsRune() {
			// when the asset is RUNE, thorchain don't need to update the RUNE balance on pool
			continue
		}
		f, ok := subsidizeAmounts[asset]
		if !ok {
			f = fund{
				stolenAsset:   sdk.ZeroUint(),
				subsidiseRune: sdk.ZeroUint(),
			}
		}

		pool, err := keeper.GetPool(ctx, asset)
		if err != nil {
			return err
		}
		f.stolenAsset = f.stolenAsset.Add(coin.Amount)
		runeValue := pool.AssetValueInRune(coin.Amount)
		// the amount of RUNE thorchain used to subsidize the pool is calculate by ratio
		// slashRune * (stealAssetRuneValue /totalStealAssetRuneValue)
		subsidizeAmt := slashRuneAmt.Mul(runeValue).Quo(yggTotalStolen)
		f.subsidiseRune = f.subsidiseRune.Add(subsidizeAmt)
		subsidizeAmounts[asset] = f
	}

	for asset, f := range subsidizeAmounts {
		pool, err := keeper.GetPool(ctx, asset)
		if err != nil {
			return err
		}
		pool.BalanceRune = pool.BalanceRune.Add(f.subsidiseRune)
		pool.BalanceAsset = common.SafeSub(pool.BalanceAsset, f.stolenAsset)

		if err := keeper.SetPool(ctx, pool); err != nil {
			return fmt.Errorf("fail to save pool: %w", err)
		}
	}
	return nil
}

// getTotalYggValueInRune will go through all the coins in ygg , and calculate the total value in RUNE
// return value will be totalValueInRune,error
func getTotalYggValueInRune(ctx sdk.Context, keeper Keeper, ygg Vault) (sdk.Uint, error) {
	yggRune := sdk.ZeroUint()
	for _, coin := range ygg.Coins {
		if coin.Asset.IsRune() {
			yggRune = yggRune.Add(coin.Amount)
		} else {
			pool, err := keeper.GetPool(ctx, coin.Asset)
			if err != nil {
				return sdk.ZeroUint(), err
			}
			yggRune = yggRune.Add(pool.AssetValueInRune(coin.Amount))
		}
	}
	return yggRune, nil
}

func refundBond(ctx sdk.Context, tx common.Tx, nodeAcc NodeAccount, keeper Keeper, txOut TxOutStore, eventMgr EventManager) error {
	if nodeAcc.Status == NodeActive {
		ctx.Logger().Info("node still active , cannot refund bond", "node address", nodeAcc.NodeAddress, "node pub key", nodeAcc.PubKeySet.Secp256k1)
		return nil
	}

	ygg := Vault{}
	if keeper.VaultExists(ctx, nodeAcc.PubKeySet.Secp256k1) {
		var err error
		ygg, err = keeper.GetVault(ctx, nodeAcc.PubKeySet.Secp256k1)
		if err != nil {
			return err
		}
		if !ygg.IsYggdrasil() {
			return errors.New("this is not a Yggdrasil vault")
		}
	}

	// Calculate total value (in rune) the Yggdrasil pool has
	yggRune, err := getTotalYggValueInRune(ctx, keeper, ygg)
	if err != nil {
		return fmt.Errorf("fail to get total ygg value in RUNE: %w", err)
	}

	if nodeAcc.Bond.LT(yggRune) {
		ctx.Logger().Error(fmt.Sprintf("Node Account (%s) left with more funds in their Yggdrasil vault than their bond's value (%s / %s)", nodeAcc.NodeAddress, yggRune, nodeAcc.Bond))
	}
	// slashing 1.5 * yggdrasil remains
	slashRune := yggRune.MulUint64(3).QuoUint64(2)
	bondBeforeSlash := nodeAcc.Bond
	nodeAcc.Bond = common.SafeSub(nodeAcc.Bond, slashRune)

	if !nodeAcc.Bond.IsZero() {
		active, err := keeper.GetAsgardVaultsByStatus(ctx, ActiveVault)
		if err != nil {
			ctx.Logger().Error("fail to get active vaults", "error", err)
			return err
		}

		vault := active.SelectByMinCoin(common.RuneAsset())
		if vault.IsEmpty() {
			return fmt.Errorf("unable to determine asgard vault to send funds")
		}

		bondEvent := NewEventBond(nodeAcc.Bond, BondReturned, tx)
		if err := eventMgr.EmitBondEvent(ctx, keeper, bondEvent); err != nil {
			return fmt.Errorf("fail to emit bond event: %w", err)
		}

		refundAddress := nodeAcc.BondAddress
		if common.RuneAsset().Chain.Equals(common.THORChain) {
			refundAddress = common.Address(nodeAcc.NodeAddress.String())
		}

		// refund bond
		txOutItem := &TxOutItem{
			Chain:       common.RuneAsset().Chain,
			ToAddress:   refundAddress,
			VaultPubKey: vault.PubKey,
			InHash:      tx.ID,
			Coin:        common.NewCoin(common.RuneAsset(), nodeAcc.Bond),
		}
		_, err = txOut.TryAddTxOutItem(ctx, txOutItem)
		if err != nil {
			return fmt.Errorf("fail to add outbound tx: %w", err)
		}
	} else {
		// if it get into here that means the node account doesn't have any bond left after slash.
		// which means the real slashed RUNE could be the bond they have before slash
		slashRune = bondBeforeSlash
	}

	nodeAcc.Bond = sdk.ZeroUint()
	// disable the node account
	nodeAcc.UpdateStatus(NodeDisabled, ctx.BlockHeight())
	if err := keeper.SetNodeAccount(ctx, nodeAcc); err != nil {
		ctx.Logger().Error(fmt.Sprintf("fail to save node account(%s)", nodeAcc), "error", err)
		return err
	}
	if err := subsidizePoolWithSlashBond(ctx, keeper, ygg, yggRune, slashRune); err != nil {
		ctx.Logger().Error("fail to subsidize pool with slashed bond", "error", err)
		return err
	}
	// delete the ygg vault, there is nothing left in the ygg vault
	if !ygg.HasFunds() {
		return keeper.DeleteVault(ctx, ygg.PubKey)
	}
	return nil
}

// Checks if the observed vault pubkey is a valid asgard or ygg vault
func isCurrentVaultPubKey(ctx sdk.Context, keeper Keeper, tx ObservedTx) bool {
	return keeper.VaultExists(ctx, tx.ObservedPubKey)
}

func isSignedByActiveNodeAccounts(ctx sdk.Context, keeper Keeper, signers []sdk.AccAddress) bool {
	if len(signers) == 0 {
		return false
	}
	supplier := keeper.Supply()
	for _, signer := range signers {
		if signer.Equals(supplier.GetModuleAddress(AsgardName)) {
			continue
		}
		nodeAccount, err := keeper.GetNodeAccount(ctx, signer)
		if err != nil {
			ctx.Logger().Error("unauthorized account", "address", signer.String(), "error", err)
			return false
		}
		if nodeAccount.IsEmpty() {
			ctx.Logger().Error("unauthorized account", "address", signer.String())
			return false
		}
		if nodeAccount.Status != NodeActive {
			ctx.Logger().Error("unauthorized account, node account not active", "address", signer.String(), "status", nodeAccount.Status)
			return false
		}
	}
	return true
}

func updateEventStatus(ctx sdk.Context, keeper Keeper, eventID int64, txs common.Txs, eventStatus EventStatus) error {
	event, err := keeper.GetEvent(ctx, eventID)
	if err != nil {
		return fmt.Errorf("fail to get event: %w", err)
	}

	// if the event is already successful, don't append more transactions
	if event.Status == EventSuccess {
		return nil
	}

	ctx.Logger().Info(fmt.Sprintf("set event to %s,eventID (%d) , txs:%s", eventStatus, eventID, txs))
	outTxs := append(event.OutTxs, txs...)
	for i := 0; i < len(outTxs); i++ {
		duplicate := false
		for j := i + 1; j < len(outTxs); j++ {
			if outTxs[i].Equals(outTxs[j]) {
				duplicate = true
			}
		}
		if !duplicate {
			event.OutTxs = append(event.OutTxs, outTxs[i])
		}
	}
	if eventStatus == RefundStatus {
		// we need to check we refunded all the coins that need to be refunded from in tx
		// before updating status to complete, we use the count of voter actions to check
		voter, err := keeper.GetObservedTxVoter(ctx, event.InTx.ID)
		if err != nil {
			return fmt.Errorf("fail to get observed tx voter: %w", err)
		}
		if len(voter.Actions) == len(event.OutTxs) {
			event.Status = eventStatus
		}
	} else {
		event.Status = eventStatus
	}
	return keeper.UpsertEvent(ctx, event)
}

func updateEventFee(ctx sdk.Context, keeper Keeper, txID common.TxID, fee common.Fee) error {
	ctx.Logger().Info("update event fee txid", "tx", txID.String())
	eventIDs, err := keeper.GetEventsIDByTxHash(ctx, txID)
	if err != nil {
		if err == ErrEventNotFound {
			ctx.Logger().Error(fmt.Sprintf("could not find the event(%s)", txID))
			return nil
		}
		return fmt.Errorf("fail to get event id: %w", err)
	}
	if len(eventIDs) == 0 {
		return errors.New("no event found")
	}
	// There are two events for double swap with the same the same txID. Only the second one has fee
	eventID := eventIDs[len(eventIDs)-1]
	event, err := keeper.GetEvent(ctx, eventID)
	if err != nil {
		return fmt.Errorf("fail to get event: %w", err)
	}

	ctx.Logger().Info(fmt.Sprintf("Update fee for event %d, fee:%s", eventID, fee))
	event.Fee.Coins = append(event.Fee.Coins, fee.Coins...)
	event.Fee.PoolDeduct = event.Fee.PoolDeduct.Add(fee.PoolDeduct)
	return keeper.UpsertEvent(ctx, event)
}

func completeEvents(ctx sdk.Context, keeper Keeper, txID common.TxID, txs common.Txs, eventStatus EventStatus) error {
	ctx.Logger().Info(fmt.Sprintf("txid(%s)", txID))
	eventIDs, err := keeper.GetPendingEventID(ctx, txID)
	if err != nil {
		if err == ErrEventNotFound {
			ctx.Logger().Error(fmt.Sprintf("could not find the event(%s)", txID))
			return nil
		}
		return fmt.Errorf("fail to get pending event id: %w", err)
	}
	for _, item := range eventIDs {
		if err := updateEventStatus(ctx, keeper, item, txs, eventStatus); err != nil {
			return fmt.Errorf("fail to set event(%d) to %s: %w", item, eventStatus, err)
		}
	}
	return nil
}

func enableNextPool(ctx sdk.Context, keeper Keeper, eventManager EventManager) error {
	var pools []Pool
	iterator := keeper.GetPoolIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var pool Pool
		if err := keeper.Cdc().UnmarshalBinaryBare(iterator.Value(), &pool); err != nil {
			return err
		}

		if pool.Status == PoolBootstrap && !pool.BalanceAsset.IsZero() && !pool.BalanceRune.IsZero() {
			pools = append(pools, pool)
		}
	}

	if len(pools) == 0 {
		return nil
	}

	pool := pools[0]
	for _, p := range pools {
		// find the pool that has most RUNE, also exclude those pool that doesn't have asset
		if pool.BalanceRune.LT(p.BalanceRune) {
			pool = p
		}
	}

	poolEvt := NewEventPool(pool.Asset, PoolEnabled)
	if err := eventManager.EmitPoolEvent(ctx, keeper, common.BlankTxID, EventSuccess, poolEvt); err != nil {
		return fmt.Errorf("fail to emit pool event: %w", err)
	}

	pool.Status = PoolEnabled
	return keeper.SetPool(ctx, pool)
}

func wrapError(ctx sdk.Context, err error, wrap string) error {
	err = fmt.Errorf("%s: %w", wrap, err)
	ctx.Logger().Error(err.Error())
	return err
}

func AddGasFees(ctx sdk.Context, keeper Keeper, tx ObservedTx, gasManager GasManager) error {
	if len(tx.Tx.Gas) == 0 {
		return nil
	}

	// update state with new gas info
	if len(tx.Tx.Coins) > 0 {
		gasAsset := tx.Tx.Coins[0].Asset.Chain.GetGasAsset()
		gasInfo, err := keeper.GetGas(ctx, gasAsset)
		if err == nil {
			gasInfo = common.UpdateGasPrice(tx.Tx, gasAsset, gasInfo)
			if gasInfo != nil {
				keeper.SetGas(ctx, gasAsset, gasInfo)
			} else {
				ctx.Logger().Error(fmt.Sprintf("fail to update gas price for chain: %s", gasAsset))
			}
		}
	}

	gasManager.AddGasAsset(tx.Tx.Gas)

	// Subtract from the vault
	if keeper.VaultExists(ctx, tx.ObservedPubKey) {
		vault, err := keeper.GetVault(ctx, tx.ObservedPubKey)
		if err != nil {
			return err
		}

		vault.SubFunds(tx.Tx.Gas.ToCoins())

		if err := keeper.SetVault(ctx, vault); err != nil {
			return err
		}
	}
	return nil
}

func getErrMessageFromABCILog(content string) (string, error) {
	var humanReadableError struct {
		Codespace sdk.CodespaceType `json:"codespace"`
		Code      sdk.CodeType      `json:"code"`
		Message   string            `json:"message"`
	}
	if err := json.Unmarshal([]byte(content), &humanReadableError); err != nil {
		return "", err
	}
	return humanReadableError.Message, nil
}
