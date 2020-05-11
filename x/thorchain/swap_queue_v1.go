package thorchain

import (
	"fmt"
	"sort"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// SwapQv1 is going to manage the swaps queue
type SwapQv1 struct {
	k                     Keeper
	versionedTxOutStore   VersionedTxOutStore
	versionedEventManager VersionedEventManager
}

type swapItem struct {
	msg  MsgSwap
	fee  sdk.Uint
	slip sdk.Uint
}
type swapItems []swapItem

// NewSwapQv1 create a new vault manager
func NewSwapQv1(k Keeper, versionedTxOutStore VersionedTxOutStore, versionedEventManager VersionedEventManager) *SwapQv1 {
	return &SwapQv1{
		k:                     k,
		versionedTxOutStore:   versionedTxOutStore,
		versionedEventManager: versionedEventManager,
	}
}

// FetchQueue - grabs all swap queue items from the kvstore and returns them
func (vm *SwapQv1) FetchQueue(ctx sdk.Context) ([]MsgSwap, error) {
	msgs := make([]MsgSwap, 0)
	iterator := vm.k.GetSwapQueueIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var msg MsgSwap
		if err := vm.k.Cdc().UnmarshalBinaryBare(iterator.Value(), &msg); err != nil {
			return msgs, err
		}
		msgs = append(msgs, msg)
	}

	return msgs, nil
}

// EndBlock move funds from retiring asgard vaults
func (vm *SwapQv1) EndBlock(ctx sdk.Context, version semver.Version, constAccessor constants.ConstantValues) error {
	handler := NewSwapHandler(vm.k, vm.versionedTxOutStore, vm.versionedEventManager)

	txOutStore, err := vm.versionedTxOutStore.GetTxOutStore(ctx, vm.k, version)
	if err != nil {
		ctx.Logger().Error("fail to get txout store", "error", err)
		return err
	}
	eventMgr, err := vm.versionedEventManager.GetEventManager(ctx, version)
	if err != nil {
		ctx.Logger().Error("fail to get event manager", "error", err)
		return fmt.Errorf("fail to get event manager: %w", err)
	}
	msgs, err := vm.FetchQueue(ctx)
	if err != nil {
		ctx.Logger().Error("fail to fetch swap queue from store", "error", err)
		return err
	}

	swaps, err := vm.ScoreMsgs(ctx, msgs)
	if err != nil {
		ctx.Logger().Error("fail to fetch swap items", "error", err)
		// continue, don't exit, just do them out of order (instead of not
		// at all)
	}
	swaps = swaps.Sort()

	for i := 0; i < vm.getTodoNum(len(swaps)); i++ {
		pick := swaps[i]

		result := handler.handle(ctx, pick.msg, version, constAccessor)
		if !result.IsOK() {
			ctx.Logger().Error("fail to swap", "msg", pick.msg.Tx.String(), "error", result.Log)
			refundMsg, err := getErrMessageFromABCILog(result.Log)
			if err != nil {
				ctx.Logger().Error("fail to get refund msg", "err", err.Error())
			}
			if newErr := refundTx(ctx, ObservedTx{Tx: pick.msg.Tx}, txOutStore, vm.k, constAccessor, result.Code, refundMsg, eventMgr); nil != newErr {
				ctx.Logger().Error("fail to refund swap", "error", err)
			}
		}
		vm.k.RemoveSwapQueueItem(ctx, pick.msg.Tx.ID)
	}

	return nil
}

// getTodoNum - determine how many swaps to do.
func (vm *SwapQv1) getTodoNum(queueLen int) int {
	// Do half the length of the queue. Unless...
	//	1. The queue length is greater than 200
	//  2. The queue legnth is less than 10
	maxSwaps := 100 // TODO: make this a constant
	minSwaps := 10  // TODO: make this a constant
	todo := queueLen / 2
	if minSwaps >= queueLen {
		todo = queueLen
	}
	if maxSwaps < todo {
		todo = maxSwaps
	}
	return todo
}

// ScoreMsgs - this takes a list of MsgSwap, and converts them to a scored
// swapItem list
func (vm *SwapQv1) ScoreMsgs(ctx sdk.Context, msgs []MsgSwap) (swapItems, error) {
	pools := make(map[common.Asset]Pool, 0)
	items := make(swapItems, 0)

	for _, msg := range msgs {
		if _, ok := pools[msg.TargetAsset]; !ok {
			var err error
			pools[msg.TargetAsset], err = vm.k.GetPool(ctx, msg.TargetAsset)
			if err != nil {
				return items, err
			}
		}

		item := swapItem{
			msg:  msg,
			fee:  sdk.ZeroUint(),
			slip: sdk.ZeroUint(),
		}

		pool := pools[msg.TargetAsset]
		if pool.Empty() || !pool.IsEnabled() || pool.BalanceRune.IsZero() || pool.BalanceAsset.IsZero() {
			items = append(items, item)
			continue
		}

		sourceCoin := msg.Tx.Coins[0]

		// Get our X, x, Y values
		var X, x, Y sdk.Uint
		x = sourceCoin.Amount
		if sourceCoin.Asset.IsRune() {
			X = pool.BalanceRune
			Y = pool.BalanceAsset
		} else {
			Y = pool.BalanceRune
			X = pool.BalanceAsset
		}

		item.fee = calcLiquidityFee(X, x, Y)
		if sourceCoin.Asset.IsRune() {
			item.fee = pool.AssetValueInRune(item.fee)
		}
		item.slip = calcTradeSlip(X, x)

		items = append(items, item)
	}

	return items, nil
}

func (items swapItems) Sort() swapItems {
	// sort by liquidity fee
	byFee := items
	sort.SliceStable(byFee, func(i, j int) bool {
		return byFee[i].fee.GT(byFee[j].fee)
	})

	// sort by slip fee
	bySlip := items
	sort.SliceStable(bySlip, func(i, j int) bool {
		return bySlip[i].fee.GT(bySlip[j].fee)
	})

	type score struct {
		msg   MsgSwap
		score int
	}

	// add liquidity fee score
	scores := make([]score, len(items))
	for i, item := range byFee {
		scores[i] = score{
			msg:   item.msg,
			score: i,
		}
	}

	// add slip score
	for i, item := range bySlip {
		for j, score := range scores {
			if score.msg.Tx.ID.Equals(item.msg.Tx.ID) {
				scores[j].score += i
				break
			}
		}
	}

	// sort by score
	sort.SliceStable(scores, func(i, j int) bool {
		return scores[i].score < scores[j].score
	})

	// sort our items by score
	sorted := make(swapItems, len(items))
	for i, score := range scores {
		for _, item := range items {
			if item.msg.Tx.ID.Equals(score.msg.Tx.ID) {
				sorted[i] = item
				break
			}
		}
	}

	return sorted
}
