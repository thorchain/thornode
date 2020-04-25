package chainclients

import (
	"gitlab.com/thorchain/thornode/bifrost/config"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/common"
)

// ChainClient is the interface that wraps basic chain client methods
//
// SignTx       signs transactions
// BroadcastTx  broadcast transactions on the chain associated with the client
// GetChain     get chain id
// SignTx       sign transaction
// GetHeight    get chain height
// GetAddress   gets address for public key pool in chain
// GetAccount   gets account from thorclient in cain
// GetGasFee    calculates gas fee based on number of simple transfer sents
// GetConfig	gets the chain configuration
// Start
// Stop
type ChainClient interface {
	SignTx(tx stypes.TxOutItem, height int64) ([]byte, error)
	BroadcastTx(_ stypes.TxOutItem, _ []byte) error
	GetHeight() (int64, error)
	GetAddress(poolPubKey common.PubKey) string
	GetAccount(poolPubKey common.PubKey) (common.Account, error)
	GetChain() common.Chain
	Start(globalTxsQueue chan stypes.TxIn, globalErrataQueue chan stypes.ErrataBlock)
	GetConfig() config.ChainConfiguration
	Stop()
}
