package chainclients

import (
	"io"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
)

// BlockScannerStorage defines all methods for block scanner storage used in observer and client
type BlockScannerStorage interface {
	GetScanPos() (int64, error)
	SetScanPos(block int64) error

	SetBlockScanStatus(block int64, status blockscanner.BlockScanStatus) error
	RemoveBlockStatus(block int64) error

	GetBlocksForRetry(failedOnly bool) ([]int64, error)
	io.Closer

	SetTxInStatus(types.TxIn, types.TxInStatus) error
	RemoveTxIn(types.TxIn) error
	GetTxInForRetry(failedOnly bool) ([]types.TxIn, error)
}