package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"gitlab.com/thorchain/thornode/common"
)

func refundTx(ctx sdk.Context, tx ObservedTx, store *TxOutStore, keeper Keeper, poolAddr common.PubKey, chain common.Chain, deductFee bool) error {
	// If THORNode recognize one of the coins, and therefore able to refund
	// withholding fees, refund all coins.
	for _, coin := range tx.Tx.Coins {
		pool, err := keeper.GetPool(ctx, coin.Asset)
		if err != nil {
			return fmt.Errorf("fail to get pool: %s", err)
		}
		if coin.Asset.IsRune() || !pool.BalanceRune.IsZero() {
			toi := &TxOutItem{
				Chain:       chain,
				InHash:      tx.Tx.ID,
				ToAddress:   tx.Tx.FromAddress,
				PoolAddress: poolAddr,
				Coin:        coin,
			}
			store.AddTxOutItem(ctx, keeper, toi, false)
			continue
		}

		// Zombie coins are just dropped.
	}
	return nil
}

// RefundBond use to return validator's bond
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
	if err := keeper.SetNodeAccount(ctx, nodeAcc); nil != err {
		ctx.Logger().Error(fmt.Sprintf("fail to save node account(%s)", nodeAcc), err)
	}
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
		lastEventID++
		evt.ID = lastEventID
		evt.OutTxs = txs
		keeper.SetCompletedEvent(ctx, evt)
	}
	keeper.SetIncompleteEvents(ctx, incomplete)
	keeper.SetLastEventID(ctx, lastEventID)
	return nil
}

func enableNextPool(ctx sdk.Context, keeper Keeper) error {
	var pools []Pool
	iterator := keeper.GetPoolIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var pool Pool
		keeper.Cdc().MustUnmarshalBinaryBare(iterator.Value(), &pool)
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
