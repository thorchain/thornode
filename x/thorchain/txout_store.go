package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type TxOutStore interface {
	NewBlock(height uint64)
	CommitBlock(ctx sdk.Context)
	GetBlockOut() *TxOut
	GetOutboundItems() []*TxOutItem
	GetAsgardPoolPubKey(_ common.Chain) *common.PoolPubKey
	AddTxOutItem(ctx sdk.Context, toi *TxOutItem)
	CollectYggdrasilPools(ctx sdk.Context, tx ObservedTx) Yggdrasils
}

// TxOutStorage is going to manage all the outgoing tx
type TxOutStorage struct {
	blockOut    *TxOut
	poolAddrMgr PoolAddressManager
	keeper      Keeper
}

// NewTxOutStorage will create a new instance of TxOutStore.
func NewTxOutStorage(keeper Keeper, poolAddrMgr PoolAddressManager) *TxOutStorage {
	return &TxOutStorage{
		poolAddrMgr: poolAddrMgr,
		keeper:      keeper,
	}
}

// NewBlock create a new block
func (tos *TxOutStorage) NewBlock(height uint64) {
	tos.blockOut = NewTxOut(height)
}

func (tos *TxOutStorage) GetAsgardPoolPubKey(chain common.Chain) *common.PoolPubKey {
	return tos.poolAddrMgr.GetAsgardPoolPubKey(chain)
}

// CommitBlock THORNode write the block into key value store , thus THORNode could send to signer later.
func (tos *TxOutStorage) CommitBlock(ctx sdk.Context) {
	// if THORNode don't have anything in the array, THORNode don't need to save
	if len(tos.blockOut.TxArray) == 0 {
		return
	}

	// write the tos to keeper
	if err := tos.keeper.SetTxOut(ctx, tos.blockOut); nil != err {
		ctx.Logger().Error("fail to save tx out", err)
	}
}

func (tos *TxOutStorage) GetBlockOut() *TxOut {
	return tos.blockOut
}

func (tos *TxOutStorage) GetOutboundItems() []*TxOutItem {
	return tos.blockOut.TxArray
}

// AddTxOutItem add an item to internal structure
func (tos *TxOutStorage) AddTxOutItem(ctx sdk.Context, toi *TxOutItem) {
	// Default the memo to the standard outbound memo
	if toi.Memo == "" {
		toi.Memo = NewOutboundMemo(toi.InHash).String()
	}

	// If THORNode don't have a pool already selected to send from, discover one.
	if toi.VaultPubKey.IsEmpty() {
		// When deciding which Yggdrasil pool will send out our tx out, we
		// should consider which ones observed the inbound request tx, as
		// yggdrasil pools can go offline. Here THORNode get the voter record and
		// only consider Yggdrasils where their observed saw the "correct"
		// tx.

		activeNodeAccounts, err := tos.keeper.ListActiveNodeAccounts(ctx)
		if len(activeNodeAccounts) > 0 && err == nil {
			voter, err := tos.keeper.GetObservedTxVoter(ctx, toi.InHash)
			if err != nil {
				ctx.Logger().Error("fail to get observed tx voter", err)
			}
			tx := voter.GetTx(activeNodeAccounts)

			// collect yggdrasil pools
			yggs := tos.CollectYggdrasilPools(ctx, tx)
			yggs = yggs.SortBy(toi.Coin.Asset)

			// if none of our Yggdrasil pools have enough funds to fulfil
			// the order, fallback to our Asguard pool
			if len(yggs) > 0 {
				if toi.Coin.Amount.LT(yggs[0].GetCoin(toi.Coin.Asset).Amount) {
					toi.VaultPubKey = yggs[0].PubKey
				}
			}

		}

	}

	// Apparently we couldn't find a yggdrasil vault to send from, so use asgard
	if toi.VaultPubKey.IsEmpty() {
		toi.VaultPubKey = tos.GetAsgardPoolPubKey(toi.Chain).PubKey
	}

	// Ensure THORNode are not sending from and to the same address
	// THORNode check for a
	fromAddr, err := toi.VaultPubKey.GetAddress(toi.Chain)
	if err != nil || fromAddr.IsEmpty() || toi.ToAddress.Equals(fromAddr) {
		return
	}

	// Deduct TransactionFee from TOI and add to Reserve
	nodes, err := tos.keeper.TotalActiveNodeAccount(ctx)

	if nodes >= (constants.MinmumNodesForBFT) && err == nil {
		var runeFee sdk.Uint
		if toi.Coin.Asset.IsRune() {
			if toi.Coin.Amount.LTE(sdk.NewUint(constants.TransactionFee)) {
				runeFee = toi.Coin.Amount // Fee is the full amount
			} else {
				runeFee = sdk.NewUint(constants.TransactionFee) // Fee is the prescribed fee
			}
			toi.Coin.Amount = common.SafeSub(toi.Coin.Amount, runeFee)
			if err := tos.keeper.AddFeeToReserve(ctx, runeFee); nil != err {
				// Add to reserve
				ctx.Logger().Error("fail to add fee to reserve", err)
			}
		} else {
			pool, err := tos.keeper.GetPool(ctx, toi.Coin.Asset) // Get pool
			if err != nil {
				// the error is already logged within kvstore
				return
			}

			assetFee := pool.RuneValueInAsset(sdk.NewUint(constants.TransactionFee)) // Get fee in Asset value
			if toi.Coin.Amount.LTE(assetFee) {
				assetFee = toi.Coin.Amount // Fee is the full amount
				runeFee = pool.RuneValueInAsset(assetFee)
			} else {
				runeFee = sdk.NewUint(constants.TransactionFee) // Fee is the prescribed fee
			}
			toi.Coin.Amount = common.SafeSub(toi.Coin.Amount, assetFee)  // Deduct Asset fee
			pool.BalanceAsset = pool.BalanceAsset.Add(assetFee)          // Add Asset fee to Pool
			pool.BalanceRune = common.SafeSub(pool.BalanceRune, runeFee) // Deduct Rune from Pool
			if err := tos.keeper.SetPool(ctx, pool); err != nil {        // Set Pool
				ctx.Logger().Error("fail to save pool", err)
			}
			if err := tos.keeper.AddFeeToReserve(ctx, runeFee); nil != err {
				ctx.Logger().Error("fail to add fee to reserve", err)
				// Add to reserve
			}
		}
	}

	if toi.Coin.IsEmpty() {
		return
	}

	// increment out number of out tx for this in tx
	voter, err := tos.keeper.GetObservedTxVoter(ctx, toi.InHash)
	if err != nil {
		ctx.Logger().Error("fail to get observed tx voter", err)
	}
	voter.Actions = append(voter.Actions, *toi)
	tos.keeper.SetObservedTxVoter(ctx, voter)

	// add tx to block out
	tos.addToBlockOut(toi)
}

