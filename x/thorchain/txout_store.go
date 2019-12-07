package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// TxOutSetter define a method that is required to be used in TxOutStore
// We need this interface thus THORNode could test the refund logic accordingly
type TxOutSetter interface {
	SetTxOut(sdk.Context, *TxOut) error
}

type TxOutStore interface {
	NewBlock(height uint64)
	CommitBlock(ctx sdk.Context)
	GetOutboundItems() []*TxOutItem
	AddTxOutItem(ctx sdk.Context, keeper Keeper, toi *TxOutItem, asgard bool)
	addToBlockOut(toi *TxOutItem)
	getBlockOut() *TxOut
	getSeqNo(chain common.Chain) uint64
	CollectYggdrasilPools(ctx sdk.Context, keeper Keeper, tx ObservedTx) Yggdrasils
}

// TxOutStorage is going to manage all the outgoing tx
type TxOutStorage struct {
	txOutSetter TxOutSetter
	blockOut    *TxOut
	poolAddrMgr PoolAddressManager
}

// NewTxOutStorage will create a new instance of TxOutStore.
func NewTxOutStorage(txOutSetter TxOutSetter, poolAddrMgr PoolAddressManager) *TxOutStorage {
	return &TxOutStorage{
		txOutSetter: txOutSetter,
		poolAddrMgr: poolAddrMgr,
	}
}

// NewBlock create a new block
func (tos *TxOutStorage) NewBlock(height uint64) {
	tos.blockOut = NewTxOut(height)
}

// CommitBlock THORNode write the block into key value store , thus THORNode could send to signer later.
func (tos *TxOutStorage) CommitBlock(ctx sdk.Context) {
	// if THORNode don't have anything in the array, THORNode don't need to save
	if len(tos.blockOut.TxArray) == 0 {
		return
	}

	// write the tos to keeper
	if err := tos.txOutSetter.SetTxOut(ctx, tos.blockOut); nil != err {
		ctx.Logger().Error("fail to save tx out", err)
	}
}

func (tos *TxOutStorage) getBlockOut() *TxOut {
	return tos.blockOut
}

func (tos *TxOutStorage) GetOutboundItems() []*TxOutItem {
	return tos.blockOut.TxArray
}

// AddTxOutItem add an item to internal structure
func (tos *TxOutStorage) AddTxOutItem(ctx sdk.Context, keeper Keeper, toi *TxOutItem, asgard bool) {
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
				voter, err := keeper.GetObservedTxVoter(ctx, toi.InHash)
				if err != nil {
					ctx.Logger().Error("fail to get observed tx voter", err)
				}
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
			pool, err := keeper.GetPool(ctx, toi.Coin.Asset) // Get pool
			if err != nil {
				// the error is already logged within kvstore
				return
			}

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
			if err := keeper.SetPool(ctx, pool); err != nil {            // Set Pool
				ctx.Logger().Error("fail to save pool", err)
			}
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
	voter, err := keeper.GetObservedTxVoter(ctx, toi.InHash)
	if err != nil {
		ctx.Logger().Error("fail to get observed tx voter", err)
	}
	voter.Actions = append(voter.Actions, *toi)
	keeper.SetObservedTxVoter(ctx, voter)

	// add tx to block out
	tos.addToBlockOut(toi)
}

func (tos *TxOutStorage) addToBlockOut(toi *TxOutItem) {
	toi.SeqNo = tos.getSeqNo(toi.Chain)
	tos.blockOut.TxArray = append(tos.blockOut.TxArray, toi)
}

func (tos *TxOutStorage) getSeqNo(chain common.Chain) uint64 {
	// need to get the sequence no
	currentChainPoolAddr := tos.poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(chain)
	if nil != currentChainPoolAddr {
		return currentChainPoolAddr.GetSeqNo()
	}
	if nil != tos.poolAddrMgr.GetCurrentPoolAddresses().Previous {
		previousChainPoolAddr := tos.poolAddrMgr.GetCurrentPoolAddresses().Previous.GetByChain(chain)
		if nil != previousChainPoolAddr {
			return previousChainPoolAddr.GetSeqNo()
		}
	}
	if nil != tos.poolAddrMgr.GetCurrentPoolAddresses().Next {
		nextChainPoolAddr := tos.poolAddrMgr.GetCurrentPoolAddresses().Next.GetByChain(chain)
		if nil != nextChainPoolAddr {
			return nextChainPoolAddr.GetSeqNo()
		}
	}
	return uint64(0)
}

func (tos *TxOutStorage) CollectYggdrasilPools(ctx sdk.Context, keeper Keeper, tx ObservedTx) Yggdrasils {
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
