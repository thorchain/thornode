package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

// TODO: make this admin configs instead of hard coded
var singleTransactionFee uint64 = 37500
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
	poolAddrMgr *PoolAddressManager
}

// NewTxOutStore will create a new instance of TxOutStore.
func NewTxOutStore(txOutSetter TxOutSetter, poolAddrMgr *PoolAddressManager) *TxOutStore {
	return &TxOutStore{
		txOutSetter: txOutSetter,
		poolAddrMgr: poolAddrMgr,
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

	if len(toi.Coins) > 0 {
		switch toi.Coins[0].Asset.Chain {
		case common.BNBChain:
			tos.ApplyBNBFees(ctx, keeper, toi)
		default:
			// No gas policy for this chain (yet)
		}

		tos.addToBlockOut(toi)
	}
}

func (tos *TxOutStore) ApplyBNBFees(ctx sdk.Context, keeper Keeper, toi *TxOutItem) {
	// detect if one of our coins is bnb or rune. We use this to help determine
	// which coin we should deduct fees from. The priority, in order, is BNB,
	// Rune, other.
	hasBNB := false
	hasRune := false
	for _, item := range toi.Coins {
		if common.IsBNBAsset(item.Asset) {
			hasBNB = true
		}
		if common.IsRuneAsset(item.Asset) {
			hasRune = true
		}
	}

	// TODO: if we don't have enough coin amount to pay for gas, we just take
	// it all and don't take the rest from another coin

	hasDeductedGas := false // monitor if we've already pulled out coins for gas.
	var gas uint64
	if len(toi.Coins) == 1 {
		gas = singleTransactionFee
	} else {
		gas = batchTransactionFee * uint64(len(toi.Coins))
	}
	for i, item := range toi.Coins {
		if !hasDeductedGas && common.IsBNBAsset(item.Asset) {
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

		if !hasDeductedGas && hasBNB == false && common.IsRuneAsset(item.Asset) {
			bnbPool := keeper.GetPool(ctx, common.BNBAsset)

			if bnbPool.BalanceRune.IsZero() {
				toi.Coins[i] = item
				hasDeductedGas = true
				continue
			}

			var runeAmt uint64
			runeAmt = uint64(float64(bnbPool.BalanceRune.Uint64()) / (float64(bnbPool.BalanceAsset.Uint64()) / float64(gas)))

			if item.Amount.LT(sdk.NewUint(gas)) {
				item.Amount = sdk.ZeroUint()
			} else {
				item.Amount = item.Amount.SubUint64(runeAmt)
			}

			// add the rune to the bnb pool that we are subtracting from
			// the refund
			bnbPool.BalanceRune = bnbPool.BalanceRune.AddUint64(runeAmt)
			bnbPool.BalanceAsset = bnbPool.BalanceAsset.SubUint64(gas)
			keeper.SetPool(ctx, bnbPool)

			toi.Coins[i] = item
			hasDeductedGas = true
			continue
		}

		if !hasDeductedGas && hasBNB == false && hasRune == false {
			bnbPool := keeper.GetPool(ctx, common.BNBAsset)
			assetPool := keeper.GetPool(ctx, item.Asset)

			if bnbPool.BalanceRune.IsZero() || assetPool.BalanceRune.IsZero() {
				toi.Coins[i] = item
				hasDeductedGas = true
				continue
			}

			var runeAmt, assetAmt uint64
			runeAmt = uint64(float64(bnbPool.BalanceRune.Uint64()) / (float64(bnbPool.BalanceAsset.Uint64()) / float64(gas)))
			assetAmt = uint64(float64(assetPool.BalanceRune.Uint64()) / (float64(assetPool.BalanceAsset.Uint64()) / float64(runeAmt)))

			if item.Amount.LT(sdk.NewUint(assetAmt)) {
				item.Amount = sdk.ZeroUint()
			} else {
				item.Amount = item.Amount.SubUint64(assetAmt)
			}

			// add the rune to the bnb pool that we are subtracting from
			// the refund
			bnbPool.BalanceRune = bnbPool.BalanceRune.AddUint64(runeAmt)
			bnbPool.BalanceAsset = bnbPool.BalanceAsset.SubUint64(gas)
			keeper.SetPool(ctx, bnbPool)
			assetPool.BalanceRune = assetPool.BalanceRune.SubUint64(runeAmt)
			assetPool.BalanceAsset = assetPool.BalanceAsset.AddUint64(assetAmt)
			keeper.SetPool(ctx, assetPool)

			toi.Coins[i] = item
			hasDeductedGas = true
			continue

		}
	}
}

func (tos *TxOutStore) addToBlockOut(toi *TxOutItem) {
	// count the total coins we are sending to the user.
	countCoins := sdk.ZeroUint()
	for _, item := range toi.Coins {
		countCoins = countCoins.Add(item.Amount)
	}

	// if we are sending zero coins, don't bother adding to the txarray
	if !countCoins.IsZero() {
		toi.SeqNo = tos.getSeqNo(toi.Chain)
		tos.blockOut.TxArray = append(tos.blockOut.TxArray, toi)
	}
}
func (tos *TxOutStore) getSeqNo(chain common.Chain) uint64 {
	// need to get the sequence no
	currentChainPoolAddr := tos.poolAddrMgr.currentPoolAddresses.Current.GetByChain(chain)
	if nil != currentChainPoolAddr {
		return currentChainPoolAddr.GetSeqNo()
	}
	if nil != tos.poolAddrMgr.currentPoolAddresses.Previous {
		previousChainPoolAddr := tos.poolAddrMgr.currentPoolAddresses.Previous.GetByChain(chain)
		if nil != previousChainPoolAddr {
			return previousChainPoolAddr.GetSeqNo()
		}
	}
	if nil != tos.poolAddrMgr.currentPoolAddresses.Next {
		nextChainPoolAddr := tos.poolAddrMgr.currentPoolAddresses.Next.GetByChain(chain)
		if nil != nextChainPoolAddr {
			return nextChainPoolAddr.GetSeqNo()
		}
	}
	return uint64(0)
}