func (tos *TxOutStorage) addToBlockOut(toi *TxOutItem) {
	fmt.Printf("Add TxOut Item: %+v\n", toi)
	toi.SeqNo = tos.getSeqNo(toi.VaultPubKey, toi.Chain)
	tos.blockOut.TxArray = append(tos.blockOut.TxArray, toi)
}

func (tos *TxOutStorage) getSeqNo(pk common.PubKey, chain common.Chain) uint64 {
	fmt.Printf("SeqNum: Pubkey: %s\n", pk.String())

	// need to get the sequence no
	currentChainPoolAddr := tos.poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(chain)
	if nil != currentChainPoolAddr && pk.Equals(currentChainPoolAddr.PubKey) {
		fmt.Printf("Seq Num: Current\n")
		return currentChainPoolAddr.GetSeqNo()
	}

	if nil != tos.poolAddrMgr.GetCurrentPoolAddresses().Previous {
		previousChainPoolAddr := tos.poolAddrMgr.GetCurrentPoolAddresses().Previous.GetByChain(chain)
		if nil != previousChainPoolAddr && pk.Equals(previousChainPoolAddr.PubKey) {
			fmt.Printf("Seq Num: Prev\n")
			return previousChainPoolAddr.GetSeqNo()
		}
	}

	if nil != tos.poolAddrMgr.GetCurrentPoolAddresses().Next {
		nextChainPoolAddr := tos.poolAddrMgr.GetCurrentPoolAddresses().Next.GetByChain(chain)
		if nil != nextChainPoolAddr && pk.Equals(nextChainPoolAddr.PubKey) {
			fmt.Printf("Seq Num: Next\n")
			return nextChainPoolAddr.GetSeqNo()
		}
	}

	fmt.Printf("Seq Num: None: %d\n", 0)
	return uint64(0)
}

func (tos *TxOutStorage) CollectYggdrasilPools(ctx sdk.Context, tx ObservedTx) Yggdrasils {
	// collect yggdrasil pools
	var yggs Yggdrasils
	iterator := tos.keeper.GetYggdrasilIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var ygg Yggdrasil
		tos.keeper.Cdc().MustUnmarshalBinaryBare(iterator.Value(), &ygg)
		// if THORNode are already sending assets from this ygg pool, deduct
		// them.
		addr, _ := ygg.PubKey.GetThorAddress()
		if !tx.HasSigned(addr) {
			continue
		}
		for _, tx := range tos.blockOut.TxArray {
			if !tx.VaultPubKey.Equals(ygg.PubKey) {
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
