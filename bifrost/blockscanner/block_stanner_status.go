package blockscanner

// BlockScanStatus
type BlockScanStatus byte

const (
	Processing BlockScanStatus = iota
	Failed
	Finished
	NotStarted
)
