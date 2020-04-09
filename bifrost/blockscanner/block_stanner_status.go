package blockscanner

// BlockScanStatus
type BlockScanStatus byte

const (
	Processing BlockScanStatus = iota
	Finished
	NotStarted
)
