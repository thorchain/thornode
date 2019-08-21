package blockscanner

import (
	"io"
)

// ScannerStorage define the method need to be used by scanner
type ScannerStorage interface {
	GetScanPos() (int64, error)
	SetScanPos(block int64) error

	SetBlockScanStatus(block int64, status BlockScanStatus) error
	RemoveBlockStatus(block int64) error

	GetBlocksForRetry(failedOnly bool) ([]int64, error)
	io.Closer
}
