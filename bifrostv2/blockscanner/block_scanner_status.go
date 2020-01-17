package blockscanner

// BlockScannerStatus
type BlockScannerStatus byte

const (
	Processing BlockScannerStatus = iota
	Failed
	Finished
	NotStarted
)
