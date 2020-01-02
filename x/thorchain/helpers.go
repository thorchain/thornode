package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"gitlab.com/thorchain/thornode/common"
)

func refundTx(ctx sdk.Context, tx ObservedTx, store TxOutStore, keeper Keeper, deductFee bool) error {
	// If THORNode recognize one of the coins, and therefore able to refund
	// withholding fees, refund all coins.
	for _, coin := range tx.Tx.Coins {
		pool, err := keeper.GetPool(ctx, coin.Asset)
		if err != nil {
			return fmt.Errorf("fail to get pool: %s", err)
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
			store.AddTxOutItem(ctx, toi)

		}

		// Zombie coins are just dropped.
	}
	// emit refund event
	ev := NewEvent("refund", ctx.BlockHeight(), tx.Tx, nil, EventPending)
	if err := keeper.UpsertEvent(ctx, ev); err != nil {
		return fmt.Errorf("fail to write refund event")
	}

	return nil
}

func refundBond(ctx sdk.Context, txID common.TxID, nodeAcc NodeAccount, keeper Keeper, txOut TxOutStore) error {
	ygg, err := keeper.GetVault(ctx, nodeAcc.PubKeySet.Secp256k1)
	if err != nil {
		return err
	}
	if !ygg.IsYggdrasil() {
		return fmt.Errorf("this is not a Yggdrasil vault")
	}

	// Calculate total value (in rune) the Yggdrasil pool has
	yggRune := sdk.ZeroUint()
	for _, coin := range ygg.Coins {
		if coin.Asset.IsRune() {
			yggRune = yggRune.Add(coin.Amount)
		} else {
			pool, err := keeper.GetPool(ctx, coin.Asset)
			if err != nil {
				return err
			}
			yggRune = yggRune.Add(pool.AssetValueInRune(coin.Amount))
		}
	}

	if nodeAcc.Bond.LT(yggRune) {
		ctx.Logger().Error("Node Account (%s) left with more funds in their Yggdrasil vault than their bond's value (%d/%d)", yggRune, nodeAcc.Bond)
	}

	nodeAcc.Bond = common.SafeSub(nodeAcc.Bond, yggRune)

	if nodeAcc.Bond.GT(sdk.ZeroUint()) {

		active, err := keeper.GetAsgardVaultsByStatus(ctx, ActiveVault)
		if err != nil {
			ctx.Logger().Error("fail to get active vaults", err)
			return err
		}

		vault := active.SelectByMinCoin(common.RuneAsset())
		if vault.IsEmpty() {
			return fmt.Errorf("unable to determine asgard vault to send funds")
		}

		// refund bond
		txOutItem := &TxOutItem{
			Chain:       common.BNBChain,
			ToAddress:   nodeAcc.BondAddress,
			VaultPubKey: vault.PubKey,
			InHash:      txID,
			Coin:        common.NewCoin(common.RuneAsset(), nodeAcc.Bond),
		}

		txOut.AddTxOutItem(ctx, txOutItem)
	}

	nodeAcc.Bond = sdk.ZeroUint()
	// disable the node account
	nodeAcc.UpdateStatus(NodeDisabled, ctx.BlockHeight())
	if err := keeper.SetNodeAccount(ctx, nodeAcc); nil != err {
		ctx.Logger().Error(fmt.Sprintf("fail to save node account(%s)", nodeAcc), err)
		return err
	}

	return nil
}

// Checks if the observed vault pubkey is a valid asgard or ygg vault
func isCurrentVaultPubKey(ctx sdk.Context, keeper Keeper, tx ObservedTx) bool {
	return keeper.VaultExists(ctx, tx.ObservedPubKey)
}

// isSignedByActiveObserver check whether the signers are all active observer
func isSignedByActiveObserver(ctx sdk.Context, keeper Keeper, signers []sdk.AccAddress) bool {
	if len(signers) == 0 {
		return false
	}
	for _, signer := range signers {
		if !keeper.IsActiveObserver(ctx, signer) {
			return false
		}
	}
	return true
}

