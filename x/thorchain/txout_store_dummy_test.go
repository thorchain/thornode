package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// TxOutStoreDummy is going to manage all the outgoing tx
type TxOutStoreDummy struct {
	blockOut *TxOut
	asgard   common.PubKey
}

// NewTxOutStoreDummy will create a new instance of TxOutStore.
func NewTxStoreDummy() *TxOutStoreDummy {
	return &TxOutStoreDummy{
		blockOut: NewTxOut(100),
		asgard:   GetRandomPubKey(),
	}
}

// NewBlock create a new block
func (tos *TxOutStoreDummy) NewBlock(height uint64, constAccessor constants.ConstantValues) {
	tos.blockOut = NewTxOut(height)
}

// CommitBlock THORNode write the block into key value store , thus THORNode could send to signer later.
func (tos *TxOutStoreDummy) CommitBlock(ctx sdk.Context) {}

func (tos *TxOutStoreDummy) GetBlockOut() *TxOut {
	return tos.blockOut
}

func (tos *TxOutStoreDummy) ClearOutboundItems() {
	tos.blockOut = NewTxOut(tos.blockOut.Height)
}

func (tos *TxOutStoreDummy) GetOutboundItems() []*TxOutItem {
	return tos.blockOut.TxArray
}

func (tos *TxOutStoreDummy) GetOutboundItemByToAddress(to common.Address) []TxOutItem {
	items := make([]TxOutItem, 0)
	for _, item := range tos.blockOut.TxArray {
		if item.ToAddress.Equals(to) {
			items = append(items, *item)
		}
	}
	return items
}

// AddTxOutItem add an item to internal structure
func (tos *TxOutStoreDummy) TryAddTxOutItem(ctx sdk.Context, toi *TxOutItem) (bool, error) {
	tos.addToBlockOut(toi)
	return true, nil
}

func (tos *TxOutStoreDummy) addToBlockOut(toi *TxOutItem) {
	tos.blockOut.TxArray = append(tos.blockOut.TxArray, toi)
}
