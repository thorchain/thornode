package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type VersionedTxOutStoreDummy struct {
	txoutStore *TxOutStoreDummy
}

func NewVersionedTxOutStoreDummy() *VersionedTxOutStoreDummy {
	return &VersionedTxOutStoreDummy{
		txoutStore: NewTxStoreDummy(),
	}
}

func (v *VersionedTxOutStoreDummy) GetTxOutStore(ctx sdk.Context, keeper Keeper, version semver.Version) (TxOutStore, error) {
	return v.txoutStore, nil
}

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
func (tos *TxOutStoreDummy) NewBlock(height int64, constAccessor constants.ConstantValues) {
	tos.blockOut = NewTxOut(height)
}

func (tos *TxOutStoreDummy) GetBlockOut(_ sdk.Context) (*TxOut, error) {
	return tos.blockOut, nil
}

func (tos *TxOutStoreDummy) ClearOutboundItems(ctx sdk.Context) {
	tos.blockOut = NewTxOut(tos.blockOut.Height)
}

func (tos *TxOutStoreDummy) GetOutboundItems(ctx sdk.Context) ([]*TxOutItem, error) {
	return tos.blockOut.TxArray, nil
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
	tos.addToBlockOut(ctx, toi)
	return true, nil
}

func (tos *TxOutStoreDummy) UnSafeAddTxOutItem(ctx sdk.Context, toi *TxOutItem) error {
	tos.addToBlockOut(ctx, toi)
	return nil
}

func (tos *TxOutStoreDummy) addToBlockOut(_ sdk.Context, toi *TxOutItem) {
	tos.blockOut.TxArray = append(tos.blockOut.TxArray, toi)
}
