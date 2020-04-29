package thorchain

import (
	"math/rand"
	"sort"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// SwapQv1 is going to manage the vaults
type SwapQv1 struct {
	k                   Keeper
	versionedTxOutStore VersionedTxOutStore
}

type swapItem struct {
	msg  MsgSwap
	fee  sdk.Uint
	slip sdk.Uint
}
type swapItems []swapItem

// NewSwapQv1 create a new vault manager
func NewSwapQv1(k Keeper, versionedTxOutStore VersionedTxOutStore) *SwapQv1 {
	return &SwapQv1{
		k:                   k,
		versionedTxOutStore: versionedTxOutStore,
	}
}

// EndBlock move funds from retiring asgard vaults
func (vm *SwapQv1) EndBlock(ctx sdk.Context, version semver.Version, constAccessor constants.ConstantValues) error {
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

	limit := 100 // TODO: make this a constant
	r := rand.New(rand.NewSource(swaps.randSeed()))
	for i := 0; i < limit; i++ {
		pick := swaps.Pick(r)
		msg := swaps[pick].msg // grab our msg

		// remove pick from swaps
		swaps = append(swaps[:pick], swaps[i+pick:]...)

		// TODO: process msg
		_ = msg
	}

	return nil
}

// ScoreMsgs - this takes a list of MsgSwap, and converts them to a scored
// swapItem list
func (vm *SwapQv1) ScoreMsgs(ctx sdk.Context, msgs []MsgSwap) (swapItems, error) {
	pools := make(map[common.Asset]Pool, 0)
	items := make(swapItems, len(msgs))

	for i, msg := range msgs {
		if _, ok := pools[msg.TargetAsset]; !ok {
			var err error
			pools[msg.TargetAsset], err = vm.k.GetPool(ctx, msg.TargetAsset)
			if err != nil {
				return items, err
			}
		}

		pool := pools[msg.TargetAsset]
		sourceCoin := msg.Tx.Coins[0]

		// Get our X, x, Y values
		var X, x, Y, liquidityFee, slip sdk.Uint
		x = sourceCoin.Amount
		if sourceCoin.Asset.IsRune() {
			X = pool.BalanceRune
			Y = pool.BalanceAsset
		} else {
			Y = pool.BalanceRune
			X = pool.BalanceAsset
		}

		liquidityFee = calcLiquidityFee(X, x, Y)
		if sourceCoin.Asset.IsRune() {
			liquidityFee = pool.AssetValueInRune(liquidityFee)
		}
		slip = calcTradeSlip(X, x)

		items[i] = swapItem{
			msg:  msg,
			fee:  liquidityFee,
			slip: slip,
		}
	}

	return items, nil
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

// randSeed - builds a rand generator based on the tx hash of all swaps in the
// queue
func (items swapItems) randSeed() int64 {
	var ans int64
	for _, item := range items {
		txID := []byte(item.msg.Tx.ID)
		for _, b := range txID {
			ans += int64(b)
		}
	}
	return ans
}

// Pick - picks a random transaction based on weight of liquidity fee
// Much of this code borrowed from https://github.com/mroth/weightedrand
func (items swapItems) Pick(r *rand.Rand) int {
	sort.Slice(items, func(i, j int) bool {
		return items[i].fee.LT(items[j].fee)
	})
	totals := make([]int, len(items))
	var runningTotal int
	for i, item := range items {
		runningTotal = runningTotal + int(item.fee.Uint64())
		totals[i] = runningTotal
	}

	r2 := r.Intn(runningTotal + 1)
	return sort.SearchInts(totals, r2)
}
