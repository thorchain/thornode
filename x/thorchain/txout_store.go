package thorchain

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/bepswap/thornode/common"
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
func (tos *TxOutStore) AddTxOutItem(ctx sdk.Context, keeper Keeper, toi *TxOutItem, asgard bool) {
	// Default the memo to the standard outbound memo
	if toi.Memo == "" {
		toi.Memo = NewOutboundMemo(toi.InHash).String()
	}

	// If we don't have a pool already selected to send from, discover one.
	if toi.PoolAddress.IsEmpty() {
		if !asgard {
			// When deciding which Yggdrasil pool will send out our tx out, we
			// should consider which ones observed the inbound request tx, as
			// yggdrasil pools can go offline. Here we get the voter record and
			// only consider Yggdrasils where their observed saw the "correct"
			// tx.

			activeNodeAccounts, err := keeper.ListActiveNodeAccounts(ctx)
			if len(activeNodeAccounts) > 0 && err == nil {
				voter := keeper.GetTxInVoter(ctx, toi.InHash)
				tx := voter.GetTx(activeNodeAccounts)

				// collect yggdrasil pools
				var yggs Yggdrasils
				iterator := keeper.GetYggdrasilIterator(ctx)
				defer iterator.Close()
				for ; iterator.Valid(); iterator.Next() {
					var ygg Yggdrasil
					keeper.cdc.MustUnmarshalBinaryBare(iterator.Value(), &ygg)
					// if we are already sending assets from this ygg pool, deduct
					// them.
					addr, _ := ygg.PubKey.GetThorAddress()
					if !tx.HasSigned(addr) {
						continue
					}
					for _, tx := range tos.blockOut.TxArray {
						if !tx.PoolAddress.Equals(ygg.PubKey) {
							continue
						}
						for i, yggcoin := range ygg.Coins {
							if !yggcoin.Asset.Equals(tx.Coin.Asset) {
								continue
							}
							ygg.Coins[i].Amount = ygg.Coins[i].Amount.Sub(tx.Coin.Amount)
						}
					}
					yggs = append(yggs, ygg)
				}

				// use the ygg pool with the highest quantity of our coin
				sort.Slice(yggs[:], func(i, j int) bool {
					return yggs[i].GetCoin(toi.Coin.Asset).Amount.GT(
						yggs[j].GetCoin(toi.Coin.Asset).Amount,
					)
				})

				// if none of our Yggdrasil pools have enough funds to fulfil
				// the order, fallback to our Asguard pool
				if len(yggs) > 0 {
					if toi.Coin.Amount.LT(yggs[0].GetCoin(toi.Coin.Asset).Amount) {
						toi.PoolAddress = yggs[0].PubKey
					}
				}

			}
		}

	}

	if toi.PoolAddress.IsEmpty() {
		toi.PoolAddress = tos.poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(toi.Chain).PubKey
	}

	// Ensure we are not sending from and to the same address
	// we check for a
	fromAddr, err := toi.PoolAddress.GetAddress(toi.Chain)
	if err != nil || fromAddr.IsEmpty() || toi.ToAddress.Equals(fromAddr) {
		return
	}

	if toi.Coin.IsEmpty() {
		return
	}

	// increment out number of out tx for this in tx
	voter := keeper.GetTxInVoter(ctx, toi.InHash)
	voter.NumOuts += 1
	keeper.SetTxInVoter(ctx, voter)

	// add tx to block out
	tos.addToBlockOut(toi)
}

func (tos *TxOutStore) addToBlockOut(toi *TxOutItem) {
	toi.SeqNo = tos.getSeqNo(toi.Chain)
	tos.blockOut.TxArray = append(tos.blockOut.TxArray, toi)
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

func ApplyGasFees(ctx sdk.Context, keeper Keeper, tx common.Tx) {

	for _, gasCoin := range tx.Gas {
		if len(tx.Coins) == 0 {
			return
		}
		// find our coin to take gas from. Prefer non-rune coin so its easier
		// to know which pool to take gas from and move it to the coin the gas
		// was paid in.
		txCoin := tx.Coins[0]
		for _, coin := range tx.Coins {
			if !coin.Asset.IsRune() {
				txCoin = coin
			}
		}

		gasPool := keeper.GetPool(ctx, gasCoin.Asset)
		gas := gasCoin.Amount

		if txCoin.Asset.Equals(gasCoin.Asset) {
			gasPool.BalanceAsset = gasPool.BalanceAsset.Sub(gas)
			keeper.SetPool(ctx, gasPool)
			return
		}

		if txCoin.Asset.IsRune() {
			// Try to detect the pool the tx was made from the memo
			memo, err := ParseMemo(tx.Memo)
			if err != nil {
				fmt.Printf("Unable to parse memo for gas deduction: %s\n", tx.Memo)
				return
			}
			asset := memo.GetAsset()
			if asset.IsEmpty() {
				fmt.Printf("Unable to determine which pool this rune came from: %s\n", tx.Memo)
				return
			}

			txPool := keeper.GetPool(ctx, asset)

			// add the rune to the bnb pool that we are subtracting from
			// the refund
			if txPool.BalanceRune.LT(gas) {
				// we don't have enough asset to pay for gas. Set it to zero
				txPool.BalanceRune = sdk.ZeroUint()
				txPool.Status = PoolBootstrap
			} else {
				txPool.BalanceRune = txPool.BalanceRune.Sub(gas)
				gasPool.BalanceRune = gasPool.BalanceRune.Sub(gas)
			}
			keeper.SetPool(ctx, gasPool)
			keeper.SetPool(ctx, txPool)
			return
		}

		txPool := keeper.GetPool(ctx, txCoin.Asset)

		var runeAmt, assetAmt uint64
		runeAmt = uint64(float64(gasPool.BalanceRune.Uint64()) / (float64(gasPool.BalanceAsset.Uint64()) / float64(gas.Uint64())))
		assetAmt = uint64(float64(txPool.BalanceRune.Uint64()) / (float64(txPool.BalanceAsset.Uint64()) / float64(runeAmt)))

		// add the rune to the bnb pool that we are subtracting from
		// the refund
		if gasPool.BalanceAsset.LT(gas) {
			gasPool.BalanceAsset = sdk.ZeroUint()
			gasPool.Status = PoolBootstrap
		} else {
			gasPool.BalanceRune = gasPool.BalanceRune.AddUint64(runeAmt)
			gasPool.BalanceAsset = gasPool.BalanceAsset.Sub(gas)
		}
		keeper.SetPool(ctx, gasPool)
		if txPool.BalanceRune.LT(sdk.NewUint(runeAmt)) {
			txPool.BalanceRune = sdk.ZeroUint()
			txPool.Status = PoolBootstrap
		} else {
			txPool.BalanceRune = txPool.BalanceRune.SubUint64(runeAmt)
			txPool.BalanceAsset = txPool.BalanceAsset.AddUint64(assetAmt)
		}
		keeper.SetPool(ctx, txPool)
	}
}
