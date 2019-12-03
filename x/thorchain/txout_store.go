package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// TxOutSetter define a method that is required to be used in TxOutStore
// We need this interface thus THORNode could test the refund logic accordingly
type TxOutSetter interface {
	SetTxOut(sdk.Context, *TxOut) error
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

// CommitBlock THORNode write the block into key value store , thus THORNode could send to signer later.
func (tos *TxOutStore) CommitBlock(ctx sdk.Context) {
	// if THORNode don't have anything in the array, THORNode don't need to save
	if len(tos.blockOut.TxArray) == 0 {
		return
	}

	// write the tos to keeper
	if err := tos.txOutSetter.SetTxOut(ctx, tos.blockOut); nil != err {
		ctx.Logger().Error("fail to save tx out", err)
	}
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

	// If THORNode don't have a pool already selected to send from, discover one.
	if toi.PoolAddress.IsEmpty() {
		if !asgard {
			// When deciding which Yggdrasil pool will send out our tx out, we
			// should consider which ones observed the inbound request tx, as
			// yggdrasil pools can go offline. Here THORNode get the voter record and
			// only consider Yggdrasils where their observed saw the "correct"
			// tx.

			activeNodeAccounts, err := keeper.ListActiveNodeAccounts(ctx)
			if len(activeNodeAccounts) > 0 && err == nil {
				voter := keeper.GetTxInVoter(ctx, toi.InHash)
				tx := voter.GetTx(activeNodeAccounts)

				// collect yggdrasil pools
				yggs := tos.CollectYggdrasilPools(ctx, keeper, tx)
				yggs = yggs.SortBy(toi.Coin.Asset)

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

	// Ensure THORNode are not sending from and to the same address
	// THORNode check for a
	fromAddr, err := toi.PoolAddress.GetAddress(toi.Chain)
	if err != nil || fromAddr.IsEmpty() || toi.ToAddress.Equals(fromAddr) {
		return
	}

	// Deduct TransactionFee from TOI and add to Reserve
	nodes, err := keeper.TotalActiveNodeAccount(ctx)

	if nodes >= (constants.MinmumNodesForBFT) && err == nil {
		var runeFee sdk.Uint
		if toi.Coin.Asset.IsRune() {
			if toi.Coin.Amount.LTE(sdk.NewUint(constants.TransactionFee)) {
				runeFee = toi.Coin.Amount // Fee is the full amount
			} else {
				runeFee = sdk.NewUint(constants.TransactionFee) // Fee is the prescribed fee
			}
			toi.Coin.Amount = common.SafeSub(toi.Coin.Amount, runeFee)
			if err := keeper.AddFeeToReserve(ctx, runeFee); nil != err {
				// Add to reserve
				ctx.Logger().Error("fail to add fee to reserve", err)
			}
		} else {
			pool := keeper.GetPool(ctx, toi.Coin.Asset)                              // Get pool
			assetFee := pool.AssetValueInRune(sdk.NewUint(constants.TransactionFee)) // Get fee in Asset value
			if toi.Coin.Amount.LTE(assetFee) {
				assetFee = toi.Coin.Amount // Fee is the full amount
				runeFee = pool.RuneValueInAsset(assetFee)
			} else {
				runeFee = sdk.NewUint(constants.TransactionFee) // Fee is the prescribed fee
			}
			toi.Coin.Amount = common.SafeSub(toi.Coin.Amount, assetFee)  // Deduct Asset fee
			pool.BalanceAsset = pool.BalanceAsset.Add(assetFee)          // Add Asset fee to Pool
			pool.BalanceRune = common.SafeSub(pool.BalanceRune, runeFee) // Deduct Rune from Pool
			keeper.SetPool(ctx, pool)                                    // Set Pool
			if err := keeper.AddFeeToReserve(ctx, runeFee); nil != err {
				ctx.Logger().Error("fail to add fee to reserve", err)
				// Add to reserve
			}
		}
	}

	if toi.Coin.IsEmpty() {
		return
	}

	// increment out number of out tx for this in tx
	voter := keeper.GetTxInVoter(ctx, toi.InHash)
	voter.Actions = append(voter.Actions, *toi)
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

func AddGasFees(ctx sdk.Context, keeper Keeper, gas common.Gas) error {
	vault, err := keeper.GetVaultData(ctx)
	if nil != err {
		return fmt.Errorf("fail to get vault: %w", err)
	}
	vault.Gas = vault.Gas.Add(gas)
	return keeper.SetVaultData(ctx, vault)
}

func (tos *TxOutStore) CollectYggdrasilPools(ctx sdk.Context, keeper Keeper, tx TxIn) Yggdrasils {
	// collect yggdrasil pools
	var yggs Yggdrasils
	iterator := keeper.GetYggdrasilIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var ygg Yggdrasil
		keeper.Cdc().MustUnmarshalBinaryBare(iterator.Value(), &ygg)
		// if THORNode are already sending assets from this ygg pool, deduct
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
				ygg.Coins[i].Amount = common.SafeSub(ygg.Coins[i].Amount, tx.Coin.Amount)
			}
		}
		yggs = append(yggs, ygg)
	}

	return yggs
}
