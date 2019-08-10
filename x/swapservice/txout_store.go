package swapservice

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TxOutStore is going to manage all the outgoing tx
type TxOutStore struct {
	keeper   Keeper
	blockOut *TxOut
}

// NewTxOutStore will create a new instance of TxOutStore.
func NewTxOutStore(keeper Keeper) *TxOutStore {
	return &TxOutStore{
		keeper: keeper,
	}
}

// NewBlock create a new block
func (tos *TxOutStore) NewBlock(height int64) {
	tos.blockOut = NewTxOut(height)
}

// CommitBlock we write the block into key value store , thus we could send to signer later.
func (tos *TxOutStore) CommitBlock(ctx sdk.Context) {
	// if we don't have anything in the array, we don't need to save
	if len(tos.blockOut.TxArray) == 0 {
		return
	}
	// write the tos to keeper
	tos.keeper.SetTxOut(ctx, tos.blockOut)
}

// AddTxOutItem add an item to internal structure
func (tos *TxOutStore) AddTxOutItem(toi *TxOutItem) {
	tos.blockOut.TxArray = append(tos.blockOut.TxArray, toi)
}
