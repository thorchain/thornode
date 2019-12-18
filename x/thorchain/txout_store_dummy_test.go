package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// TxOutStoreDummy is going to manage all the outgoing tx
type TxOutStoreDummy struct {
	blockOut *TxOut
	asgard   common.PoolPubKeys
}

// NewTxOutStoreDummy will create a new instance of TxOutStore.
func NewTxStoreDummy() *TxOutStoreDummy {
	return &TxOutStoreDummy{
		blockOut: NewTxOut(100),
		asgard:   GetRandomPoolPubKeys(),
	}
}

// NewBlock create a new block
func (tos *TxOutStoreDummy) NewBlock(height uint64, constAccessor constants.ConstantValues) {
	tos.blockOut = NewTxOut(height)
}

func (tos *TxOutStoreDummy) GetAsgardPoolPubKey(chain common.Chain) *common.PoolPubKey {
	return tos.asgard.GetByChain(chain)
}

// CommitBlock THORNode write the block into key value store , thus THORNode could send to signer later.
func (tos *TxOutStoreDummy) CommitBlock(ctx sdk.Context) {}

func (tos *TxOutStoreDummy) GetBlockOut() *TxOut {
	return tos.blockOut
}

func (tos *TxOutStoreDummy) GetOutboundItems() []*TxOutItem {
	return tos.blockOut.TxArray
}

// AddTxOutItem add an item to internal structure
func (tos *TxOutStoreDummy) AddTxOutItem(ctx sdk.Context, toi *TxOutItem) {
	tos.addToBlockOut(toi)
}

func (tos *TxOutStoreDummy) addToBlockOut(toi *TxOutItem) {
	tos.blockOut.TxArray = append(tos.blockOut.TxArray, toi)
}

func (tos *TxOutStoreDummy) getSeqNo(chain common.Chain) uint64 {
	return uint64(0)
}

func (tos *TxOutStoreDummy) CollectYggdrasilPools(ctx sdk.Context, tx ObservedTx) (Yggdrasils, error) {
	return nil, nil
}
