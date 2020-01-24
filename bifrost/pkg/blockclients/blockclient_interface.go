package blockclients

import (
	"gitlab.com/thorchain/thornode/bifrost/types"
	"gitlab.com/thorchain/thornode/common"
	stypes "gitlab.com/thorchain/thornode/x/thorchain/types"
)

// BlockChainClient is the interface that wraps basic chain client methods
//
// SignTx       signs transactions
// BroadcastTx  broadcast transactions on the chain associated with the client
// Start        starts block chain client scanning for new blocks
// Stop         stops block scanner
type BlockChainClient interface {
	SignTx(tx *stypes.TxOutItem, blockHeight int64) (*stypes.TxOutItem, error)
	EqualsChain(chain common.Chain) bool
	BroadcastTx(tx *stypes.TxOutItem) error
	Start(txInChan chan<- types.Block, startHeight types.FnLastScannedBlockHeight) error
	Stop() error
}