func isSignedByActiveNodeAccounts(ctx sdk.Context, keeper Keeper, signers []sdk.AccAddress) bool {
	if len(signers) == 0 {
		return false
	}
	for _, signer := range signers {
		nodeAccount, err := keeper.GetNodeAccount(ctx, signer)
		if err != nil {
			ctx.Logger().Error("unauthorized account", "address", signer.String(), err)
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

func completeEventsByID(ctx sdk.Context, keeper Keeper, eventID int64, txs common.Txs, eventStatus EventStatus) error {
	event, err := keeper.GetEvent(ctx, eventID)
	if nil != err {
		return fmt.Errorf("fail to get event: %w", err)
	}
	ctx.Logger().Info(fmt.Sprintf("set event to %s,eventID (%d) , txs:%s", eventStatus, eventID, txs))
	event.Status = eventStatus
	event.OutTxs = txs
	return keeper.UpsertEvent(ctx, event)
}

func completeEvents(ctx sdk.Context, keeper Keeper, txID common.TxID, txs common.Txs, eventStatus EventStatus) error {
	ctx.Logger().Info(fmt.Sprintf("txid(%s)", txID))
	eventIDs, err := keeper.GetPendingEventID(ctx, txID)
	if nil != err {
		if err == ErrEventNotFound {
			ctx.Logger().Error(fmt.Sprintf("could not find the event(%s)", txID))
			return nil
		}
		return fmt.Errorf("fail to get pending event id: %w", err)
	}
	for _, item := range eventIDs {
		if err := completeEventsByID(ctx, keeper, item, txs, eventStatus); nil != err {
			return fmt.Errorf("fail to set event(%d) to %s: %w", item, eventStatus, err)
		}
	}
	return nil
}

func enableNextPool(ctx sdk.Context, keeper Keeper) error {
	var pools []Pool
	iterator := keeper.GetPoolIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var pool Pool
		if err := keeper.Cdc().UnmarshalBinaryBare(iterator.Value(), &pool); err != nil {
			return err
		}
		if pool.Status == PoolBootstrap {
			pools = append(pools, pool)
		}
	}

	if len(pools) > 0 {
		pool := pools[0]
		for _, p := range pools {
			if pool.BalanceRune.LT(p.BalanceRune) {
				pool = p
			}
		}
		// ensure THORNode don't enable a pool that doesn't have any rune or assets
		if pool.BalanceAsset.IsZero() || pool.BalanceRune.IsZero() {
			return nil
		}
		pool.Status = PoolEnabled
		return keeper.SetPool(ctx, pool)
	}
	return nil
}

func wrapError(ctx sdk.Context, err error, wrap string) error {
	err = errors.Wrap(err, wrap)
	ctx.Logger().Error(err.Error())
	return err
}

func AddGasFees(ctx sdk.Context, keeper Keeper, tx ObservedTx) error {
	if len(tx.Tx.Gas) == 0 {
		return nil
	}

	vault, err := keeper.GetVaultData(ctx)
	if nil != err {
		return fmt.Errorf("fail to get vault: %w", err)
	}
	vault.Gas = vault.Gas.Add(tx.Tx.Gas)
	if err := keeper.SetVaultData(ctx, vault); err != nil {
		return err
	}

	// Subtract gas from pools (will be reimbursed later with rune at the end
	// of the block)
	for _, gas := range tx.Tx.Gas {
		pool, err := keeper.GetPool(ctx, gas.Asset)
		if err != nil {
			return err
		}
		pool.BalanceAsset = common.SafeSub(pool.BalanceAsset, gas.Amount)
		if err := keeper.SetPool(ctx, pool); err != nil {
			return err
		}
	}

	if keeper.VaultExists(ctx, tx.ObservedPubKey) {
		ygg, err := keeper.GetVault(ctx, tx.ObservedPubKey)
		if err != nil {
			return err
		}

		ygg.SubFunds(tx.Tx.Gas.ToCoins())

		return keeper.SetVault(ctx, ygg)
	}

	return nil
}
