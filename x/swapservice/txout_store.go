package swapservice

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
)

// TODO: make this admin configs instead of hard coded
// var singleTransactionFee uint64 = 37500
var batchTransactionFee uint64 = 30000

// TxOutSetter define a method that is required to be used in TxOutStore
// We need this interface thus we could test the refund logic accordingly
type TxOutSetter interface {
	SetTxOut(sdk.Context, *TxOut)
}

// TxOutStore is going to manage all the outgoing tx
type TxOutStore struct {
	txOutSetter TxOutSetter
	blockOut    *TxOut
}

// NewTxOutStore will create a new instance of TxOutStore.
func NewTxOutStore(txOutSetter TxOutSetter) *TxOutStore {
	return &TxOutStore{
		txOutSetter: txOutSetter,
	}
}

// NewBlock create a new block
func (tos *TxOutStore) NewBlock(height uint64) {
	tos.blockOut = NewTxOut(height)
}

// CommitBlock we write the block into key value store , thus we could send to signer later.
func (tos *TxOutStore) CommitBlock(ctx sdk.Context) {
	// if we don't have anything in the array, we don't need to save
	if len(tos.blockOut.TxArray) == 0 {
		return
	}
	// write the tos to keeper
	tos.txOutSetter.SetTxOut(ctx, tos.blockOut)
}

func (tos *TxOutStore) GetOutboundItems() []*TxOutItem {
	return tos.blockOut.TxArray
}

// AddTxOutItem add an item to internal structure
func (tos *TxOutStore) AddTxOutItem(ctx sdk.Context, keeper Keeper, toi *TxOutItem, deductFee bool) {

	if !deductFee {
		tos.addToBlockOut(toi)
		return
	}

	// detect if one of our coins is bnb or rune. We use this to help determine
	// which coin we should deduct fees from. The priority, in order, is BNB,
	// Rune, other.
	hasBNB := false
	hasRune := false
	for _, item := range toi.Coins {
		if common.IsBNB(item.Denom) {
			hasBNB = true
		}
		if common.IsRune(item.Denom) {
			hasRune = true
		}
	}

	// TODO: if we don't have enough coin amount to pay for gas, we just take
	// it all and don't take the rest from another coin

	hasDeductedGas := false // monitor if we've already pulled out coins for gas.
	gas := batchTransactionFee * uint64(len(toi.Coins))
	for i, item := range toi.Coins {
		if !hasDeductedGas && common.IsBNB(item.Denom) {
			if item.Amount.LT(sdk.NewUint(gas)) {
				item.Amount = sdk.ZeroUint()
			} else {
				item.Amount = item.Amount.SubUint64(gas)
			}

			// no need to update the bnb pool with new amounts.

			toi.Coins[i] = item
			hasDeductedGas = true
			continue
		}

		if !hasDeductedGas && hasBNB == false && common.IsRune(item.Denom) {
			bnbPool := keeper.GetPool(ctx, common.BNBTicker)

			if bnbPool.BalanceRune.IsZero() {
				toi.Coins[i] = item
				hasDeductedGas = true
				continue
			}

			var runeAmt uint64
			runeAmt = uint64((float64(gas) / float64(bnbPool.BalanceToken.Uint64())) * float64(bnbPool.BalanceRune.Uint64()))

			if item.Amount.LT(sdk.NewUint(gas)) {
				item.Amount = sdk.ZeroUint()
			} else {
				item.Amount = item.Amount.SubUint64(runeAmt)
			}

			// add the rune to the bnb pool that we are subtracting from
			// the refund
			bnbPool.BalanceRune = bnbPool.BalanceRune.AddUint64(runeAmt)
			bnbPool.BalanceToken = bnbPool.BalanceToken.SubUint64(gas)
			keeper.SetPool(ctx, bnbPool)

			toi.Coins[i] = item
			hasDeductedGas = true
			continue
		}

		if !hasDeductedGas && hasBNB == false && hasRune == false {
			bnbPool := keeper.GetPool(ctx, common.BNBTicker)
			tokenPool := keeper.GetPool(ctx, item.Denom)

			if bnbPool.BalanceRune.IsZero() || tokenPool.BalanceRune.IsZero() {
				toi.Coins[i] = item
				hasDeductedGas = true
				continue
			}

			var runeAmt, tokenAmt uint64
			runeAmt = uint64((float64(gas) / float64(bnbPool.BalanceToken.Uint64())) * float64(bnbPool.BalanceRune.Uint64()))
			tokenAmt = uint64((float64(runeAmt) / float64(tokenPool.BalanceRune.Uint64())) * float64(tokenPool.BalanceToken.Uint64()))

			if item.Amount.LT(sdk.NewUint(tokenAmt)) {
				item.Amount = sdk.ZeroUint()
			} else {
				item.Amount = item.Amount.SubUint64(tokenAmt)
			}

			// add the rune to the bnb pool that we are subtracting from
			// the refund
			bnbPool.BalanceRune = bnbPool.BalanceRune.AddUint64(runeAmt)
			bnbPool.BalanceToken = bnbPool.BalanceToken.SubUint64(gas)
			keeper.SetPool(ctx, bnbPool)
			tokenPool.BalanceRune = tokenPool.BalanceRune.SubUint64(runeAmt)
			tokenPool.BalanceToken = tokenPool.BalanceToken.AddUint64(tokenAmt)
			keeper.SetPool(ctx, tokenPool)

			toi.Coins[i] = item
			hasDeductedGas = true
			continue

		}
	}
	tos.addToBlockOut(toi)

}
func (tos *TxOutStore) addToBlockOut(toi *TxOutItem) {
	// count the total coins we are sending to the user.
	countCoins := sdk.ZeroUint()
	for _, item := range toi.Coins {
		countCoins = countCoins.Add(item.Amount)
	}

	// if we are sending zero coins, don't bother adding to the txarray
	if !countCoins.IsZero() {
		tos.blockOut.TxArray = append(tos.blockOut.TxArray, toi)
	}
}
