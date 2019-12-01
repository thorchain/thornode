package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
)

func refundTx(ctx sdk.Context, txID common.TxID, tx TxIn, store *TxOutStore, keeper Keeper, poolAddr common.PubKey, chain common.Chain, deductFee bool) {
	// If we recognize one of the coins, and therefore able to refund
	// withholding fees, refund all coins.
	for _, coin := range tx.Coins {
		pool := keeper.GetPool(ctx, coin.Asset)
		if coin.Asset.IsRune() || !pool.BalanceRune.IsZero() {
			toi := &TxOutItem{
				Chain:       chain,
				InHash:      txID,
				ToAddress:   tx.Sender,
				PoolAddress: poolAddr,
				Coin:        coin,
			}
			store.AddTxOutItem(ctx, keeper, toi, false)
			continue
		}

		// Since we have assets, we don't have a pool for, we don't know how to
		// refund and withhold for fees. Instead, we'll create a pool with the
		// amount of assets, and associate them with no stakers (meaning up for
		// grabs). This could be like an airdrop scenario, for example.
		// Don't assume this is the first time we've seen this coin (ie second
		// airdrop).
		pool.BalanceAsset = pool.BalanceAsset.Add(coin.Amount)
		pool.Asset = coin.Asset
		if pool.BalanceRune.IsZero() && pool.Status != PoolBootstrap {
			pool.Status = PoolBootstrap
			eventPoolStatusWrapper(ctx, keeper, pool)
		}
		keeper.SetPool(ctx, pool)
	}
}

func RefundBond(ctx sdk.Context, txID common.TxID, nodeAcc NodeAccount, keeper Keeper, txOut *TxOutStore) {
	if nodeAcc.Bond.GT(sdk.ZeroUint()) {
		// refund bond
		txOutItem := &TxOutItem{
			Chain:     common.BNBChain,
			ToAddress: nodeAcc.BondAddress,
			InHash:    txID,
			Coin:      common.NewCoin(common.RuneAsset(), nodeAcc.Bond),
		}

		txOut.AddTxOutItem(ctx, keeper, txOutItem, true)
	}

	nodeAcc.Bond = sdk.ZeroUint()
	// disable the node account
	nodeAcc.UpdateStatus(NodeDisabled, ctx.BlockHeight())
	keeper.SetNodeAccount(ctx, nodeAcc)
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
			ctx.Logger().Error("unauthorized account", "address", signer.String())
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

func completeEvents(ctx sdk.Context, keeper Keeper, txID common.TxID, txs common.Txs) error {
	lastEventID, err := keeper.GetLastEventID(ctx)
	if err != nil {
		return err
	}
	incomplete, err := keeper.GetIncompleteEvents(ctx)
	if err != nil {
		return err
	}
	todo, incomplete := incomplete.PopByInHash(txID)
	for _, evt := range todo {
		lastEventID += 1
		evt.ID = lastEventID
		evt.OutTxs = txs
		keeper.SetCompletedEvent(ctx, evt)
	}
	keeper.SetIncompleteEvents(ctx, incomplete)
	keeper.SetLastEventID(ctx, lastEventID)
	return nil
}
