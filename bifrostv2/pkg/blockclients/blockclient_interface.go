package blockclients

import "gitlab.com/thorchain/thornode/bifrostv2/types"

// BlockChainClient is the interface that wraps basic chain client methods
//
// SignTx       signs transactions
// BroadcastTx  broadcast transactions on the chain associated with the client
// Start        starts block chain client scanning for new blocks
// Stop         stops block scanner
type BlockChainClient interface {
	SignTx() error
	BroadcastTx() error
	Start(txInChan chan<- types.Block, startHeight types.FnLastScannedBlockHeight) error
	Stop() error
}
