package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
)

// TxOutStoreDummy is going to manage all the outgoing tx
type TxOutStoreDummy struct {
	txOutSetter TxOutSetter
	blockOut    *TxOut
}

// NewTxOutStoreDummy will create a new instance of TxOutStore.
func NewTxStoreDummy(txOutSetter TxOutSetter) *TxOutStoreDummy {
	return &TxOutStoreDummy{
		txOutSetter: txOutSetter,
	}
}

// NewBlock create a new block
func (tos *TxOutStoreDummy) NewBlock(height uint64) {
	tos.blockOut = NewTxOut(height)
}

// CommitBlock THORNode write the block into key value store , thus THORNode could send to signer later.
func (tos *TxOutStoreDummy) CommitBlock(ctx sdk.Context) {}

func (tos *TxOutStoreDummy) getBlockOut() *TxOut {
	return tos.blockOut
}

func (tos *TxOutStoreDummy) GetOutboundItems() []*TxOutItem {
	return tos.blockOut.TxArray
}

// AddTxOutItem add an item to internal structure
func (tos *TxOutStoreDummy) AddTxOutItem(ctx sdk.Context, keeper Keeper, toi *TxOutItem, asgard bool) {
	tos.addToBlockOut(toi)
}

func (tos *TxOutStoreDummy) addToBlockOut(toi *TxOutItem) {
	toi.SeqNo = tos.getSeqNo(toi.Chain)
	tos.blockOut.TxArray = append(tos.blockOut.TxArray, toi)
}

func (tos *TxOutStoreDummy) getSeqNo(chain common.Chain) uint64 {
	return uint64(0)
}

func (tos *TxOutStoreDummy) CollectYggdrasilPools(ctx sdk.Context, keeper Keeper, tx ObservedTx) Yggdrasils {
	return nil
}
