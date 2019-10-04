package swapservice

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

// AddTxOutItem add an item to internal structure
func (tos *TxOutStore) AddTxOutItem(ctx sdk.Context, keeper Keeper, toi *TxOutItem) {

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

	hasDeductedGas := false // monitor if we've already pulled out coins for gas.
	for i, item := range toi.Coins {
		if !hasDeductedGas && common.IsBNB(item.Denom) {
			if len(toi.Coins) == 1 {
				item.Amount = item.Amount.SubUint64(singleTransactionFee)
			} else {
				item.Amount = item.Amount.SubUint64(batchTransactionFee * uint64(len(toi.Coins)))
			}

			// no need to update the bnb pool with new amounts.

			if item.Amount.GT(sdk.ZeroUint()) {
				toi.Coins[i] = item
				hasDeductedGas = true
				continue
			}
		}

		if !hasDeductedGas && hasBNB == false && common.IsRune(item.Denom) {
			bnbPool := keeper.GetPool(ctx, common.BNBTicker)

			var runeAmt uint64
			if len(toi.Coins) == 1 {
				runeAmt = (singleTransactionFee / bnbPool.BalanceToken.Uint64()) * (bnbPool.BalanceRune.Uint64())
			} else {
				runeAmt = (batchTransactionFee / bnbPool.BalanceToken.Uint64()) * (bnbPool.BalanceRune.Uint64()) * uint64(len(toi.Coins))
			}

			item.Amount = item.Amount.SubUint64(runeAmt)
			if item.Amount.GT(sdk.ZeroUint()) {
				// add the rune to the bnb pool that we are subtracting from
				// the refund
				bnbPool.BalanceRune = bnbPool.BalanceRune.AddUint64(runeAmt)
				keeper.SetPool(ctx, bnbPool)

				toi.Coins[i] = item
				hasDeductedGas = true
				continue
			}
		}

		if !hasDeductedGas && hasBNB == false && hasRune == false {
			bnbPool := keeper.GetPool(ctx, common.BNBTicker)
			tokenPool := keeper.GetPool(ctx, item.Denom)

			var runeAmt, tokenAmt uint64
			if len(toi.Coins) == 1 {
				runeAmt = (singleTransactionFee / bnbPool.BalanceToken.Uint64()) * (bnbPool.BalanceRune.Uint64())
				tokenAmt = (runeAmt / tokenPool.BalanceRune.Uint64()) * (tokenPool.BalanceToken.Uint64())
			} else {
				runeAmt = (batchTransactionFee / bnbPool.BalanceToken.Uint64()) * (bnbPool.BalanceRune.Uint64()) * uint64(len(toi.Coins))
				tokenAmt = (runeAmt / tokenPool.BalanceRune.Uint64()) * (tokenPool.BalanceToken.Uint64())
			}

			item.Amount = item.Amount.SubUint64(tokenAmt)
			if item.Amount.GT(sdk.ZeroUint()) {
				// add the rune to the bnb pool that we are subtracting from
				// the refund
				bnbPool.BalanceRune = bnbPool.BalanceRune.AddUint64(runeAmt)
				keeper.SetPool(ctx, bnbPool)
				tokenPool.BalanceRune = bnbPool.BalanceRune.SubUint64(runeAmt)
				tokenPool.BalanceToken = bnbPool.BalanceToken.AddUint64(tokenAmt)
				keeper.SetPool(ctx, tokenPool)

				toi.Coins[i] = item
				hasDeductedGas = true
				continue
			}

		}
	}

	tos.blockOut.TxArray = append(tos.blockOut.TxArray, toi)
}
